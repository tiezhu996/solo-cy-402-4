package service

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// uniqueViolation 构造一个真实 PG SQLSTATE 23505 错误，模拟编号撞唯一索引。
func uniqueViolation(constraint string) error {
	return &pgconn.PgError{Code: "23505", ConstraintName: constraint, Message: "duplicate key"}
}

// TestRunWithNumberTxRetriesUniqueConflict 编号唯一冲突时整事务回滚并用新事务重试，成功只执行一次提交。
func TestRunWithNumberTxRetriesUniqueConflict(t *testing.T) {
	dsn := os.Getenv("CY_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("skip postgres integration test: CY_TEST_DB_DSN not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	txStore := repository.NewTxStore(db)

	var attempts atomic.Int32
	err = runWithNumberTx(logger, txStore, "Case", func(tx *repository.TxRepos) error {
		n := attempts.Add(1)
		if n == 1 {
			// 首次尝试撞号：事务整体回滚，序列值作废，必须换全新事务重试。
			return uniqueViolation("uni_cases_case_no")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("retry should succeed, got %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

// TestRunWithNumberTxBusinessErrorNotRetried 业务错误（如已归档禁建账单）不重试、原样透传。
func TestRunWithNumberTxBusinessErrorNotRetried(t *testing.T) {
	dsn := os.Getenv("CY_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("skip postgres integration test: CY_TEST_DB_DSN not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	txStore := repository.NewTxStore(db)

	var attempts atomic.Int32
	err = runWithNumberTx(logger, txStore, "Billing", func(tx *repository.TxRepos) error {
		attempts.Add(1)
		return util.NewAppError(constants.CodeBillingCreateArchivedCase, "archived")
	})
	if err == nil {
		t.Fatal("business error must be returned")
	}
	if code := appErrorCode(t, err); code != constants.CodeBillingCreateArchivedCase {
		t.Fatalf("code = %d, want 40904", code)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("business error must not retry, attempts = %d", got)
	}
}

// TestRunWithNumberTxConflictExhausted 冲突多次仍失败时稳定返回 50301 可重试错误。
func TestRunWithNumberTxConflictExhausted(t *testing.T) {
	dsn := os.Getenv("CY_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("skip postgres integration test: CY_TEST_DB_DSN not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	txStore := repository.NewTxStore(db)

	var attempts atomic.Int32
	err = runWithNumberTx(logger, txStore, "Case", func(tx *repository.TxRepos) error {
		attempts.Add(1)
		return uniqueViolation("uni_cases_case_no")
	})
	if err == nil {
		t.Fatal("must fail after max attempts")
	}
	if code := appErrorCode(t, err); code != constants.CodeNumberAllocateFailed {
		t.Fatalf("code = %d, want 50301", code)
	}
	if got := attempts.Load(); got != int32(numberAllocateMaxAttempts) {
		t.Fatalf("attempts = %d, want %d", got, numberAllocateMaxAttempts)
	}
}

// TestRunWithNumberTxRetriesRecheckArchive 首次撞号重试时，案件在期间被归档，
// 新事务重新锁行并复核归档终态，返回 40904 而不会落单——重试不破坏归档费用终态。
func TestRunWithNumberTxRetriesRecheckArchive(t *testing.T) {
	s := setupArchiveSuite(t)
	caseID, _ := s.newClosedCase(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	txStore := repository.NewTxStore(s.db)

	var attempts atomic.Int32
	archiverDone := make(chan struct{})
	err := runWithNumberTx(logger, txStore, "Billing", func(tx *repository.TxRepos) error {
		n := attempts.Add(1)
		if n >= 2 {
			// 重试是全新事务：在加锁前先等首次回滚触发的归档完成，
			// 避免持锁等待造成与归档者的死锁。
			<-archiverDone
		}
		c, err := tx.Cases.FindByIDForUpdate(caseID)
		if err != nil {
			return err
		}
		if n == 1 {
			// 模拟首次写库撞号：归档者已在锁队列中等待，事务回滚后它获得锁并归档。
			go func() {
				defer close(archiverDone)
				_, _ = s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
			}()
			return uniqueViolation("uni_billings_bill_no")
		}
		// 第二次尝试重新加锁后必须复核到已归档终态，拒绝落单。
		if c.Status == constants.CaseStatusArchived {
			return util.NewAppError(constants.CodeBillingCreateArchivedCase, "archived")
		}
		t.Fatal("second attempt must observe archived status and return 40904 before insert")
		return nil
	})
	if err == nil {
		t.Fatal("must reject billing create after archive")
	}
	if code := appErrorCode(t, err); code != constants.CodeBillingCreateArchivedCase {
		t.Fatalf("code = %d, want 40904", code)
	}
	if n := s.pendingCount(t, caseID); n != 0 {
		t.Fatalf("no billing may be inserted, pending = %d", n)
	}
	if got := s.caseStatus(t, caseID); got != constants.CaseStatusArchived {
		t.Fatalf("case status = %s, want archived", got)
	}
}

// TestNumberingMissingSequenceRetryable 序列缺失时取号失败：稳定返回 50301 且不落任何记录，
// 序列恢复（EnsureSequences 自动对齐历史编号）后同一请求可成功（可重试语义），无需清库或改号。
func TestNumberingMissingSequenceRetryable(t *testing.T) {
	s := setupArchiveSuite(t)
	lawyerID := s.newLawyer(t)
	clientID := s.newClient(t)

	// 先在序列存在时建一个已结案案件，供账单路径使用。
	closedID, _ := s.newClosedCase(t)

	// 临时移除两条序列，模拟取号依赖不可用。
	for _, name := range []string{repository.CaseNoSeqName, repository.BillNoSeqName} {
		if err := s.db.Exec("DROP SEQUENCE IF EXISTS " + name).Error; err != nil {
			t.Fatalf("drop sequence %s: %v", name, err)
		}
	}

	casesBefore := s.countCasesByTitlePrefix(t, "MISSING_SEQ_CASE")
	_, err := s.caseSvc.Create(clientID, lawyerID, "MISSING_SEQ_CASE 缺序列建案",
		constants.CaseTypeCivil, "", nil, nil)
	if err == nil || appErrorCode(t, err) != constants.CodeNumberAllocateFailed {
		t.Fatalf("case create without sequence must be 50301, got %v", err)
	}
	if got := s.countCasesByTitlePrefix(t, "MISSING_SEQ_CASE"); got != casesBefore {
		t.Fatalf("no case row may be inserted on 50301, delta=%d", got-casesBefore)
	}

	billsBefore := s.countBillingsByCase(t, closedID)
	_, err = s.billingSvc.Create(closedID, s.newClient(t), constants.BillingTypeOther, 1, "")
	if err == nil || appErrorCode(t, err) != constants.CodeNumberAllocateFailed {
		t.Fatalf("billing create without sequence must be 50301, got %v", err)
	}
	if got := s.countBillingsByCase(t, closedID); got != billsBefore {
		t.Fatalf("no billing row may be inserted on 50301, delta=%d", got-billsBefore)
	}

	// 恢复序列：EnsureSequences 会对齐已有编号，之后重试同一请求应成功且编号唯一。
	if err := repository.NewSequenceRepository(s.db).EnsureSequences(); err != nil {
		t.Fatalf("restore sequences: %v", err)
	}
	c, err := s.caseSvc.Create(clientID, lawyerID, "MISSING_SEQ_CASE 恢复后建案",
		constants.CaseTypeCivil, "", nil, nil)
	if err != nil {
		t.Fatalf("retry after sequence restore must succeed: %v", err)
	}
	if c.CaseNo == "" {
		t.Fatal("case_no must be allocated after restore")
	}
	b, err := s.billingSvc.Create(closedID, clientID, constants.BillingTypeOther, 1, "")
	if err != nil {
		t.Fatalf("billing retry after restore must succeed: %v", err)
	}
	if b.BillNo == "" {
		t.Fatal("bill_no must be allocated after restore")
	}
}

// TestNumberingMultiInstanceUnique 模拟多个实例（各自独立的 service/仓储，共连同一数据库）
// 高并发连续建案/建账：编号全局唯一、成功数严格等于落库数。
func TestNumberingMultiInstanceUnique(t *testing.T) {
	s := setupArchiveSuite(t)

	const instances = 4
	const perInstance = 30

	var wg sync.WaitGroup
	start := make(chan struct{})
	caseResults := make(chan uint64, instances*perInstance)
	billResults := make(chan uint64, instances*perInstance)
	errs := make(chan error, instances*perInstance*2)

	sharedClient := s.newClient(t)
	sharedLawyer := s.newLawyer(t)
	// 建一个全新共享案件仅供本测试账单写入，避免与其他用例/历史轮次混淆。
	sharedCase, err := s.caseSvc.Create(sharedClient, sharedLawyer,
		"MULTI_BILLING_CASE 多实例账单承载案", constants.CaseTypeCivil, "", nil, nil)
	if err != nil {
		t.Fatalf("create shared case: %v", err)
	}

	titlePrefix := fmt.Sprintf("MULTI_CASE_BATCH_%d_", sharedCase.ID)
	for i := 0; i < instances; i++ {
		// 每个“实例”独立构造 service：不共享任何进程内计数器。
		inst := s.freshInstance(t)
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			for j := 0; j < perInstance; j++ {
				c, err := inst.caseSvc.Create(sharedClient, sharedLawyer,
					fmt.Sprintf("%s i%d-j%d", titlePrefix, idx, j), constants.CaseTypeCivil, "", nil, nil)
				if err != nil {
					errs <- err
					continue
				}
				caseResults <- c.ID
			}
		}(i)
	}
	for i := 0; i < instances; i++ {
		inst := s.freshInstance(t)
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			for j := 0; j < perInstance; j++ {
				b, err := inst.billingSvc.Create(sharedCase.ID, sharedClient,
					constants.BillingTypeAttorneyFee, 1, "")
				if err != nil {
					errs <- err
					continue
				}
				billResults <- b.ID
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(caseResults)
	close(billResults)
	close(errs)

	for e := range errs {
		t.Fatalf("concurrent create failed: %v", e)
	}

	var caseIDs []uint64
	for id := range caseResults {
		caseIDs = append(caseIDs, id)
	}
	var billIDs []uint64
	for id := range billResults {
		billIDs = append(billIDs, id)
	}

	wantCases := instances * perInstance
	wantBills := instances * perInstance
	if len(caseIDs) != wantCases {
		t.Fatalf("created cases = %d, want %d", len(caseIDs), wantCases)
	}
	if len(billIDs) != wantBills {
		t.Fatalf("created billings = %d, want %d", len(billIDs), wantBills)
	}

	// 落库数严格等于成功数（一次成功只落一条，重试没有产生重复行）。
	if got := s.countCasesByTitlePrefix(t, titlePrefix); got != int64(wantCases) {
		t.Fatalf("case rows = %d, want %d", got, wantCases)
	}
	if got := s.countBillingsByCase(t, sharedCase.ID); got != int64(wantBills) {
		t.Fatalf("billing rows on shared case = %d, want %d", got, wantBills)
	}

	// 案件编号全局唯一。
	var caseNos []string
	if err := s.db.Model(&model.Case{}).Where("title LIKE ?", titlePrefix+"%").
		Pluck("case_no", &caseNos).Error; err != nil {
		t.Fatalf("pluck case_no: %v", err)
	}
	assertUnique(t, "case_no", caseNos)

	// 账单编号全局唯一。
	var billNos []string
	if err := s.db.Model(&model.Billing{}).Where("case_id = ?", sharedCase.ID).
		Pluck("bill_no", &billNos).Error; err != nil {
		t.Fatalf("pluck bill_no: %v", err)
	}
	assertUnique(t, "bill_no", billNos)

	// 跨“重启”（再次全新实例）连续运行仍不重号：再建一批，全部编号互不重复。
	restarted := s.freshInstance(t)
	restartCase, err := restarted.caseSvc.Create(sharedClient, sharedLawyer,
		"MULTI_INSTANCE_CASE 重启后建案", constants.CaseTypeCivil, "", nil, nil)
	if err != nil {
		t.Fatalf("create after restart: %v", err)
	}
	for _, no := range caseNos {
		if no == restartCase.CaseNo {
			t.Fatalf("case_no reused after restart: %s", no)
		}
	}
	b2, err := restarted.billingSvc.Create(sharedCase.ID, sharedClient, constants.BillingTypeOther, 1, "")
	if err != nil {
		t.Fatalf("billing after restart: %v", err)
	}
	for _, no := range billNos {
		if no == b2.BillNo {
			t.Fatalf("bill_no reused after restart: %s", no)
		}
	}
}

// ---- 测试辅助 ----

func assertUnique(t *testing.T, field string, values []string) {
	t.Helper()
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		if _, dup := seen[v]; dup {
			t.Fatalf("duplicate %s: %s", field, v)
		}
		seen[v] = struct{}{}
	}
}

func (s *archiveSuite) countCasesByTitlePrefix(t *testing.T, prefix string) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(&model.Case{}).Where("title LIKE ?", prefix+"%").Count(&n).Error; err != nil {
		t.Fatalf("count cases: %v", err)
	}
	return n
}

func (s *archiveSuite) countBillingsByCase(t *testing.T, caseID uint64) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(&model.Billing{}).Where("case_id = ?", caseID).Count(&n).Error; err != nil {
		t.Fatalf("count billings: %v", err)
	}
	return n
}

// freshInstance 基于同一数据库构造一套全新 service（模拟新启动的独立实例）。
func (s *archiveSuite) freshInstance(t *testing.T) *archiveSuite {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	caseRepo := repository.NewCaseRepository(s.db)
	billingRepo := repository.NewBillingRepository(s.db)
	txStore := repository.NewTxStore(s.db)
	return &archiveSuite{
		db:         s.db,
		caseSvc:    NewCaseService(caseRepo, repository.NewClientRepository(s.db), repository.NewUserRepository(s.db), txStore, logger),
		billingSvc: NewBillingService(billingRepo, txStore, logger),
	}
}
