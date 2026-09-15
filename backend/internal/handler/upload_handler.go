package handler

import (
	"log/slog"
	"net/http"

	"cylawcase/internal/config"
	"cylawcase/internal/constants"
	"cylawcase/internal/util"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器。
type UploadHandler struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewUploadHandler 构造上传处理器。
func NewUploadHandler(cfg *config.Config, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{cfg: cfg, logger: logger}
}

// UploadFile 上传文件（文档/头像）。
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Upload file: missing file")
		return
	}
	url, err := util.SaveUploadedFile(h.cfg.UploadDir, h.cfg.UploadMaxMB, file)
	if err != nil {
		h.logger.Error(constants.LogUploadFileFailed, "error", err.Error())
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Upload file failed: "+err.Error())
		return
	}
	h.logger.Info(constants.LogUploadFileSuccess, "url", url)
	OK(c, gin.H{"url": url})
}
