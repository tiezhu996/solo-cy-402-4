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

// BillingHandler 账单 HTTP 处理器。
type BillingHandler struct {
	svc    *service.BillingService
	logger *slog.Logger
}

// NewBillingHandler 构造账单处理器。
func NewBillingHandler(svc *service.BillingService, logger *slog.Logger) *BillingHandler {
	return &BillingHandler{svc: svc, logger: logger}
}

// List 账单列表。
func (h *BillingHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	caseID, _ := strconv.ParseUint(c.Query("case_id"), 10, 64)
	clientID, _ := strconv.ParseUint(c.Query("client_id"), 10, 64)
	list, total, err := h.svc.List(q.Page, q.PageSize, caseID, clientID, c.Query("status"))
	if err != nil {
		h.wrapError(c, err, "Billing list failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// ListByCase 某案件账单。
func (h *BillingHandler) ListByCase(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Billing list by case: invalid case id")
		return
	}
	list, err := h.svc.ListByCase(caseID)
	if err != nil {
		h.wrapError(c, err, "Billing list by case failed")
		return
	}
	OK(c, list)
}

// Create 创建账单。
func (h *BillingHandler) Create(c *gin.Context) {
	var req dto.BillingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Billing create: "+err.Error())
		return
	}
	b, err := h.svc.Create(req.CaseID, req.ClientID, req.BillingType, req.Amount, req.InvoiceInfo)
	if err != nil {
		h.wrapError(c, err, "Billing[case_id="+strconv.FormatUint(req.CaseID, 10)+"] create failed")
		return
	}
	OKWithMessage(c, constants.MsgBillingCreated, b)
}

// MarkPaid 标记支付。
func (h *BillingHandler) MarkPaid(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Billing[id] paid: invalid id")
		return
	}
	b, err := h.svc.MarkPaid(id)
	if err != nil {
		h.wrapError(c, err, "Billing paid failed")
		return
	}
	OKWithMessage(c, constants.MsgBillingPaid, b)
}

// MarkInvoiced 开票。
func (h *BillingHandler) MarkInvoiced(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Billing[id] invoiced: invalid id")
		return
	}
	var req dto.BillingInvoiceRequest
	_ = c.ShouldBindJSON(&req)
	b, err := h.svc.MarkInvoiced(id, req.InvoiceInfo)
	if err != nil {
		h.wrapError(c, err, "Billing invoiced failed")
		return
	}
	OKWithMessage(c, constants.MsgBillingInvoiced, b)
}

// Void 作废账单。
func (h *BillingHandler) Void(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Billing[id] void: invalid id")
		return
	}
	b, err := h.svc.Void(id)
	if err != nil {
		h.wrapError(c, err, "Billing void failed")
		return
	}
	OKWithMessage(c, constants.MsgBillingVoided, b)
}

// Summary 本月汇总。
func (h *BillingHandler) Summary(c *gin.Context) {
	sum, err := h.svc.Summary()
	if err != nil {
		h.wrapError(c, err, "Billing summary failed")
		return
	}
	OK(c, sum)
}

func (h *BillingHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		c.Set("audit_detail", appErr.Message)
		h.logger.Warn("billing handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("billing handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
