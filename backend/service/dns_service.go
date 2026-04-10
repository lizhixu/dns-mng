package service

import (
	"context"
	"dns-mng/models"
	"dns-mng/provider"
	"fmt"
	"sync"
)

type DNSService struct {
	accountService     *AccountService
	domainCacheService *DomainCacheService
}

func NewDNSService(accountService *AccountService, domainCacheService *DomainCacheService) *DNSService {
	return &DNSService{
		accountService:     accountService,
		domainCacheService: domainCacheService,
	}
}

func (s *DNSService) ListAllDomains(ctx context.Context, userID int64) ([]models.Domain, error) {
	// 优先从缓存读取
	return s.ListAllDomainsFromCache(ctx, userID)
}

// ListAllDomainsFromCache returns domains from cache
func (s *DNSService) ListAllDomainsFromCache(ctx context.Context, userID int64) ([]models.Domain, error) {
	if s.domainCacheService == nil {
		return s.ListAllDomainsFromProvider(ctx, userID)
	}

	caches, err := s.domainCacheService.GetCacheByUser(userID)
	if err != nil || len(caches) == 0 {
		// 如果缓存为空，从服务商获取
		return s.ListAllDomainsFromProvider(ctx, userID)
	}

	// 从缓存构建域名列表
	domains := make([]models.Domain, 0, len(caches))
	for _, cache := range caches {
		domains = append(domains, models.Domain{
			ID:          cache.DomainID,
			Name:        cache.DomainName,
			AccountID:   cache.AccountID,
			RenewalDate: cache.RenewalDate,
			RenewalURL:  cache.RenewalURL,
			CacheSynced: true,
		})
	}

	// 补充账户名称
	accounts, err := s.accountService.List(userID)
	if err == nil {
		accountMap := make(map[int64]string)
		for _, acc := range accounts {
			accountMap[acc.ID] = acc.Name
		}
		for i := range domains {
			if name, ok := accountMap[domains[i].AccountID]; ok {
				domains[i].AccountName = name
			}
		}
	}

	return domains, nil
}

// ListAllDomainsFromProvider fetches domains from DNS providers and updates cache
func (s *DNSService) ListAllDomainsFromProvider(ctx context.Context, userID int64) ([]models.Domain, error) {
	accounts, err := s.accountService.List(userID)
	if err != nil {
		return nil, err
	}

	// Use goroutines to fetch domains concurrently
	type result struct {
		domains         []models.Domain
		domainsToDelete []string
		err             error
	}

	results := make(chan result, len(accounts))
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(acc models.Account) {
			defer wg.Done()

			domains, domainsToDelete, err := s.listDomainsFromProviderForAccount(ctx, userID, acc)
			if err != nil {
				results <- result{err: err}
				return
			}

			results <- result{domains: domains, domainsToDelete: domainsToDelete}
		}(account)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var allDomains []models.Domain
	var allDomainsToDelete []string
	for res := range results {
		if res.err == nil && res.domains != nil {
			allDomains = append(allDomains, res.domains...)
			allDomainsToDelete = append(allDomainsToDelete, res.domainsToDelete...)
		}
		// Silently ignore errors from individual accounts
	}

	// Soft delete domains that no longer exist
	// This will be handled by the handler which has account context
	_ = allDomainsToDelete

	// Merge domain cache data and update cache
	if s.domainCacheService != nil {
		cacheMap, err := s.domainCacheService.BatchGetCacheByUser(userID)
		if err == nil {
			for i := range allDomains {
				key := cacheKey(allDomains[i].AccountID, allDomains[i].ID)
				if cache, ok := cacheMap[key]; ok {
					allDomains[i].RenewalDate = cache.RenewalDate
					allDomains[i].RenewalURL = cache.RenewalURL
					allDomains[i].CacheSynced = true
				} else {
					// 为新域名创建缓存条目
					s.domainCacheService.UpsertCache(userID, allDomains[i].AccountID, allDomains[i].ID, allDomains[i].Name, &models.UpdateDomainCacheRequest{
						RenewalDate: "",
						RenewalURL:  "",
					})
				}
			}
		}
	}

	return allDomains, nil
}

