package handler

import (
	"net/http"

	"cylawcase/internal/constants"

	"github.com/gin-gonic/gin"
)

// OK 统一成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": constants.MsgOK, "data": data})
}

// OKWithMessage 成功响应并携带自定义文案。
func OKWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": message, "data": data})
}

// Fail 输出错误响应。
func Fail(c *gin.Context, status int, code int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message, "data": nil})
}

// pageResponse 分页响应结构。
func pageResponse(list any, total int64, page, pageSize int) gin.H {
	return gin.H{"list": list, "total": total, "page": page, "page_size": pageSize}
}

// appErrorStatus 错误码 -> HTTP 状态码。
func appErrorStatus(code int) int {
	switch code {
	case constants.CodeUnauthorized, constants.CodeInvalidCredentials:
		return http.StatusUnauthorized
	case constants.CodeForbidden:
		return http.StatusForbidden
	case constants.CodeNotFound:
		return http.StatusNotFound
	case constants.CodeConflict, constants.CodeCaseStatusConflict, constants.CodeBillingStatusConflict:
		return http.StatusConflict
	case constants.CodeValidationFailed:
		return http.StatusUnprocessableEntity
	case constants.CodeTooManyRequests:
		return http.StatusTooManyRequests
	case constants.CodeUploadTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusBadRequest
	}
}
