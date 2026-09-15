package service

import (
	"log/slog"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"
)

// DocumentService 文档业务逻辑。
type DocumentService struct {
	repo     *repository.DocumentRepository
	caseRepo *repository.CaseRepository
	logger   *slog.Logger
}

// NewDocumentService 构造文档服务。
func NewDocumentService(repo *repository.DocumentRepository, caseRepo *repository.CaseRepository, logger *slog.Logger) *DocumentService {
	return &DocumentService{repo: repo, caseRepo: caseRepo, logger: logger}
}

// Create 上传文档。
func (s *DocumentService) Create(caseID, uploaderID uint64, title, fileType, fileURL string) (*model.Document, error) {
	if _, err := s.caseRepo.FindByID(caseID); err != nil {
		return nil, util.Wrap(err, "Document[case_id=%d] upload: case not found", caseID)
	}
	if !contains(constants.DocumentFileTypeValues, fileType) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Document[file_type="+fileType+"] upload: invalid type")
	}
	d := &model.Document{Title: title, FileType: fileType, FileURL: fileURL, UploadTime: time.Now(),
		CaseID: caseID, UploaderID: uploaderID}
	if err := s.repo.Create(d); err != nil {
		s.logger.Error(constants.LogDocumentUploadFailed, "error", err.Error())
		return nil, util.Wrap(err, "Document[case_id=%d] upload create failed", caseID)
	}
	s.logger.Info(constants.LogDocumentUploadSuccess, "document_id", d.ID)
	return d, nil
}

// ListByCase 按案件查看文档。
func (s *DocumentService) ListByCase(caseID uint64) ([]model.Document, error) {
	return s.repo.ListByCase(caseID)
}

// List 文档中心分页查询。
func (s *DocumentService) List(page, pageSize int, fileType, keyword string) ([]model.Document, int64, error) {
	return s.repo.List(page, pageSize, fileType, keyword)
}

// Delete 删除文档。
func (s *DocumentService) Delete(id uint64) error {
	if err := s.repo.Delete(id); err != nil {
		return util.Wrap(err, "Document[id=%d] delete failed", id)
	}
	s.logger.Info(constants.LogDocumentDeleteSuccess, "document_id", id)
	return nil
}