// listDomainsFromProviderForAccount is a helper to fetch domains for a single account
func (s *DNSService) listDomainsFromProviderForAccount(ctx context.Context, userID int64, account models.Account) ([]models.Domain, []string, error) {
	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, nil, err
	}

	domains, err := p.ListDomains(ctx, account.APIKey)
	if err != nil {
		return nil, nil, err
	}

	// Add account info to domains
	for i := range domains {
		domains[i].AccountID = account.ID
		domains[i].AccountName = account.Name
	}

	// Get current cache to compare
	var domainsToDelete []string
	if s.domainCacheService != nil {
		cacheMap, err := s.domainCacheService.BatchGetCacheByUser(userID)
		if err == nil {
			// Build a set of domain IDs from provider
			providerDomainIDs := make(map[string]bool)
			for _, domain := range domains {
				key := cacheKey(account.ID, domain.ID)
				providerDomainIDs[key] = true
			}

			// Check which cached domains are not in provider response
			for key, cache := range cacheMap {
				if cache.AccountID == account.ID && !providerDomainIDs[key] {
					domainsToDelete = append(domainsToDelete, cache.DomainID)
				}
			}
		}
	}

	return domains, domainsToDelete, nil
}

// cacheKey generates a map key for domain cache lookup
func cacheKey(accountID int64, domainID string) string {
	return fmt.Sprintf("%d:%s", accountID, domainID)
}

func (s *DNSService) ListDomains(ctx context.Context, userID, accountID int64) ([]models.Domain, error) {
	// 优先从缓存读取
	return s.ListDomainsFromCache(ctx, userID, accountID)
}

// ListDomainsFromCache returns domains from cache for a specific account
func (s *DNSService) ListDomainsFromCache(ctx context.Context, userID, accountID int64) ([]models.Domain, error) {
	if s.domainCacheService == nil {
		domains, _, err := s.ListDomainsFromProvider(ctx, userID, accountID)
		return domains, err
	}

	caches, err := s.domainCacheService.GetCacheByUser(userID)
	if err != nil || len(caches) == 0 {
		// 如果缓存为空，从服务商获取
		domains, _, err := s.ListDomainsFromProvider(ctx, userID, accountID)
		return domains, err
	}

	// 过滤出指定账户的域名
	domains := make([]models.Domain, 0)
	for _, cache := range caches {
		if cache.AccountID == accountID {
			domains = append(domains, models.Domain{
				ID:          cache.DomainID,
				Name:        cache.DomainName,
				AccountID:   cache.AccountID,
				RenewalDate: cache.RenewalDate,
				RenewalURL:  cache.RenewalURL,
				CacheSynced: true,
			})
		}
	}

	if len(domains) == 0 {
		// 如果该账户没有缓存，从服务商获取
		domains, _, err := s.ListDomainsFromProvider(ctx, userID, accountID)
		return domains, err
	}

	return domains, nil
}

// ListDomainsFromProvider fetches domains from DNS provider and updates cache
// Returns domains and a list of domain IDs that exist in cache but not in provider (should be soft deleted)
func (s *DNSService) ListDomainsFromProvider(ctx context.Context, userID, accountID int64) ([]models.Domain, []string, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, nil, err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, nil, err
	}

	domains, err := p.ListDomains(ctx, account.APIKey)
	if err != nil {
		return nil, nil, err
	}

	// Get current cache to compare
	var domainsToDelete []string
	if s.domainCacheService != nil {
		cacheMap, err := s.domainCacheService.BatchGetCacheByUser(userID)
		if err == nil {
			// Build a set of domain IDs from provider
			providerDomainIDs := make(map[string]bool)
			for _, domain := range domains {
				key := cacheKey(accountID, domain.ID)
				providerDomainIDs[key] = true
			}

			// Check which cached domains are not in provider response
			for key, cache := range cacheMap {
				if cache.AccountID == accountID && !providerDomainIDs[key] {
					domainsToDelete = append(domainsToDelete, cache.DomainID)
				}
			}

			// Merge domain cache data and update cache
			for i := range domains {
				key := cacheKey(domains[i].AccountID, domains[i].ID)
				if cache, ok := cacheMap[key]; ok {
					domains[i].RenewalDate = cache.RenewalDate
					domains[i].RenewalURL = cache.RenewalURL
					domains[i].CacheSynced = true
				} else {
					// 为新域名创建缓存条目
					s.domainCacheService.UpsertCache(userID, accountID, domains[i].ID, domains[i].Name, &models.UpdateDomainCacheRequest{
						RenewalDate: "",
						RenewalURL:  "",
					})
				}
			}
		}
	}

	return domains, domainsToDelete, nil
}

