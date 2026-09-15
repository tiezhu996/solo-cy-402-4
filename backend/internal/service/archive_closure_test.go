package service

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// archiveSuite 归档费用收口测试夹具。
type archiveSuite struct {
	db         *gorm.DB
	caseSvc    *CaseService
	billingSvc *BillingService
}

// setupArchiveSuite 通过 CY_TEST_DB_DSN 连接真实 PostgreSQL（行锁/事务语义必须在 PG 上验证）。
// 未配置 DSN 时跳过，保证无数据库环境下 go test ./... 仍可通过。
func setupArchiveSuite(t *testing.T) *archiveSuite {
	t.Helper()
	dsn := os.Getenv("CY_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("skip postgres integration test: CY_TEST_DB_DSN not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Client{}, &model.Case{}, &model.Document{}, &model.Billing{}, &model.AuditLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := repository.NewSequenceRepository(db).EnsureSequences(); err != nil {
		t.Fatalf("ensure sequences: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	caseRepo := repository.NewCaseRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	txStore := repository.NewTxStore(db)
	return &archiveSuite{
		db:         db,
		caseSvc:    NewCaseService(caseRepo, clientRepo, userRepo, txStore, logger),
		billingSvc: NewBillingService(billingRepo, txStore, logger),
	}
}

var fixtureSeq uint64

func (s *archiveSuite) newLawyer(t *testing.T) uint64 {
	t.Helper()
	fixtureSeq++
	u := &model.User{
		Username:     fmt.Sprintf("t_lawyer_%d_%d", time.Now().UnixNano(), fixtureSeq),
		PasswordHash: "x",
		RealName:     "测试律师",
		Role:         constants.RoleLawyer,
	}
	if err := s.db.Create(u).Error; err != nil {
		t.Fatalf("create lawyer: %v", err)
	}
	return u.ID
}

func (s *archiveSuite) newClient(t *testing.T) uint64 {
	t.Helper()
	fixtureSeq++
	c := &model.Client{Name: fmt.Sprintf("测试客户%d", fixtureSeq)}
	if err := s.db.Create(c).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	return c.ID
}

// newClosedCase 建一个已结案案件，返回案件 ID 与结案日期。
func (s *archiveSuite) newClosedCase(t *testing.T) (uint64, *time.Time) {
	t.Helper()
	lawyerID := s.newLawyer(t)
	clientID := s.newClient(t)
	c, err := s.caseSvc.Create(clientID, lawyerID, "归档收口测试案", constants.CaseTypeCivil, "", nil, nil)
	if err != nil {
		t.Fatalf("create case: %v", err)
	}
	for _, st := range []string{constants.CaseStatusInvestigating, constants.CaseStatusHearing, constants.CaseStatusClosed} {
		c, err = s.caseSvc.ChangeStatus(c.ID, constants.RoleLawyer, st)
		if err != nil {
			t.Fatalf("flow to %s: %v", st, err)
		}
	}
	return c.ID, c.CloseDate
}

func (s *archiveSuite) pendingCount(t *testing.T, caseID uint64) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(&model.Billing{}).
		Where("case_id = ? AND status = ?", caseID, constants.BillingStatusPending).Count(&n).Error; err != nil {
		t.Fatalf("count pending: %v", err)
	}
	return n
}

func (s *archiveSuite) caseStatus(t *testing.T, caseID uint64) string {
	t.Helper()
	var c model.Case
	if err := s.db.First(&c, caseID).Error; err != nil {
		t.Fatalf("load case: %v", err)
	}
	return c.Status
}

func appErrorCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *util.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expect AppError, got %T: %v", err, err)
	}
	return appErr.Code
}

// sameTime 比较结案日期：PostgreSQL timestamp 仅微秒精度，截断后再比，且都容忍 nil。
func sameTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Truncate(time.Microsecond).Equal(b.Truncate(time.Microsecond))
}

