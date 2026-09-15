package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"cylawcase/internal/constants"
	"cylawcase/internal/dto"
	"cylawcase/internal/middleware"
	"cylawcase/internal/service"
	"cylawcase/internal/util"

	"github.com/gin-gonic/gin"
)

// DocumentHandler 文档 HTTP 处理器。
type DocumentHandler struct {
	svc    *service.DocumentService
	logger *slog.Logger
}

// NewDocumentHandler 构造文档处理器。
func NewDocumentHandler(svc *service.DocumentService, logger *slog.Logger) *DocumentHandler {
	return &DocumentHandler{svc: svc, logger: logger}
}

// Create 上传文档。
func (h *DocumentHandler) Create(c *gin.Context) {
	var req dto.DocumentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Document create: "+err.Error())
		return
	}
	d, err := h.svc.Create(req.CaseID, middleware.GetUserID(c), req.Title, req.FileType, req.FileURL)
	if err != nil {
		h.wrapError(c, err, "Document[case_id="+strconv.FormatUint(req.CaseID, 10)+"] create failed")
		return
	}
	OKWithMessage(c, constants.MsgDocumentUploaded, d)
}

// ListByCase 按案件查看文档。
func (h *DocumentHandler) ListByCase(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Document list: invalid case id")
		return
	}
	list, err := h.svc.ListByCase(caseID)
	if err != nil {
		h.wrapError(c, err, "Document list by case failed")
		return
	}
	OK(c, list)
}

// List 文档中心分页查询。
func (h *DocumentHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	list, total, err := h.svc.List(q.Page, q.PageSize, c.Query("file_type"), c.Query("keyword"))
	if err != nil {
		h.wrapError(c, err, "Document list failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Delete 删除文档。
func (h *DocumentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Document[id] delete: invalid id")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		h.wrapError(c, err, "Document delete failed")
		return
	}
	OK(c, nil)
}

func (h *DocumentHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		c.Set("audit_detail", appErr.Message)
		h.logger.Warn("document handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("document handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
