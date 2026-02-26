package handler

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DNSCheckHandler struct{}

func NewDNSCheckHandler() *DNSCheckHandler {
	return &DNSCheckHandler{}
}

type DNSCheckRequest struct {
	Domain     string `json:"domain" binding:"required"`
	RecordType string `json:"record_type" binding:"required"`
	Expected   string `json:"expected"`
}

type DNSCheckResponse struct {
	Domain     string   `json:"domain"`
	RecordType string   `json:"record_type"`
	Values     []string `json:"values"`
	Expected   string   `json:"expected,omitempty"`
	Matched    bool     `json:"matched"`
	Message    string   `json:"message"`
	Timestamp  string   `json:"timestamp"`
	DNSServer  string   `json:"dns_server,omitempty"`
}

// Public DNS servers to use
var publicDNSServers = []string{
	"8.8.8.8:53",         // Google DNS
	"1.1.1.1:53",         // Cloudflare DNS
	"208.67.222.222:53",  // OpenDNS
}

// CheckDNS checks if DNS record has propagated
func (h *DNSCheckHandler) CheckDNS(c *gin.Context) {
	var req DNSCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize domain and record type
	domain := strings.TrimSpace(req.Domain)
	recordType := strings.ToUpper(strings.TrimSpace(req.RecordType))
	expected := strings.TrimSpace(req.Expected)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var values []string
	var err error
	var usedDNS string

	// Try multiple DNS servers to avoid local DNS pollution
	for _, dnsServer := range publicDNSServers {
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				return d.DialContext(ctx, network, dnsServer)
			},
		}

		switch recordType {
		case "A":
			values, err = lookupAWithResolver(ctx, resolver, domain)
		case "AAAA":
			values, err = lookupAAAAWithResolver(ctx, resolver, domain)
		case "CNAME":
			values, err = lookupCNAMEWithResolver(ctx, resolver, domain)
		case "MX":
			values, err = lookupMXWithResolver(ctx, resolver, domain)
		case "TXT":
			values, err = lookupTXTWithResolver(ctx, resolver, domain)
		case "NS":
			values, err = lookupNSWithResolver(ctx, resolver, domain)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported record type: " + recordType})
			return
		}

		// If successful, use this result
		if err == nil && len(values) > 0 {
			usedDNS = dnsServer
			break
		}
	}

	response := DNSCheckResponse{
		Domain:     domain,
		RecordType: recordType,
		Values:     values,
		Expected:   expected,
		Timestamp:  time.Now().Format(time.RFC3339),
		DNSServer:  usedDNS,
	}

	if err != nil {
		response.Matched = false
		response.Message = "DNS query failed: " + err.Error()
		c.JSON(http.StatusOK, response)
		return
	}

	if len(values) == 0 {
		response.Matched = false
		response.Message = "No DNS records found"
		c.JSON(http.StatusOK, response)
		return
	}

	// Check if expected value matches
	if expected != "" {
		matched := false
		normalizedExpected := normalizeValue(expected)

		for _, value := range values {
			normalizedValue := normalizeValue(value)
			if normalizedValue == normalizedExpected {
				matched = true
				break
			}
		}

		response.Matched = matched
		if matched {
			response.Message = "DNS record matches expected value"
		} else {
			response.Message = "DNS record does not match expected value"
		}
	} else {
		response.Matched = true
		response.Message = "DNS record found"
	}

	c.JSON(http.StatusOK, response)
}

// normalizeValue normalizes DNS values for comparison
func normalizeValue(value string) string {
	// Trim spaces
	value = strings.TrimSpace(value)
	// Convert to lowercase
	value = strings.ToLower(value)
	// Remove trailing dot (common in DNS responses)
	value = strings.TrimSuffix(value, ".")
	return value
}

// lookupAWithResolver queries A records with custom resolver
func lookupAWithResolver(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	ips, err := resolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, ip := range ips {
		results = append(results, ip.String())
	}
	return results, nil
}

// lookupAAAAWithResolver queries AAAA records with custom resolver
func lookupAAAAWithResolver(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	ips, err := resolver.LookupIP(ctx, "ip6", domain)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, ip := range ips {
		results = append(results, ip.String())
	}
	return results, nil
}

// lookupCNAMEWithResolver queries CNAME records with custom resolver
func lookupCNAMEWithResolver(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	cname, err := resolver.LookupCNAME(ctx, domain)
	if err != nil {
		return nil, err
	}
	return []string{strings.TrimSuffix(cname, ".")}, nil
}

// lookupMXWithResolver queries MX records with custom resolver
func lookupMXWithResolver(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	mxs, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, mx := range mxs {
		results = append(results, strings.TrimSuffix(mx.Host, "."))
	}
	return results, nil
}

// lookupTXTWithResolver queries TXT records with custom resolver
func lookupTXTWithResolver(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	return resolver.LookupTXT(ctx, domain)
}

// lookupNSWithResolver queries NS records with custom resolver
func lookupNSWithResolver(ctx context.Context, resolver *net.Resolver, domain string) ([]string, error) {
	nss, err := resolver.LookupNS(ctx, domain)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, ns := range nss {
		results = append(results, strings.TrimSuffix(ns.Host, "."))
	}
	return results, nil
}
