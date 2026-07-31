package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"dns-mng/models"
	"dns-mng/provider"
	"dns-mng/provider/cloudflare"
	"dns-mng/provider/dnshe"
)

// DNSHEService provides DNSHE-specific operations (register/delete/renew/quota)
// that are not part of the generic DNSProvider interface.
type DNSHEService struct {
	accountService     *AccountService
	domainCacheService *DomainCacheService
	client             *dnshe.Client
}

func NewDNSHEService(accountService *AccountService, domainCacheService *DomainCacheService) *DNSHEService {
	return &DNSHEService{
		accountService:     accountService,
		domainCacheService: domainCacheService,
		client:             dnshe.NewClient(),
	}
}

// ListAccounts returns all DNSHE accounts for the user.
func (s *DNSHEService) ListAccounts(userID int64) ([]models.Account, error) {
	accounts, err := s.accountService.List(userID)
	if err != nil {
		return nil, err
	}
	var dnsheAccounts []models.Account
	for _, acc := range accounts {
		if acc.ProviderType == "dnshe" {
			dnsheAccounts = append(dnsheAccounts, acc)
		}
	}
	return dnsheAccounts, nil
}

// getAccountCredentials fetches the account and parses its API key/secret.
func (s *DNSHEService) getAccountCredentials(userID, accountID int64) (string, string, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return "", "", fmt.Errorf("account not found: %w", err)
	}
	if account.ProviderType != "dnshe" {
		return "", "", fmt.Errorf("account %d is not a DNSHE account", accountID)
	}
	return dnshe.ParseAPIKey(account.APIKey)
}

// GetQuota queries the quota for a DNSHE account.
func (s *DNSHEService) GetQuota(ctx context.Context, userID, accountID int64) (*dnshe.QuotaResponse, error) {
	key, secret, err := s.getAccountCredentials(userID, accountID)
	if err != nil {
		return nil, err
	}
	return s.client.GetQuota(ctx, key, secret)
}

// RegisterSubdomain registers a new subdomain under a DNSHE account.
func (s *DNSHEService) RegisterSubdomain(ctx context.Context, userID, accountID int64, subdomain, rootdomain string) (*dnshe.RegisterSubdomainResponse, error) {
	key, secret, err := s.getAccountCredentials(userID, accountID)
	if err != nil {
		return nil, err
	}
	return s.client.RegisterSubdomain(ctx, key, secret, subdomain, rootdomain)
}

