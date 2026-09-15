package service

import (
	"log/slog"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"
)

// ClientService 客户业务逻辑。
type ClientService struct {
	repo     *repository.ClientRepository
	caseRepo *repository.CaseRepository
	logger   *slog.Logger
}

// NewClientService 构造客户服务。
func NewClientService(repo *repository.ClientRepository, caseRepo *repository.CaseRepository, logger *slog.Logger) *ClientService {
	return &ClientService{repo: repo, caseRepo: caseRepo, logger: logger}
}

// Create 新建客户。
func (s *ClientService) Create(name, idNumber, contact, address, remark string) (*model.Client, error) {
	c := &model.Client{Name: name, IDNumber: idNumber, Contact: contact, Address: address, Remark: remark}
	if err := s.repo.Create(c); err != nil {
		return nil, util.Wrap(err, "Client[name=%s] create failed", name)
	}
	s.logger.Info(constants.LogClientCreateSuccess, "client_id", c.ID)
	return c, nil
}

// Update 编辑客户。
func (s *ClientService) Update(id uint64, name, idNumber, contact, address, remark string) (*model.Client, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Client[id=%d] update find failed", id)
	}
	if name != "" {
		c.Name = name
	}
	if idNumber != "" {
		c.IDNumber = idNumber
	}
	if contact != "" {
		c.Contact = contact
	}
	if address != "" {
		c.Address = address
	}
	if remark != "" {
		c.Remark = remark
	}
	if err := s.repo.Update(c); err != nil {
		return nil, util.Wrap(err, "Client[id=%d] update save failed", id)
	}
	s.logger.Info(constants.LogClientUpdateSuccess, "client_id", c.ID)
	return c, nil
}

// Delete 删除客户。
func (s *ClientService) Delete(id uint64) error {
	if err := s.repo.Delete(id); err != nil {
		return util.Wrap(err, "Client[id=%d] delete failed", id)
	}
	s.logger.Info(constants.LogClientDeleteSuccess, "client_id", id)
	return nil
}

// List 分页检索客户。
func (s *ClientService) List(page, pageSize int, keyword string) ([]model.Client, int64, error) {
	return s.repo.List(page, pageSize, keyword)
}

// GetWithCases 客户详情 + 历史案件。
func (s *ClientService) GetWithCases(id uint64) (*model.Client, []model.Case, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, nil, util.Wrap(err, "Client[id=%d] get failed", id)
	}
	cases, err := s.caseRepo.ListByClient(id)
	if err != nil {
		return nil, nil, err
	}
	return c, cases, nil
}
