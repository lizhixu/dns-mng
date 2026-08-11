package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// realIPHeaders lists headers that carry the originating client's address,
// in priority order. Cloudflare's CF-Connecting-IP is authoritative and
// single-valued, so it is checked first. X-Real-IP and X-Forwarded-For are
// generic fallbacks for other reverse proxies / CDNs.
var realIPHeaders = []string{
	"CF-Connecting-IP",
	"X-Real-IP",
	"X-Forwarded-For",
}

// firstHeaderValue returns the first comma-separated entry of a header,
// trimmed of surrounding whitespace. Empty values are ignored.
func firstHeaderValue(value string) string {
	if value == "" {
		return ""
	}
	// X-Forwarded-For may chain: "client, proxy1, proxy2". The leftmost
	// entry is the original client.
	if i := strings.IndexByte(value, ','); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}

// RealIP recovers the originating client IP when the service sits behind a
// reverse proxy or CDN (Cloudflare, Nginx, etc.). It reads the proxy-supplied
// header, validates that it holds a single IP, and overwrites
// c.Request.RemoteAddr so downstream handlers and c.ClientIP() report the
// real client instead of the proxy/CDN node.
//
// This trusts the headers at face value. Make sure the service is only
// reachable through the intended proxy/CDN (i.e. the proxy/CDN strips or
// overwrites these headers before forwarding). With Cloudflare, the
// CF-Connecting-IP header is set by the edge and cannot be spoofed by the
// client, so it is the safest source.
func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, header := range realIPHeaders {
			value := c.Request.Header.Get(header)
			if value == "" {
				continue
			}

			ip := firstHeaderValue(value)
			if ip == "" {
				continue
			}

			// Reject obviously malformed values to avoid corrupting
			// RemoteAddr with garbage a client might inject into a
			// non-CF header.
			if net.ParseIP(ip) == nil {
				continue
			}

			// Normalize to host:port form so c.ClientIP() and RemoteAddr
			// consumers keep working unchanged.
			if strings.Contains(ip, ":") {
				// IPv6, keep brackets for RemoteAddr formatting.
				c.Request.RemoteAddr = "[" + ip + "]:0"
			} else {
				c.Request.RemoteAddr = ip + ":0"
			}
			break
		}

		c.Next()
	}
}