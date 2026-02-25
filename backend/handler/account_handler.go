package handler

import (
	"fmt"
	"net/http"

	"dns-mng/middleware"
	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountService *service.AccountService
	logService     *service.LogService
}

func NewAccountHandler(accountService *service.AccountService, logService *service.LogService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logService:     logService,
	}
}

func (h *AccountHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accounts, err := h.accountService.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []models.Account{}
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *AccountHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.accountService.Create(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	h.logService.CreateLog(userID, "create", "account", fmt.Sprintf("%d", account.ID), map[string]interface{}{
		"name":     account.Name,
		"provider": account.ProviderType,
	}, c.ClientIP())

	c.JSON(http.StatusCreated, account)
}

func (h *AccountHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	// Get old account for comparison
	accounts, _ := h.accountService.List(userID)
	var oldAccount *models.Account
	for _, acc := range accounts {
		if acc.ID == accountID {
			oldAccount = &acc
			break
		}
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.accountService.Update(userID, accountID, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Log operation with changes
	logDetails := map[string]interface{}{
		"name":     account.Name,
		"provider": account.ProviderType,
	}
	if oldAccount != nil {
		changes := make(map[string]interface{})
		if oldAccount.Name != account.Name {
			changes["name"] = map[string]string{"old": oldAccount.Name, "new": account.Name}
		}
		if len(changes) > 0 {
			logDetails["changes"] = changes
		}
	}
	h.logService.CreateLog(userID, "update", "account", fmt.Sprintf("%d", account.ID), logDetails, c.ClientIP())

	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accountID, err := middleware.GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	// Get account details before deletion
	accounts, _ := h.accountService.List(userID)
	var accountDetails map[string]interface{}
	for _, acc := range accounts {
		if acc.ID == accountID {
			accountDetails = map[string]interface{}{
				"name":     acc.Name,
				"provider": acc.ProviderType,
			}
			break
		}
	}

	if err := h.accountService.Delete(userID, accountID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	if accountDetails == nil {
		accountDetails = map[string]interface{}{}
	}
	h.logService.CreateLog(userID, "delete", "account", fmt.Sprintf("%d", accountID), accountDetails, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
