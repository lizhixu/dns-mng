package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// IPGeoInfo holds the geolocation information for an IP address
type IPGeoInfo struct {
	CountryName     string `json:"country_name"`
	CountryCode     string `json:"country_code"`
	CityName        string `json:"city_name"`
	RegionName      string `json:"region_name"`
	Continent       string `json:"continent"`
	ASN             string `json:"asn"`
	ASNOrganization string `json:"asn_organization"`
	IsProxy         bool   `json:"is_proxy"`
}

type ipGeoService struct {
	client *http.Client
}

var IPGeoSvc = &ipGeoService{
	client: &http.Client{Timeout: 5 * time.Second},
}

// IPLookup is a convenience function that delegates to the global IPGeoSvc.
func IPLookup(ip string) *IPGeoInfo {
	return IPGeoSvc.Lookup(ip)
}

// Lookup queries freeipapi.com for the given IP address and returns geo info.
// Returns nil if the IP is private/loopback or the lookup fails.
func (s *ipGeoService) Lookup(ip string) *IPGeoInfo {
	if ip == "" {
		return nil
	}

	// Skip private/loopback IPs
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsPrivate() || parsed.IsLoopback() {
		return nil
	}

	url := fmt.Sprintf("https://free.freeipapi.com/api/json/%s", ip)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil
	}

	// Parse the freeipapi.com response
	var raw struct {
		CountryName     string `json:"countryName"`
		CountryCode     string `json:"countryCode"`
		CityName        string `json:"cityName"`
		RegionName      string `json:"regionName"`
		Continent       string `json:"continent"`
		ASN             string `json:"asn"`
		ASNOrganization string `json:"asnOrganization"`
		IsProxy         bool   `json:"isProxy"`
		PhoneCodes      []any  `json:"phoneCodes"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}

	info := &IPGeoInfo{
		CountryName:     raw.CountryName,
		CountryCode:     raw.CountryCode,
		CityName:        raw.CityName,
		RegionName:      raw.RegionName,
		Continent:       raw.Continent,
		ASN:             raw.ASN,
		ASNOrganization: raw.ASNOrganization,
		IsProxy:         raw.IsProxy,
	}

	return info
}

// FormatLocation returns a human-readable location string from IPGeoInfo.
// Example: "Sydney, New South Wales, Australia (AS13335 Cloudflare, Inc.)"
func FormatLocation(info *IPGeoInfo) string {
	if info == nil {
		return ""
	}

	var parts []string
	if info.CityName != "" {
		parts = append(parts, info.CityName)
	}
	if info.RegionName != "" && info.RegionName != info.CityName {
		parts = append(parts, info.RegionName)
	}
	if info.CountryName != "" {
		parts = append(parts, info.CountryName)
	}

	loc := strings.Join(parts, ", ")

	var extras []string
	if info.ASN != "" {
		org := info.ASNOrganization
		if org != "" {
			extras = append(extras, fmt.Sprintf("AS%s %s", info.ASN, org))
		} else {
			extras = append(extras, fmt.Sprintf("AS%s", info.ASN))
		}
	}
	if info.IsProxy {
		extras = append(extras, "Proxy/VPN")
	}

	if len(extras) > 0 {
		loc += " (" + strings.Join(extras, ", ") + ")"
	}

	return loc
}
