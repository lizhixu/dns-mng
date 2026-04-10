package middleware

import (
	"net/http"

	"dns-mng/service"

	"github.com/gin-gonic/gin"
)

// BasicAuthMiddleware authenticates using HTTP Basic Auth against system users.
// It sets "user_id" in gin context on success.
func BasicAuthMiddleware(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok || username == "" || password == "" {
			c.Header("WWW-Authenticate", `Basic realm="dns-mng"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "basic auth required"})
			c.Abort()
			return
		}

		user, err := userService.VerifyCredentials(username, password)
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="dns-mng"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}

