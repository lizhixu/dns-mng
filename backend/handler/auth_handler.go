package handler

import (
	"net/http"

	"dns-mng/models"
	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *service.UserService
	logService  *service.LogService
}

func NewAuthHandler(userService *service.UserService, logService *service.LogService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		logService:  logService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log registration
	h.logService.CreateLog(resp.User.ID, "create", "user", req.Username, map[string]interface{}{
		"username": req.Username,
	}, c.ClientIP())

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userService.Login(&req)
	if err != nil {
		// Log failed login attempt
		h.logService.CreateLog(0, "login", "user", req.Username, map[string]interface{}{
			"username": req.Username,
			"success":  false,
			"reason":   err.Error(),
		}, c.ClientIP())
		
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Log successful login
	h.logService.CreateLog(resp.User.ID, "login", "user", req.Username, map[string]interface{}{
		"username": req.Username,
		"success":  true,
	}, c.ClientIP())

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdatePassword(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log password change
	user, _ := h.userService.GetUser(userID)
	h.logService.CreateLog(userID, "update", "password", user.Username, map[string]interface{}{
		"username": user.Username,
	}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