// DeleteSubdomain deletes a subdomain from a DNSHE account and soft-deletes the
// corresponding domain cache entry so it no longer appears in lists.
func (s *DNSHEService) DeleteSubdomain(ctx context.Context, userID, accountID int64, subdomainID int) (*dnshe.DeleteSubdomainResponse, error) {
	key, secret, err := s.getAccountCredentials(userID, accountID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.DeleteSubdomain(ctx, key, secret, subdomainID)
	if err != nil {
		return nil, err
	}
	// Soft-delete the domain cache entry (if any) to avoid stale records.
	_ = s.domainCacheService.DeleteCache(userID, accountID, strconv.Itoa(subdomainID))
	return resp, nil
}

// RenewSubdomain renews a subdomain under a DNSHE account.
func (s *DNSHEService) RenewSubdomain(ctx context.Context, userID, accountID int64, subdomainID int) (*dnshe.RenewSubdomainResponse, error) {
	key, secret, err := s.getAccountCredentials(userID, accountID)
	if err != nil {
		return nil, err
	}
	return s.client.RenewSubdomain(ctx, key, secret, subdomainID)
}

// SetDomainResolution updates the uses_dnshe_dns flag on a domain cache entry.
func (s *DNSHEService) SetDomainResolution(userID, accountID int64, domainID string, usesDNSHEDNS bool) error {
	_, err := s.domainCacheService.UpsertCache(userID, accountID, domainID, "", &models.UpdateDomainCacheRequest{
		UsesDNSHEDNS: &usesDNSHEDNS,
	})
	return err
}

// ResolveToCloudflare switches a DNSHE domain's NS records to point to Cloudflare.
// It creates/reuses a Cloudflare zone for the domain's root, then writes the
// Cloudflare-assigned nameservers as NS records on the DNSHE side.
func (s *DNSHEService) ResolveToCloudflare(ctx context.Context, userID, dnsheAccountID int64, domainID string, cfAccountID int64) (*models.ResolveToCloudflareResult, error) {
	// 1. DNSHE credentials
	key, secret, err := s.getAccountCredentials(userID, dnsheAccountID)
	if err != nil {
		return nil, err
	}

	// 2. Get subdomain info (rootdomain)
	subdomainID, err := strconv.Atoi(domainID)
	if err != nil {
		return nil, fmt.Errorf("invalid domain id: %w", err)
	}
	subResp, err := s.client.GetSubdomain(ctx, key, secret, subdomainID)
	if err != nil {
		return nil, fmt.Errorf("get subdomain: %w", err)
	}
	domainName := subResp.Subdomain.FullDomain
	if domainName == "" {
		domainName = subResp.Subdomain.Subdomain + "." + subResp.Subdomain.RootDomain
	}
	if domainName == "" {
		return nil, fmt.Errorf("could not determine domain name for domain id %s", domainID)
	}

	// 3. Cloudflare account
	cfAccount, err := s.accountService.Get(userID, cfAccountID)
	if err != nil {
		return nil, fmt.Errorf("cloudflare account not found: %w", err)
	}
	if cfAccount.ProviderType != "cloudflare" {
		return nil, fmt.Errorf("account %d is not a Cloudflare account", cfAccountID)
	}
	cfClient := cloudflare.NewClient()
	apiToken := cfAccount.APIKey

	// 4. Get Cloudflare account ID, then find or create zone
	// 使用完整域名作为 zone（DNSHE 的 rootdomain 是共享 TLD，不能直接建 zone）
	cfAccountIDStr, err := cfClient.GetAccountID(ctx, apiToken)
	if err != nil {
		return nil, fmt.Errorf("get cloudflare account id: %w", err)
	}

	// Try to find existing zone by name
	zone, err := cfClient.GetZoneByName(ctx, apiToken, domainName)
	if err != nil {
		// Zone not found, create a new one
		zone, err = cfClient.CreateZone(ctx, apiToken, cfAccountIDStr, domainName)
		if err != nil {
			return nil, fmt.Errorf("create cloudflare zone: %w", err)
		}
	}

	if len(zone.NameServers) == 0 {
		return nil, fmt.Errorf("cloudflare zone %s has no nameservers assigned", domainName)
	}

	// 5. Create new NS records on DNSHE side FIRST (safer than deleting first)
	for _, ns := range zone.NameServers {
		if err := s.client.CreateDNSRecord(ctx, key, secret, subdomainID, "", "NS", ns, 86400, nil); err != nil {
			return nil, fmt.Errorf("create NS record %s on DNSHE: %w", ns, err)
		}
	}

	// 6. Delete old NS records (that are not the ones we just created)
	listResp, err := s.client.ListDNSRecords(ctx, key, secret, subdomainID)
	if err != nil {
		return nil, fmt.Errorf("list DNSHE dns records: %w", err)
	}
	newNSSet := make(map[string]bool)
	for _, ns := range zone.NameServers {
		newNSSet[strings.ToLower(ns)] = true
	}
	for _, r := range listResp.Records {
		if r.Type == "NS" && !newNSSet[strings.ToLower(r.Content)] {
			if err := s.client.DeleteDNSRecord(ctx, key, secret, r.ID); err != nil {
				// non-fatal: old record may linger, but new NS are already in place
			}
		}
	}

	// 7. Mark domain as third-party resolution
	usesDNSHE := false
	_, _ = s.domainCacheService.UpsertCache(userID, dnsheAccountID, domainID, "", &models.UpdateDomainCacheRequest{
		UsesDNSHEDNS: &usesDNSHE,
	})

	return &models.ResolveToCloudflareResult{
		DomainName:  domainName,
		ZoneName:    zone.Name,
		NameServers: zone.NameServers,
	}, nil
}

// domainListForAutoRenew lists domains for a DNSHE account via the registered provider.
func (s *DNSHEService) domainListForAutoRenew(ctx context.Context, account models.Account) ([]models.Domain, error) {
	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, err
	}
	return p.ListDomains(ctx, account.APIKey)
}