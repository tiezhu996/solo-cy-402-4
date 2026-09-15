package repository

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cylawcase/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// allModels 与 cmd/server/main.go 的 AutoMigrate 列表保持一致。
var allModels = []any{
	&model.User{}, &model.Client{}, &model.Case{}, &model.Document{}, &model.Billing{}, &model.AuditLog{},
}

func bootDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("CY_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("skip postgres bootstrap test: CY_TEST_DB_DSN not set")
	}
	return dsn
}

func openGorm(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

// withDBName 把 keyword 风格 DSN 的 dbname 替换为指定库。
func withDBName(dsn, name string) string {
	if strings.Contains(dsn, "dbname=") {
		return strings.Replace(dsn, "dbname="+bootDBName(dsn), "dbname="+name, 1)
	}
	return dsn + " dbname=" + name
}

func bootDBName(dsn string) string {
	for _, kv := range strings.Fields(dsn) {
		if strings.HasPrefix(kv, "dbname=") {
			return strings.TrimPrefix(kv, "dbname=")
		}
	}
	return ""
}

// freshBootstrapDB 建一个一次性空库，返回连接与清理函数（FORCE 删库）。
func freshBootstrapDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	adminDSN := bootDSN(t)
	admin := openGorm(t, adminDSN)
	name := fmt.Sprintf("cy_boot_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	db := openGorm(t, withDBName(adminDSN, name))
	cleanup := func() {
		_ = db.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ? AND pid <> pg_backend_pid()", name).Error
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		// 连回管理库删库（CREATE/DROP DATABASE 需在目标库之外执行）。
		_ = admin.Exec("DROP DATABASE IF EXISTS " + name).Error
		sqlAdmin, _ := admin.DB()
		_ = sqlAdmin.Close()
	}
	return db, cleanup
}

func bootLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestBootstrapLockSerializes 多个实例同时引导一个“缺表缺序列”的旧库：
// 必须串行执行，任何一方都不得返回错误退出，等待方最终复用已建好的表与序列继续。
func TestBootstrapLockSerializes(t *testing.T) {
	db, cleanup := freshBootstrapDB(t)
	defer cleanup()

	const instances = 5
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, instances)
	var inFlight, maxFlight int32

	for i := 0; i < instances; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			err := WithBootstrapLock(db, bootLogger(), func() error {
				cur := atomic.AddInt32(&inFlight, 1)
				// 记录临界区内的最大并发数；持锁时必须恒为 1。
				for {
					old := atomic.LoadInt32(&maxFlight)
					if cur <= old || atomic.CompareAndSwapInt32(&maxFlight, old, cur) {
						break
					}
				}
				time.Sleep(80 * time.Millisecond) // 拉大竞争窗口，暴露未串行化问题
				atomic.AddInt32(&inFlight, -1)
				if err := db.AutoMigrate(allModels...); err != nil {
					return err
				}
				return NewSequenceRepository(db).EnsureSequences()
			})
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("no instance may fail during bootstrap: %v", err)
		}
	}
	if got := atomic.LoadInt32(&maxFlight); got != 1 {
		t.Fatalf("bootstrap critical section max concurrency = %d, want 1", got)
	}
	// 等待方复用：表与序列最终恰好存在一份。
	var seqCount int64
	if err := db.Raw(`SELECT count(*) FROM pg_class WHERE relkind = 'S' AND relname IN (?, ?)`,
		CaseNoSeqName, BillNoSeqName).Scan(&seqCount).Error; err != nil {
		t.Fatalf("count sequences: %v", err)
	}
	if seqCount != 2 {
		t.Fatalf("sequences = %d, want 2", seqCount)
	}
}

// TestBootstrapLockReleased 锁在引导结束后必须释放，后续实例可立即进入，不被阻塞。
func TestBootstrapLockReleased(t *testing.T) {
	db, cleanup := freshBootstrapDB(t)
	defer cleanup()

	ran := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		if err := WithBootstrapLock(db, bootLogger(), func() error {
			ran <- struct{}{}
			return nil
		}); err != nil {
			t.Fatalf("sequential bootstrap %d failed: %v", i+1, err)
		}
	}
	if len(ran) != 2 {
		t.Fatalf("bootstrap lock not released, second caller blocked")
	}
}

