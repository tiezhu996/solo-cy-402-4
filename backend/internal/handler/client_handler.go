package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"cylawcase/internal/constants"
	"cylawcase/internal/dto"
	"cylawcase/internal/service"
	"cylawcase/internal/util"

	"github.com/gin-gonic/gin"
)

// ClientHandler 客户 HTTP 处理器。
type ClientHandler struct {
	svc    *service.ClientService
	logger *slog.Logger
}

// NewClientHandler 构造客户处理器。
func NewClientHandler(svc *service.ClientService, logger *slog.Logger) *ClientHandler {
	return &ClientHandler{svc: svc, logger: logger}
}

// List 客户列表。
func (h *ClientHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	list, total, err := h.svc.List(q.Page, q.PageSize, c.Query("keyword"))
	if err != nil {
		h.wrapError(c, err, "Client list failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Get 客户详情 + 历史案件。
func (h *ClientHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Client[id] get: invalid id")
		return
	}
	client, cases, err := h.svc.GetWithCases(id)
	if err != nil {
		h.wrapError(c, err, "Client get failed")
		return
	}
	OK(c, gin.H{"client": client, "cases": cases})
}

// Create 新建客户。
func (h *ClientHandler) Create(c *gin.Context) {
	var req dto.ClientCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Client create: "+err.Error())
		return
	}
	cl, err := h.svc.Create(req.Name, req.IDNumber, req.Contact, req.Address, req.Remark)
	if err != nil {
		h.wrapError(c, err, "Client[name="+req.Name+"] create failed")
		return
	}
	OKWithMessage(c, constants.MsgClientCreated, cl)
}

// Update 编辑客户。
func (h *ClientHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Client[id] update: invalid id")
		return
	}
	var req dto.ClientUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Client[id="+strconv.FormatUint(id, 10)+"] update: "+err.Error())
		return
	}
	cl, err := h.svc.Update(id, req.Name, req.IDNumber, req.Contact, req.Address, req.Remark)
	if err != nil {
		h.wrapError(c, err, "Client update failed")
		return
	}
	OK(c, cl)
}

// Delete 删除客户。
func (h *ClientHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Client[id] delete: invalid id")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		h.wrapError(c, err, "Client delete failed")
		return
	}
	OK(c, nil)
}

func (h *ClientHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		c.Set("audit_detail", appErr.Message)
		h.logger.Warn("client handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("client handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
