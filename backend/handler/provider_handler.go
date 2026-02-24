package handler

import (
	"net/http"

	"dns-mng/provider"

	"github.com/gin-gonic/gin"
)

type ProviderHandler struct{}

func NewProviderHandler() *ProviderHandler {
	return &ProviderHandler{}
}

func (h *ProviderHandler) List(c *gin.Context) {
	providers := provider.List()
	c.JSON(http.StatusOK, providers)
}
