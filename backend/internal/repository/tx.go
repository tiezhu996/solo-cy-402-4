package repository

import "gorm.io/gorm"

// TxStore 事务与仓储的装配器：service 通过它在一个数据库事务内组合多个实体的仓储，
// 解决“案件归档与账单状态必须在同一事务内收口”这类跨实体一致性问题，
// 同时保持依赖方向 service -> repository -> model，service 不直接依赖 gorm。
type TxStore struct {
	db *gorm.DB
}

// NewTxStore 构造事务装配器。
func NewTxStore(db *gorm.DB) *TxStore {
	return &TxStore{db: db}
}

// TxRepos 事务内可用的全部仓储集合。
type TxRepos struct {
	Cases     *CaseRepository
	Billings  *BillingRepository
	Clients   *ClientRepository
	Users     *UserRepository
	Sequences *SequenceRepository
}

// WithinTx 在一个数据库事务内执行 fn：
// fn 返回 error 时回滚，否则提交；fn 内拿到的仓储与事务绑定（同一连接，行锁/nextval 生效）。
func (s *TxStore) WithinTx(fn func(tx *TxRepos) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fn(&TxRepos{
			Cases:     NewCaseRepository(tx),
			Billings:  NewBillingRepository(tx),
			Clients:   NewClientRepository(tx),
			Users:     NewUserRepository(tx),
			Sequences: NewSequenceRepository(tx),
		})
	})
}
