package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"
)

// CaseService 案件业务逻辑。
type CaseService struct {
	repo       *repository.CaseRepository
	clientRepo *repository.ClientRepository
	userRepo   *repository.UserRepository
	txStore    *repository.TxStore
	logger     *slog.Logger
}

// NewCaseService 构造案件服务。
func NewCaseService(repo *repository.CaseRepository, clientRepo *repository.ClientRepository,
	userRepo *repository.UserRepository, txStore *repository.TxStore, logger *slog.Logger) *CaseService {
	return &CaseService{repo: repo, clientRepo: clientRepo, userRepo: userRepo,
		txStore: txStore, logger: logger}
}

// Create 创建案件。
// 案件编号在事务内由数据库序列分配：多实例/多次启动下不会重号；
// 撞唯一索引时整事务回滚并用全新 nextval 重试，成功严格落一条记录。
func (s *CaseService) Create(clientID, leadLawyerID uint64, title, caseType, summary string,
	acceptDate *time.Time, coLawyerIDs []uint64) (*model.Case, error) {
	if !constants.IsValidCaseType(caseType) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Case[case_type="+caseType+"] create: invalid type")
	}
	co := jsonCoLawyers(coLawyerIDs)
	var created *model.Case
	err := runWithNumberTx(s.logger, s.txStore, "Case", func(tx *repository.TxRepos) error {
		if _, err := tx.Clients.FindByID(clientID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound,
					fmt.Sprintf("Case[client_id=%d] create: client not found", clientID))
			}
			return util.Wrap(err, "Case[client_id=%d] create: client not found", clientID)
		}
		if _, err := tx.Users.FindByID(leadLawyerID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound,
					fmt.Sprintf("Case[lead_lawyer_id=%d] create: lawyer not found", leadLawyerID))
			}
			return util.Wrap(err, "Case[lead_lawyer_id=%d] create: lawyer not found", leadLawyerID)
		}
		caseNo, err := nextCaseNo(tx)
		if err != nil {
			s.logger.Error(constants.LogNumberAllocateFailed, "entity", "Case", "error", err.Error())
			return err // 经 runWithNumberTx 归为 50301 可重试
		}
		c := &model.Case{
			CaseNo:       caseNo,
			Title:        title,
			CaseType:     caseType,
			Status:       constants.CaseStatusFiled,
			AcceptDate:   acceptDate,
			Summary:      summary,
			ClientID:     clientID,
			LeadLawyerID: leadLawyerID,
			CoLawyerIDs:  co,
		}
		if err := tx.Cases.Create(c); err != nil {
			s.logger.Error(constants.LogCaseCreateFailed, "error", err.Error())
			return util.Wrap(err, "Case[title=%s] create failed", title)
		}
		created = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogCaseCreateSuccess, "case_id", created.ID, "case_no", created.CaseNo)
	return created, nil
}

// Update 更新案件信息。
func (s *CaseService) Update(id uint64, title, summary string, coLawyerIDs []uint64) (*model.Case, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Case[id=%d] update find failed", id)
	}
	if title != "" {
		c.Title = title
	}
	if summary != "" {
		c.Summary = summary
	}
	if coLawyerIDs != nil {
		c.CoLawyerIDs = jsonCoLawyers(coLawyerIDs)
	}
	if err := s.repo.Update(c); err != nil {
		return nil, util.Wrap(err, "Case[id=%d] update save failed", id)
	}
	s.logger.Info(constants.LogCaseUpdateSuccess, "case_id", c.ID)
	return c, nil
}

// ChangeStatus 案件状态流转。
// 进入“已归档”前必须完成费用收口：案件仍存在 pending 账单时拒绝归档，
// 案件状态与结案日期保持原样；补齐支付或作废后重试方可归档成功。
func (s *CaseService) ChangeStatus(id uint64, operatorRole string, status string) (*model.Case, error) {
	if !constants.IsValidCaseStatus(status) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Case[id="+u64(id)+"] status invalid: "+status)
	}
	if status == constants.CaseStatusArchived {
		return s.archiveWithBillingClosure(id, operatorRole)
	}
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Case[id=%d] status change find failed", id)
	}
	if operatorRole != constants.RoleAdmin && !canFlow(c.Status, status) {
		return nil, util.NewAppError(constants.CodeCaseStatusConflict, "Case[id="+u64(id)+"] status conflict: "+c.Status+" -> "+status)
	}
	c.Status = status
	if status == constants.CaseStatusClosed && c.CloseDate == nil {
		now := time.Now()
		c.CloseDate = &now
	}
	if err := s.repo.Update(c); err != nil {
		s.logger.Error(constants.LogCaseStatusChangeFailed, "error", err.Error())
		return nil, util.Wrap(err, "Case[id=%d] status change save failed", id)
	}
	s.logger.Info(constants.LogCaseStatusChangeSuccess, "case_id", c.ID, "status", status)
	return c, nil
}