func (s *DNSService) GetDomain(ctx context.Context, userID, accountID int64, domainID string) (*models.Domain, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, err
	}

	domain, err := p.GetDomain(ctx, account.APIKey, domainID)
	if err != nil {
		return nil, err
	}

	// Merge domain cache data
	if s.domainCacheService != nil && domain != nil {
		cache, err := s.domainCacheService.GetCache(userID, accountID, domainID)
		if err == nil && cache != nil {
			domain.RenewalDate = cache.RenewalDate
			domain.RenewalURL = cache.RenewalURL
			domain.CacheSynced = true
		}
	}

	return domain, nil
}

func (s *DNSService) ListRecords(ctx context.Context, userID, accountID int64, domainID string) ([]models.Record, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, err
	}

	return p.ListRecords(ctx, account.APIKey, domainID)
}

func (s *DNSService) CreateRecord(ctx context.Context, userID, accountID int64, domainID string, req *models.CreateRecordRequest) (*models.Record, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, err
	}

	state := true
	if req.State != nil {
		state = *req.State
	}

	record := &models.Record{
		NodeName:   req.NodeName,
		RecordType: req.RecordType,
		TTL:        req.TTL,
		State:      state,
		Content:    req.Content,
		Priority:   req.Priority,
	}

	if record.TTL == 0 {
		record.TTL = 300
	}

	return p.CreateRecord(ctx, account.APIKey, domainID, record)
}

func (s *DNSService) UpdateRecord(ctx context.Context, userID, accountID int64, domainID, recordID string, req *models.UpdateRecordRequest) (*models.Record, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, err
	}

	state := true
	if req.State != nil {
		state = *req.State
	}

	record := &models.Record{
		ID:         recordID,
		NodeName:   req.NodeName,
		RecordType: req.RecordType,
		TTL:        req.TTL,
		State:      state,
		Content:    req.Content,
		Priority:   req.Priority,
	}

	return p.UpdateRecord(ctx, account.APIKey, domainID, record)
}

func (s *DNSService) DeleteRecord(ctx context.Context, userID, accountID int64, domainID, recordID string) error {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return err
	}

	return p.DeleteRecord(ctx, account.APIKey, domainID, recordID)
}

// UpdateDomainCache updates the renewal info for a domain
func (s *DNSService) UpdateDomainCache(ctx context.Context, userID, accountID int64, domainID, domainName string, req *models.UpdateDomainCacheRequest) (*models.Domain, error) {
	if s.domainCacheService == nil {
		return nil, fmt.Errorf("domain cache service not available")
	}

	_, err := s.domainCacheService.UpsertCache(userID, accountID, domainID, domainName, req)
	if err != nil {
		return nil, err
	}

	return s.GetDomain(ctx, userID, accountID, domainID)
}

// BatchUpdateDomainCache updates multiple domain cache entries
func (s *DNSService) BatchUpdateDomainCache(ctx context.Context, userID int64, items []models.BatchCacheItem) error {
	if s.domainCacheService == nil {
		return fmt.Errorf("domain cache service not available")
	}

	return s.domainCacheService.BatchUpsertCache(userID, items)
}

// BatchDeleteDomainCache deletes multiple domain cache entries
func (s *DNSService) BatchDeleteDomainCache(ctx context.Context, userID int64, items []models.BatchCacheDeleteItem) error {
	if s.domainCacheService == nil {
		return fmt.Errorf("domain cache service not available")
	}

	return s.domainCacheService.BatchDeleteCache(userID, items)
}

// GetCacheStats returns statistics about cached domains
func (s *DNSService) GetCacheStats(ctx context.Context, userID int64) (*models.CacheStats, error) {
	if s.domainCacheService == nil {
		return nil, fmt.Errorf("domain cache service not available")
	}

	return s.domainCacheService.GetCacheStats(userID)
}

// BatchSoftDeleteDomains soft deletes multiple domains
func (s *DNSService) BatchSoftDeleteDomains(ctx context.Context, userID int64, items []models.BatchCacheDeleteItem) error {
	if s.domainCacheService == nil {
		return fmt.Errorf("domain cache service not available")
	}

	return s.domainCacheService.BatchDeleteCache(userID, items)
}

// BatchRestoreDomains restores multiple soft deleted domains
func (s *DNSService) BatchRestoreDomains(ctx context.Context, userID int64, items []models.BatchCacheDeleteItem) error {
	if s.domainCacheService == nil {
		return fmt.Errorf("domain cache service not available")
	}

	return s.domainCacheService.BatchRestoreCache(userID, items)
}
