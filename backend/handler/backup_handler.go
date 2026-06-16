package handler

import (
	"fmt"
	"net/http"
	"time"

	"dns-mng/middleware"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct {
	backupService *service.BackupService
}

func NewBackupHandler(backupService *service.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// Export 导出用户配置为 JSON 文件（可选加密）。
// GET /api/backup/export?password=xxx
func (h *BackupHandler) Export(c *gin.Context) {
	userID := middleware.GetUserID(c)
	password := c.Query("password")

	data, err := h.backupService.Export(userID, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("dns-mng-backup-%s.json", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/json", data)
}

// Import 从备份文件还原配置。
// POST /api/backup/import  body: {"password":"","overwrite":false,"content":"...base64..."}
// content 为上传的备份文件原始内容（JSON 字符串）。
func (h *BackupHandler) Import(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Password  string `json:"password"`
		Overwrite bool   `json:"overwrite"`
		Content   string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "备份文件内容不能为空"})
		return
	}

	result, err := h.backupService.Import(userID, []byte(req.Content), req.Password, req.Overwrite)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