// archiveWithBillingClosure 在事务内完成归档费用收口。
// 先对案件行加排他锁，使“归档”与“新建账单”两类并发请求在同一案件上串行执行，
// 冲突只能收口为两种稳定终态之一：仍有 pending 账单 -> 案件保持已结案；
// 无 pending 账单 -> 案件归档。任一方重试都得到相同终态，不会出现“已归档挂待支付账单”。
func (s *CaseService) archiveWithBillingClosure(id uint64, operatorRole string) (*model.Case, error) {
	var result *model.Case
	err := s.txStore.WithinTx(func(tx *repository.TxRepos) error {
		c, err := tx.Cases.FindByIDForUpdate(id)
		if err != nil {
			return util.Wrap(err, "Case[id=%d] archive find failed", id)
		}
		// 幂等：已是归档终态直接成功，重复归档结果不变。
		if c.Status == constants.CaseStatusArchived {
			result = c
			return nil
		}
		// 锁内复核流转规则，避免锁外快照过期导致越权跳状态。
		if operatorRole != constants.RoleAdmin && !canFlow(c.Status, constants.CaseStatusArchived) {
			return util.NewAppError(constants.CodeCaseStatusConflict,
				"Case[id="+u64(id)+"] status conflict: "+c.Status+" -> "+constants.CaseStatusArchived)
		}
		pending, err := tx.Billings.CountPendingByCase(id)
		if err != nil {
			return util.Wrap(err, "Case[id=%d] archive count pending billings failed", id)
		}
		if pending > 0 {
			// 拒绝归档：不写库，案件状态与结案日期保持不变。
			s.logger.Warn(constants.LogCaseArchiveRejected,
				"case_id", id, "status", c.Status, "pending_billings", pending)
			return util.NewAppError(constants.CodeCaseArchivePendingBilling,
				"Case[id="+u64(id)+"] archive rejected: pending_billings="+fmt.Sprintf("%d", pending))
		}
		c.Status = constants.CaseStatusArchived
		if err := tx.Cases.Update(c); err != nil {
			s.logger.Error(constants.LogCaseStatusChangeFailed, "error", err.Error())
			return util.Wrap(err, "Case[id=%d] archive save failed", id)
		}
		result = c
		s.logger.Info(constants.LogCaseStatusChangeSuccess, "case_id", c.ID, "status", c.Status)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Assign 分配主办律师。
func (s *CaseService) Assign(id, leadLawyerID uint64, coLawyerIDs []uint64) (*model.Case, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Case[id=%d] assign find failed", id)
	}
	if _, err := s.userRepo.FindByID(leadLawyerID); err != nil {
		return nil, util.Wrap(err, "Case[id=%d] assign failed: lawyer not match", id)
	}
	c.LeadLawyerID = leadLawyerID
	if coLawyerIDs != nil {
		c.CoLawyerIDs = jsonCoLawyers(coLawyerIDs)
	}
	if err := s.repo.Update(c); err != nil {
		s.logger.Error(constants.LogCaseAssignFailed, "error", err.Error())
		return nil, util.Wrap(err, "Case[id=%d] assign save failed", id)
	}
	s.logger.Info(constants.LogCaseAssignSuccess, "case_id", c.ID, "lead_lawyer_id", leadLawyerID)
	return c, nil
}

// List 分页查询案件。
func (s *CaseService) List(page, pageSize int, caseType, status string, lawyerID uint64, startDate, endDate *time.Time) ([]model.Case, int64, error) {
	return s.repo.List(page, pageSize, caseType, status, lawyerID, startDate, endDate)
}

// Get 案件详情。
func (s *CaseService) Get(id uint64) (*model.Case, error) {
	return s.repo.FindByID(id)
}

// canFlow 案件状态机：filed->investigating->hearing->closed->archived，允许回退到上一步。
func canFlow(from, to string) bool {
	idx := map[string]int{constants.CaseStatusFiled: 0, constants.CaseStatusInvestigating: 1,
		constants.CaseStatusHearing: 2, constants.CaseStatusClosed: 3, constants.CaseStatusArchived: 4}
	a, okA := idx[from]
	b, okB := idx[to]
	if !okA || !okB {
		return false
	}
	return b == a+1 || b == a-1 || b == a
}

func jsonCoLawyers(ids []uint64) model.CoLawyerJSON {
	raw, _ := json.Marshal(ids)
	return model.CoLawyerJSON(raw)
}

func u64(v uint64) string {
	return fmt.Sprintf("%d", v)
}