// TestArchiveRejectedWhilePendingBilling 有待支付账单时拒绝归档，状态与结案日期保持不变。
func TestArchiveRejectedWhilePendingBilling(t *testing.T) {
	s := setupArchiveSuite(t)
	caseID, closeDate := s.newClosedCase(t)

	b, err := s.billingSvc.Create(caseID, s.newClient(t), constants.BillingTypeAttorneyFee, 1000, "")
	if err != nil {
		t.Fatalf("create billing: %v", err)
	}

	_, err = s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
	if err == nil {
		t.Fatal("archive must be rejected when pending billing exists")
	}
	if code := appErrorCode(t, err); code != constants.CodeCaseArchivePendingBilling {
		t.Fatalf("reject code = %d, want %d", code, constants.CodeCaseArchivePendingBilling)
	}
	if got := s.caseStatus(t, caseID); got != constants.CaseStatusClosed {
		t.Fatalf("status after reject = %s, want closed", got)
	}
	var c model.Case
	_ = s.db.First(&c, caseID).Error
	// PostgreSQL timestamp 仅微秒精度，使用辅助函数比较。
	if !sameTime(c.CloseDate, closeDate) {
		t.Fatalf("close_date changed: got %v, want %v", c.CloseDate, closeDate)
	}

	// 补齐支付后允许归档，结案日期仍保持不变。
	if _, err := s.billingSvc.MarkPaid(b.ID); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	archived, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
	if err != nil {
		t.Fatalf("archive after paid: %v", err)
	}
	if archived.Status != constants.CaseStatusArchived {
		t.Fatalf("status = %s, want archived", archived.Status)
	}
	if !sameTime(archived.CloseDate, closeDate) {
		t.Fatalf("close_date changed after archive: got %v, want %v", archived.CloseDate, closeDate)
	}
}

// TestArchiveAllowedAfterVoid 作废待支付账单后可以归档。
func TestArchiveAllowedAfterVoid(t *testing.T) {
	s := setupArchiveSuite(t)
	caseID, _ := s.newClosedCase(t)
	clientID := s.newClient(t)
	b, err := s.billingSvc.Create(caseID, clientID, constants.BillingTypeCourtFee, 200, "")
	if err != nil {
		t.Fatalf("create billing: %v", err)
	}
	if _, err := s.billingSvc.Void(b.ID); err != nil {
		t.Fatalf("void billing: %v", err)
	}
	c, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
	if err != nil {
		t.Fatalf("archive after void: %v", err)
	}
	if c.Status != constants.CaseStatusArchived {
		t.Fatalf("status = %s, want archived", c.Status)
	}
}

// TestCreateBillingBlockedAfterArchive 归档后不得再新增待支付账单。
func TestCreateBillingBlockedAfterArchive(t *testing.T) {
	s := setupArchiveSuite(t)
	caseID, _ := s.newClosedCase(t)
	if _, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived); err != nil {
		t.Fatalf("archive: %v", err)
	}
	_, err := s.billingSvc.Create(caseID, s.newClient(t), constants.BillingTypeOther, 50, "")
	if err == nil {
		t.Fatal("create billing on archived case must be rejected")
	}
	if code := appErrorCode(t, err); code != constants.CodeBillingCreateArchivedCase {
		t.Fatalf("block code = %d, want %d", code, constants.CodeBillingCreateArchivedCase)
	}
	if n := s.pendingCount(t, caseID); n != 0 {
		t.Fatalf("pending count on archived case = %d, want 0", n)
	}
}

// TestArchiveIdempotent 重复归档结果不变。
func TestArchiveIdempotent(t *testing.T) {
	s := setupArchiveSuite(t)
	caseID, closeDate := s.newClosedCase(t)
	for i := 0; i < 3; i++ {
		c, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
		if err != nil {
			t.Fatalf("archive attempt %d: %v", i+1, err)
		}
		if c.Status != constants.CaseStatusArchived {
			t.Fatalf("attempt %d status = %s", i+1, c.Status)
		}
		if !sameTime(c.CloseDate, closeDate) {
			t.Fatalf("attempt %d close_date drifted", i+1)
		}
	}
}

