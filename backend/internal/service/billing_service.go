package service

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"
)

// BillingService 账单业务逻辑。
type BillingService struct {
	repo    *repository.BillingRepository
	txStore *repository.TxStore
	logger  *slog.Logger
}

// NewBillingService 构造账单服务。
func NewBillingService(repo *repository.BillingRepository, txStore *repository.TxStore,
	logger *slog.Logger) *BillingService {
	return &BillingService{repo: repo, txStore: txStore, logger: logger}
}

// Create 创建账单。
// 案件已归档后不得再新增待支付账单：与归档共用同一把案件行锁，
// 保证“归档与新建账单同时到达”时二者串行，收口为唯一稳定终态。
func (s *BillingService) Create(caseID, clientID uint64, billingType string, amount float64, invoiceInfo string) (*model.Billing, error) {
	if !constants.IsValidBillingType(billingType) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Billing[billing_type="+billingType+"] create: invalid type")
	}
	if amount < 0 {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Billing[amount="+strconv.FormatFloat(amount, 'f', 2, 64)+"] create: amount must be >= 0")
	}
	var created *model.Billing
	err := s.txStore.WithinTx(func(tx *repository.TxRepos) error {
		// 锁案件行，与归档事务互斥串行。
		c, err := tx.Cases.FindByIDForUpdate(caseID)
		if err != nil {
			return util.Wrap(err, "Billing[case_id=%d] create: case not found", caseID)
		}
		if c.Status == constants.CaseStatusArchived {
			s.logger.Warn(constants.LogBillingCreateBlocked,
				"case_id", caseID, "status", c.Status, "client_id", clientID)
			return util.NewAppError(constants.CodeBillingCreateArchivedCase,
				"Billing[case_id="+u64(caseID)+"] create rejected: case already archived")
		}
		if _, err := tx.Clients.FindByID(clientID); err != nil {
			return util.Wrap(err, "Billing[client_id=%d] create: client not found", clientID)
		}
		b := &model.Billing{BillNo: genBillNo(), BillingType: billingType,
			Amount: amount, Status: constants.BillingStatusPending,
			CaseID: caseID, ClientID: clientID, InvoiceInfo: invoiceInfo}
		if err := tx.Billings.Create(b); err != nil {
			s.logger.Error(constants.LogBillingCreateFailed, "error", err.Error())
			return util.Wrap(err, "Billing[case_id=%d] create failed", caseID)
		}
		created = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogBillingCreateSuccess, "billing_id", created.ID, "bill_no", created.BillNo)
	return created, nil
}

// MarkPaid 标记支付（pending -> paid）。
func (s *BillingService) MarkPaid(id uint64) (*model.Billing, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Billing[id=%d] paid find failed", id)
	}
	if b.Status != constants.BillingStatusPending {
		s.logger.Warn(constants.LogBillingPaidFailed, "billing_id", id, "status", b.Status)
		return nil, util.NewAppError(constants.CodeBillingStatusConflict, "Billing[id="+u64(id)+"] paid failed: status="+b.Status)
	}
	b.Status = constants.BillingStatusPaid
	if err := s.repo.Update(b); err != nil {
		return nil, util.Wrap(err, "Billing[id=%d] paid save failed", id)
	}
	s.logger.Info(constants.LogBillingPaidSuccess, "billing_id", b.ID)
	return b, nil
}

// MarkInvoiced 开票（paid -> invoiced）。
func (s *BillingService) MarkInvoiced(id uint64, invoiceInfo string) (*model.Billing, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Billing[id=%d] invoiced find failed", id)
	}
	if b.Status != constants.BillingStatusPaid {
		return nil, util.NewAppError(constants.CodeBillingStatusConflict, "Billing[id="+u64(id)+"] invoiced failed: status="+b.Status)
	}
	b.Status = constants.BillingStatusInvoiced
	if invoiceInfo != "" {
		b.InvoiceInfo = invoiceInfo
	}
	if err := s.repo.Update(b); err != nil {
		return nil, util.Wrap(err, "Billing[id=%d] invoiced save failed", id)
	}
	s.logger.Info(constants.LogBillingInvoicedSuccess, "billing_id", b.ID)
	return b, nil
}

// Void 作废账单。
func (s *BillingService) Void(id uint64) (*model.Billing, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Billing[id=%d] void find failed", id)
	}
	if b.Status == constants.BillingStatusVoid {
		return nil, util.NewAppError(constants.CodeBillingStatusConflict, "Billing[id="+u64(id)+"] void failed: already void")
	}
	b.Status = constants.BillingStatusVoid
	if err := s.repo.Update(b); err != nil {
		return nil, util.Wrap(err, "Billing[id=%d] void save failed", id)
	}
	s.logger.Info(constants.LogBillingVoidSuccess, "billing_id", b.ID)
	return b, nil
}

// List 分页查询账单。
func (s *BillingService) List(page, pageSize int, caseID, clientID uint64, status string) ([]model.Billing, int64, error) {
	return s.repo.List(page, pageSize, caseID, clientID, status)
}

// ListByCase 查询某案件账单。
func (s *BillingService) ListByCase(caseID uint64) ([]model.Billing, error) {
	return s.repo.ListByCase(caseID)
}

// Summary 本月应收/已收/待收汇总。
func (s *BillingService) Summary() (map[string]float64, error) {
	sum, err := s.repo.Summary(time.Now())
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogBillingSummary, "summary", fmt.Sprintf("%v", sum))
	return sum, nil
}

func genBillNo() string {
	// 原子序号兜底：并发新建账单时即便纳秒相同也不会撞 bill_no 唯一索引。
	seq := billNoSeq.Add(1) % 1000
	return fmt.Sprintf("BILL%d%06d%03d", time.Now().Year(), time.Now().UnixNano()%1000000, seq)
}

// billNoSeq bill_no 进程内发号器，防止并发创建时单号碰撞。
var billNoSeq atomic.Uint64
