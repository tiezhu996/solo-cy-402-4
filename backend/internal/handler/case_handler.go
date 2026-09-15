package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/dto"
	"cylawcase/internal/middleware"
	"cylawcase/internal/service"
	"cylawcase/internal/util"

	"github.com/gin-gonic/gin"
)

// CaseHandler 案件 HTTP 处理器。
type CaseHandler struct {
	svc    *service.CaseService
	logger *slog.Logger
}

// NewCaseHandler 构造案件处理器。
func NewCaseHandler(svc *service.CaseService, logger *slog.Logger) *CaseHandler {
	return &CaseHandler{svc: svc, logger: logger}
}

// List 案件列表。
func (h *CaseHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	lawyerID, _ := strconv.ParseUint(c.Query("lead_lawyer_id"), 10, 64)
	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if s := c.Query("end_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			endDate = &t
		}
	}
	list, total, err := h.svc.List(q.Page, q.PageSize, c.Query("case_type"), c.Query("status"), lawyerID, startDate, endDate)
	if err != nil {
		h.wrapError(c, err, "Case list failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Get 案件详情。
func (h *CaseHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id] get: invalid id")
		return
	}
	cs, err := h.svc.Get(id)
	if err != nil {
		h.wrapError(c, err, "Case get failed")
		return
	}
	OK(c, cs)
}

// Create 创建案件。
func (h *CaseHandler) Create(c *gin.Context) {
	var req dto.CaseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case create: "+err.Error())
		return
	}
	acceptDate, err := dto.ParseAcceptDate(valueOrEmpty(req.AcceptDate))
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case create: invalid accept_date")
		return
	}
	cs, err := h.svc.Create(req.ClientID, req.LeadLawyerID, req.Title, req.CaseType, req.Summary, acceptDate, req.CoLawyerIDs)
	if err != nil {
		h.wrapError(c, err, "Case[title="+req.Title+"] create failed")
		return
	}
	OKWithMessage(c, constants.MsgCaseCreated, cs)
}

// Update 更新案件。
func (h *CaseHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id] update: invalid id")
		return
	}
	var req dto.CaseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id="+strconv.FormatUint(id, 10)+"] update: "+err.Error())
		return
	}
	cs, err := h.svc.Update(id, req.Title, req.Summary, req.CoLawyerIDs)
	if err != nil {
		h.wrapError(c, err, "Case update failed")
		return
	}
	OK(c, cs)
}

// ChangeStatus 状态流转。
func (h *CaseHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id] status: invalid id")
		return
	}
	var req dto.CaseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id="+strconv.FormatUint(id, 10)+"] status: "+err.Error())
		return
	}
	cs, err := h.svc.ChangeStatus(id, middleware.GetUserRole(c), req.Status)
	if err != nil {
		h.wrapError(c, err, "Case status change failed")
		return
	}
	OK(c, cs)
}

// Assign 分配律师。
func (h *CaseHandler) Assign(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id] assign: invalid id")
		return
	}
	var req dto.CaseAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Case[id="+strconv.FormatUint(id, 10)+"] assign: "+err.Error())
		return
	}
	cs, err := h.svc.Assign(id, req.LeadLawyerID, req.CoLawyerIDs)
	if err != nil {
		h.wrapError(c, err, "Case assign failed")
		return
	}
	OK(c, cs)
}

func (h *CaseHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		c.Set("audit_detail", appErr.Message)
		h.logger.Warn("case handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("case handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