// TestConcurrentArchiveVsCreate 归档与新建账单同时到达：
// 无论谁先拿到案件行锁，最终只能是“已归档且无待支付账单”或“已结案且存在待支付账单”两种稳定终态，
// 绝不能出现“已归档案件仍挂待支付账单”；且对终态重复重试结果不变。
func TestConcurrentArchiveVsCreate(t *testing.T) {
	s := setupArchiveSuite(t)

	const rounds = 12
	for round := 0; round < rounds; round++ {
		caseID, _ := s.newClosedCase(t)
		clientID := s.newClient(t)

		var wg sync.WaitGroup
		start := make(chan struct{})

		// 1 个归档者。
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
		}()

		// 多个新建账单者，在案件未归档前持续创建。
		const creators = 8
		for i := 0; i < creators; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				for j := 0; j < 30; j++ {
					_, _ = s.billingSvc.Create(caseID, clientID, constants.BillingTypeAttorneyFee, 10, "")
				}
			}()
		}
		close(start)
		wg.Wait()

		status := s.caseStatus(t, caseID)
		pending := s.pendingCount(t, caseID)

		// 核心不变量：已归档 ⇒ 无待支付账单。
		if status == constants.CaseStatusArchived && pending != 0 {
			t.Fatalf("round %d: archived case still has %d pending billings", round, pending)
		}

		switch status {
		case constants.CaseStatusArchived:
			// 终态 1：已归档。重试创建必须始终被拒绝，重复归档幂等。
			for i := 0; i < 3; i++ {
				_, err := s.billingSvc.Create(caseID, clientID, constants.BillingTypeOther, 1, "")
				if err == nil || appErrorCode(t, err) != constants.CodeBillingCreateArchivedCase {
					t.Fatalf("round %d: create on archived must be 40904, got %v", round, err)
				}
				c, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
				if err != nil || c.Status != constants.CaseStatusArchived {
					t.Fatalf("round %d: re-archive must stay archived, got %v %v", round, c, err)
				}
			}
			if n := s.pendingCount(t, caseID); n != 0 {
				t.Fatalf("round %d: pending = %d after retries", round, n)
			}
		case constants.CaseStatusClosed:
			// 终态 2：已结案且挂有待支付账单。重复归档必须始终被拒绝且状态不变。
			if pending == 0 {
				// 理论上归档失败必因为至少 1 笔 pending；防御性检查。
				t.Fatalf("round %d: closed case unexpectedly has no pending billings", round)
			}
			for i := 0; i < 3; i++ {
				_, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
				if err == nil || appErrorCode(t, err) != constants.CodeCaseArchivePendingBilling {
					t.Fatalf("round %d: re-archive must be 40903, got %v", round, err)
				}
				if got := s.caseStatus(t, caseID); got != constants.CaseStatusClosed {
					t.Fatalf("round %d: status drifted to %s after reject", round, got)
				}
			}
			// 补齐全部待支付账单（支付或作废）后才能归档成功，之后创建被永久拒绝。
			var bills []model.Billing
			if err := s.db.Where("case_id = ? AND status = ?", caseID, constants.BillingStatusPending).Find(&bills).Error; err != nil {
				t.Fatalf("load pending bills: %v", err)
			}
			for _, b := range bills {
				if _, err := s.billingSvc.MarkPaid(b.ID); err != nil {
					t.Fatalf("mark paid: %v", err)
				}
			}
			c, err := s.caseSvc.ChangeStatus(caseID, constants.RoleLawyer, constants.CaseStatusArchived)
			if err != nil {
				t.Fatalf("archive after settle: %v", err)
			}
			if c.Status != constants.CaseStatusArchived || s.pendingCount(t, caseID) != 0 {
				t.Fatalf("round %d: not converged after settlement", round)
			}
			_, err = s.billingSvc.Create(caseID, clientID, constants.BillingTypeOther, 1, "")
			if err == nil || appErrorCode(t, err) != constants.CodeBillingCreateArchivedCase {
				t.Fatalf("round %d: create must be blocked post-archive, got %v", round, err)
			}
		default:
			t.Fatalf("round %d: unexpected status %s", round, status)
		}
	}
}