// TestConcurrentEnsureSequencesToleratesCreateRace 旧库缺序列时多个连接并发建序列：
// 任何一方都不得因 23505/42P07 失败退出，最终两条序列各仅一份，可正常 nextval。
func TestConcurrentEnsureSequencesToleratesCreateRace(t *testing.T) {
	db, cleanup := freshBootstrapDB(t)
	defer cleanup()
	if err := db.AutoMigrate(allModels...); err != nil {
		t.Fatalf("migrate: %v", err)
	} // 只建表，不建序列

	const n = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errCh <- NewSequenceRepository(db).EnsureSequences()
		}()
	}
	close(start)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent ensure sequences must not fail: %v", err)
		}
	}

	var seqCount int64
	if err := db.Raw(`SELECT count(*) FROM pg_class WHERE relkind = 'S' AND relname IN (?, ?)`,
		CaseNoSeqName, BillNoSeqName).Scan(&seqCount).Error; err != nil {
		t.Fatalf("count sequences: %v", err)
	}
	if seqCount != 2 {
		t.Fatalf("sequences = %d, want 2", seqCount)
	}
	for _, name := range []string{CaseNoSeqName, BillNoSeqName} {
		var v int64
		if err := db.Raw("SELECT nextval(?::regclass)", name).Scan(&v).Error; err != nil {
			t.Fatalf("nextval %s: %v", name, err)
		}
		if v <= 0 {
			t.Fatalf("%s nextval = %d, want > 0", name, v)
		}
	}
}

// TestSequenceAlignmentForwardOnly 对齐只能按已有编号向后取 max，不回退、不覆盖已有记录。
func TestSequenceAlignmentForwardOnly(t *testing.T) {
	db, cleanup := freshBootstrapDB(t)
	defer cleanup()
	if err := db.AutoMigrate(allModels...); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// 历史混合编号：旧时间戳格式与新序列格式共存。
	legacyCases := []model.Case{
		{CaseNo: "CY20260050", Title: "legacy1", CaseType: "civil", Status: "closed", ClientID: 1, LeadLawyerID: 1},
		{CaseNo: "CY20260001", Title: "legacy2", CaseType: "civil", Status: "filed", ClientID: 1, LeadLawyerID: 1},
		{CaseNo: "CY202600000300", Title: "newfmt", CaseType: "civil", Status: "filed", ClientID: 1, LeadLawyerID: 1},
	}
	for i := range legacyCases {
		if err := db.Create(&legacyCases[i]).Error; err != nil {
			t.Fatalf("create case: %v", err)
		}
	}
	legacyBills := []model.Billing{
		{BillNo: "BILL2026080001", BillingType: "court_fee", Amount: 1, Status: "paid", CaseID: 1, ClientID: 1},
		{BillNo: "BILL2026000000009", BillingType: "other", Amount: 1, Status: "void", CaseID: 1, ClientID: 1},
	}
	for i := range legacyBills {
		if err := db.Create(&legacyBills[i]).Error; err != nil {
			t.Fatalf("create billing: %v", err)
		}
	}

	seq := NewSequenceRepository(db)
	if err := seq.EnsureSequences(); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	// 对齐到历史最大数字后缀：案件最大后缀 300 -> 下一个 301；账单最大后缀 80001 -> 下一个 80002。
	if v := mustNextval(t, db, CaseNoSeqName); v != 301 {
		t.Fatalf("next case seq = %d, want 301", v)
	}
	if v := mustNextval(t, db, BillNoSeqName); v != 80002 {
		t.Fatalf("next bill seq = %d, want 80002", v)
	}

	// 序列已被运行期 nextval 推到更高位置后，再次对齐不得回退（历史后缀远小于当前值）。
	if err := db.Exec("SELECT setval(?::regclass, ?)", CaseNoSeqName, int64(999999)).Error; err != nil {
		t.Fatalf("setval high: %v", err)
	}
	if err := seq.EnsureSequences(); err != nil {
		t.Fatalf("re-ensure: %v", err)
	}
	var last int64
	if err := db.Raw("SELECT last_value FROM " + CaseNoSeqName).Scan(&last).Error; err != nil {
		t.Fatalf("read last_value: %v", err)
	}
	if last != 999999 {
		t.Fatalf("sequence regressed to %d, want stay 999999 (align forward-only)", last)
	}

	// 对齐过程绝不覆盖/删除已有记录：编号与数量原样保留。
	var caseCount, billCount int64
	db.Model(&model.Case{}).Count(&caseCount)
	db.Model(&model.Billing{}).Count(&billCount)
	if caseCount != int64(len(legacyCases)) || billCount != int64(len(legacyBills)) {
		t.Fatalf("existing records mutated: cases=%d bills=%d", caseCount, billCount)
	}
	for _, want := range []string{"CY20260050", "CY20260001", "CY202600000300"} {
		var n int64
		db.Model(&model.Case{}).Where("case_no = ?", want).Count(&n)
		if n != 1 {
			t.Fatalf("existing case_no %s not preserved (count=%d)", want, n)
		}
	}
}

func mustNextval(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	var v int64
	if err := db.Raw("SELECT nextval(?::regclass)", name).Scan(&v).Error; err != nil {
		t.Fatalf("nextval %s: %v", name, err)
	}
	return v
}
