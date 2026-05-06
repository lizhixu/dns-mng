package middleware

import (
	"bytes"
	"dns-mng/models"
	"dns-mng/service"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body      *bytes.Buffer
	remaining int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.remaining > 0 {
		toWrite := len(b)
		if toWrite > w.remaining {
			toWrite = w.remaining
		}
		w.body.Write(b[:toWrite])
		w.remaining -= toWrite
	}
	return w.ResponseWriter.Write(b)
}

const maxLoggedBodyBytes = 64 * 1024

func shouldSkipAPILogging(path string) bool {
	if path == "/health" || path == "/ping" {
		return true
	}

	// Log query endpoints are intentionally excluded. Logging their large
	// responses creates recursive growth and increases SQLite lock pressure.
	return strings.HasPrefix(path, "/api/api-logs") || strings.HasPrefix(path, "/api/scheduler-logs")
}

// APILogger middleware records complete API call information
func APILogger(logService *service.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for certain paths if needed
		if shouldSkipAPILogging(c.Request.URL.Path) {
			c.Next()
			return
		}

		startTime := time.Now()

		// Read and store request body
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// Restore the body for handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Capture request headers
		headersMap := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				// Mask sensitive headers
				if key == "Authorization" || key == "Cookie" {
					headersMap[key] = "[MASKED]"
				} else {
					headersMap[key] = values[0]
				}
			}
		}
		requestHeaders, _ := json.Marshal(headersMap)

		// Wrap response writer to capture response body
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
			remaining:      maxLoggedBodyBytes,
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Get user ID (0 if not authenticated)
		userID, exists := c.Get("user_id")
		var uid int64
		if exists {
			uid = userID.(int64)
		}

		// Capture error message if any
		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		// Create API call log
		apiLog := &models.APICallLog{
			UserID:         uid,
			Method:         c.Request.Method,
			Path:           c.Request.URL.Path,
			Query:          c.Request.URL.RawQuery,
			RequestHeaders: string(requestHeaders),
			RequestBody:    requestBody,
			StatusCode:     c.Writer.Status(),
			ResponseBody:   blw.body.String(),
			IPAddress:      c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			DurationMs:     int(duration.Milliseconds()),
			ErrorMessage:   errorMessage,
		}

		// Save log asynchronously to avoid blocking response
		go func() {
			if err := logService.CreateAPICallLog(apiLog); err != nil {
				log.Printf("Failed to create API call log path=%s method=%s user_id=%d: %v", apiLog.Path, apiLog.Method, apiLog.UserID, err)
			}
		}()
	}
}
