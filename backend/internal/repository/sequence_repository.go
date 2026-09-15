package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// 业务编号序列名：案件编号与账单编号各一条 PostgreSQL sequence。
// 序列值由数据库集中分配，天然支持多实例、多次重启与同一数据库连续运行下的唯一性，
// 不依赖进程内计数器（进程内计数器无法跨实例/重启保持唯一）。
const (
	CaseNoSeqName = "case_no_seq"
	BillNoSeqName = "bill_no_seq"
)

// SequenceRepository 业务编号序列仓储。
type SequenceRepository struct {
	db *gorm.DB
}

// NewSequenceRepository 构造序列仓储。
func NewSequenceRepository(db *gorm.DB) *SequenceRepository {
	return &SequenceRepository{db: db}
}

// EnsureSequences 幂等创建业务编号序列，并把序列起点对齐历史数据中的最大数字后缀。
// CREATE SEQUENCE IF NOT EXISTS 保证：多次启动重复执行无副作用；多个实例启动竞争时，
// 已存在的序列会被复用而不是报错（仅忽略 duplicate_table/duplicate_object 两类冲突）。
// 对齐历史最大值保证：老版本时间戳编号升级到序列发号、或序列被意外重建后，
// nextval 也不会复用已被占用的编号——同一数据库连续运行即可，无需清库或人工改号。
// CACHE 1 保证分配出的值在并发下严格递增；回滚事务会放弃已取号（产生空洞），不影响唯一性。
func (r *SequenceRepository) EnsureSequences() error {
	stmts := []string{
		`CREATE SEQUENCE IF NOT EXISTS ` + CaseNoSeqName + ` AS BIGINT START WITH 1 INCREMENT BY 1 CACHE 1`,
		`CREATE SEQUENCE IF NOT EXISTS ` + BillNoSeqName + ` AS BIGINT START WITH 1 INCREMENT BY 1 CACHE 1`,
	}
	for _, stmt := range stmts {
		if err := r.db.Exec(stmt).Error; err != nil {
			// 容忍并发创建：42P07/42P06 对象已存在，以及 IF NOT EXISTS 仍可能抛出的
			// pg_class 唯一冲突 23505（两个后端同时 CREATE）。出现任一都说明序列已建好，直接复用。
			if !IsAlreadyExistsError(err) {
				return fmt.Errorf("ensure sequence: %w", err)
			}
		}
	}
	// 标识符为代码内固定常量，不存在用户输入，可安全内联到 SQL；参数仅传数值。
	type spec struct{ seq, table, column string }
	specs := []spec{
		{CaseNoSeqName, "cases", "case_no"},
		{BillNoSeqName, "billings", "bill_no"},
	}
	for _, sp := range specs {
		if err := r.alignSequence(sp.seq, sp.table, sp.column); err != nil {
			return err
		}
	}
	return nil
}

// alignSequence 将序列当前值提升到“历史编号最大数字后缀”与“序列当前值”的较大者。
// 编号前缀（字母 + 4 位年份）长度：案件为 6（CY+yyyy），账单为 8（BILL+yyyy），数字后缀取其后部分。
func (r *SequenceRepository) alignSequence(seq, table, column string) error {
	prefixLen := 6
	if seq == BillNoSeqName {
		prefixLen = 8
	}
	var maxSuffix int64
	stmt := fmt.Sprintf(
		`SELECT COALESCE(MAX(NULLIF(regexp_replace(substr(%s, ?), '[^0-9]', '', 'g'), '')::bigint), 0) FROM %s`,
		column, table)
	if err := r.db.Raw(stmt, prefixLen+1).Scan(&maxSuffix).Error; err != nil {
		return fmt.Errorf("align sequence %s scan: %w", seq, err)
	}
	if maxSuffix <= 0 {
		return nil
	}
	// setval 取 GREATEST(序列现值, 历史最大值)，绝不回退序列，避免删除数据后重新发号撞号。
	if err := r.db.Exec(
		fmt.Sprintf(`SELECT setval(?::regclass, GREATEST((SELECT last_value FROM %s), ?), true)`, seq),
		seq, maxSuffix).Error; err != nil {
		return fmt.Errorf("align sequence %s setval: %w", seq, err)
	}
	return nil
}

// NextValue 在当前连接/事务上取下一个序列值（nextval）。
// nextval 自身不在事务回滚时退还：失败重试必须重新取号，不能复用旧值。
func (r *SequenceRepository) NextValue(name string) (int64, error) {
	var v int64
	if err := r.db.Raw("SELECT nextval(?::regclass)", name).Scan(&v).Error; err != nil {
		return 0, fmt.Errorf("nextval %s: %w", name, err)
	}
	return v, nil
}

// IsAlreadyExistsError 是否为“对象已存在”错误（并发创建表/序列时容忍）。
func IsAlreadyExistsError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "42P06", "42P07":
			// duplicate_schema / duplicate_table：对象已被其他实例抢先创建。
			return true
		case "23505":
			// unique_violation：并发 CREATE 命中 pg_class_relname_nsp_index 等目录唯一索引，
			// 语义等价于“已存在”，调用方应复用而不是让实例退出。
			return true
		}
	}
	return false
}

// IsUniqueViolation 是否为唯一约束冲突（SQLSTATE 23505）。
// 编号撞唯一索引时用它识别，从而整事务回滚并用新 nextval 重试，而不是把 500 抛给调用方。
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
