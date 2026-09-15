package service

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"
)

// BillingService 账单业务逻辑。
type BillingService struct {
	repo       *repository.BillingRepository
	caseRepo   *repository.CaseRepository
	clientRepo *repository.ClientRepository
	logger     *slog.Logger
}

// NewBillingService 构造账单服务。
func NewBillingService(repo *repository.BillingRepository, caseRepo *repository.CaseRepository,
	clientRepo *repository.ClientRepository, logger *slog.Logger) *BillingService {
	return &BillingService{repo: repo, caseRepo: caseRepo, clientRepo: clientRepo, logger: logger}
}

// Create 创建账单。
func (s *BillingService) Create(caseID, clientID uint64, billingType string, amount float64, invoiceInfo string) (*model.Billing, error) {
	if !constants.IsValidBillingType(billingType) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Billing[billing_type="+billingType+"] create: invalid type")
	}
	if _, err := s.caseRepo.FindByID(caseID); err != nil {
		return nil, util.Wrap(err, "Billing[case_id=%d] create: case not found", caseID)
	}
	if _, err := s.clientRepo.FindByID(clientID); err != nil {
		return nil, util.Wrap(err, "Billing[client_id=%d] create: client not found", clientID)
	}
	if amount < 0 {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Billing[amount="+strconv.FormatFloat(amount, 'f', 2, 64)+"] create: amount must be >= 0")
	}
	b := &model.Billing{BillNo: genBillNo(), BillingType: billingType,
		Amount: amount, Status: constants.BillingStatusPending,
		CaseID: caseID, ClientID: clientID, InvoiceInfo: invoiceInfo}
	if err := s.repo.Create(b); err != nil {
		s.logger.Error(constants.LogBillingCreateFailed, "error", err.Error())
		return nil, util.Wrap(err, "Billing[case_id=%d] create failed", caseID)
	}
	s.logger.Info(constants.LogBillingCreateSuccess, "billing_id", b.ID, "bill_no", b.BillNo)
	return b, nil
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
	return fmt.Sprintf("BILL%d%06d", time.Now().Year(), time.Now().UnixNano()%1000000)
}
