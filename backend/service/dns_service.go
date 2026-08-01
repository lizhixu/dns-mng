package service

import (
	"context"
	"dns-mng/models"
	"dns-mng/provider"
	"fmt"
	"sync"
	"time"
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

	// Build a set of DNSHE account IDs to filter out domains that delegate DNS to third parties
	accounts, _ := s.accountService.List(userID)
	accountMap := make(map[int64]string)
	dnsheAccountIDs := make(map[int64]bool)
	for _, acc := range accounts {
		accountMap[acc.ID] = acc.Name
		if acc.ProviderType == "dnshe" {
			dnsheAccountIDs[acc.ID] = true
		}
	}

	// 从缓存构建域名列表
	domains := make([]models.Domain, 0, len(caches))
	for _, cache := range caches {
		// DNSHE 账户下未使用 DNSHE 自身解析的域名不进入「所有域名」
		// （解析已托管到第三方平台，可能在其他服务商账户下重复出现）
		if dnsheAccountIDs[cache.AccountID] && !cache.UsesDNSHEDNS {
			continue
		}

		domain := models.Domain{
			ID:          cache.DomainID,
			Name:        cache.DomainName,
			AccountID:   cache.AccountID,
			RenewalDate: cache.RenewalDate,
			RenewalURL:  cache.RenewalURL,
			CacheSynced: true,
		}
		// 将 provider_updated_on 映射到 updated_on
		if cache.ProviderUpdatedOn != nil {
			domain.UpdatedOn = cache.ProviderUpdatedOn.Format("2006-01-02T15:04:05Z")
		}
		// DNSHE 账户下的域名回填解析归属标记
		if dnsheAccountIDs[cache.AccountID] {
			uses := cache.UsesDNSHEDNS
			domain.UsesDNSHEDNS = &uses
		}
		domains = append(domains, domain)
	}

	// 补充账户名称
	for i := range domains {
		if name, ok := accountMap[domains[i].AccountID]; ok {
			domains[i].AccountName = name
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
	for res := range results {
		if res.err == nil && res.domains != nil {
			allDomains = append(allDomains, res.domains...)
		}
		// Silently ignore errors from individual accounts
	}

	// Merge domain cache data and save provider's renewal date to cache
	if s.domainCacheService != nil {
		cacheMap, err := s.domainCacheService.BatchGetCacheByUser(userID)
		if err == nil {
			// Build a set of DNSHE account IDs for UsesDNSHEDNS backfill
			dnsheAccountIDs := make(map[int64]bool)
			for _, acc := range accounts {
				if acc.ProviderType == "dnshe" {
					dnsheAccountIDs[acc.ID] = true
				}
			}

			// 跨账户回填（前置）：DNSHE 域名解析至第三方供应商后，第三方账户下的同名域名
			// 没有过期时间/续费地址。这里先用 DNSHE 缓存中的 renewal_date + renewal_url 回填到内存域名，
			// 随后统一写缓存，避免被缓存里的旧空值盖住。必须在合并循环之前做，否则合并循环会把
			// 缓存里的空 renewal_date 保留下来，回填条件（RenewalDate==""）虽满足但顺序已晚。
			dnsheExpiryByName := dnsheExpiryFromCache(cacheMap, dnsheAccountIDs)
			backfillExpiryFromDNSHE(allDomains, dnsheAccountIDs, dnsheExpiryByName, cacheMap)

			for i := range allDomains {
				key := cacheKey(allDomains[i].AccountID, allDomains[i].ID)
				if cache, ok := cacheMap[key]; ok {
					if cache.RenewalManual {
						// 用户已手动维护到期时间/续费地址，同步时保留手动值，不被 provider 返回值覆盖
						allDomains[i].RenewalDate = cache.RenewalDate
						allDomains[i].RenewalURL = cache.RenewalURL
					} else {
						// Provider returned empty renewal date - preserve cached value
						if allDomains[i].RenewalDate == "" {
							allDomains[i].RenewalDate = cache.RenewalDate
						}
						// Provider returned empty renewal URL - preserve cached value
						if allDomains[i].RenewalURL == "" {
							allDomains[i].RenewalURL = cache.RenewalURL
						}
					}
					allDomains[i].CacheSynced = true
					// DNSHE 账户下的域名回填解析归属标记
					if dnsheAccountIDs[allDomains[i].AccountID] {
						uses := cache.UsesDNSHEDNS
						allDomains[i].UsesDNSHEDNS = &uses
					}
				}
				// Always save to cache (UpsertCache preserves existing renewal info when new values are empty;
				// RenewalManual 传 nil，同步不改动手动锁定状态)
				s.domainCacheService.UpsertCache(userID, allDomains[i].AccountID, allDomains[i].ID, allDomains[i].Name, &models.UpdateDomainCacheRequest{
					RenewalDate: allDomains[i].RenewalDate,
					RenewalURL:  allDomains[i].RenewalURL,
				})
			}
		}
	}

	// 过滤掉 DNSHE 第三方解析的域名（uses_dnshe_dns=false）：这些域名的解析已托管到第三方平台，
	// 会在第三方账户下重复出现，在「所有域名」列表中展示无意义，统一不展示。
	dnsheAccountIDs := dnsheAccountIDsFrom(accounts)
	filtered := make([]models.Domain, 0, len(allDomains))
	for _, d := range allDomains {
		if dnsheAccountIDs[d.AccountID] && d.UsesDNSHEDNS != nil && !*d.UsesDNSHEDNS {
			continue
		}
		filtered = append(filtered, d)
	}

	return filtered, nil
}
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

// dnsheAccountIDsFrom returns a set of DNSHE account IDs built from the given accounts.
// Reuses an already-loaded account list to avoid an extra DB query.
func dnsheAccountIDsFrom(accounts []models.Account) map[int64]bool {
	ids := make(map[int64]bool)
	for _, acc := range accounts {
		if acc.ProviderType == "dnshe" {
			ids[acc.ID] = true
		}
	}
	return ids
}

// dnsheExpiryFromCache builds name -> (renewal_date, renewal_url) from cached domains
// belonging to DNSHE accounts. Used to backfill third-party domains (e.g. Cloudflare) that
// share a name with a DNSHE domain resolved to them — those third-party entries have no
// expiry/renewal URL, so we copy them from the DNSHE domain.
func dnsheExpiryFromCache(cacheMap map[string]*models.DomainCache, dnsheAccountIDs map[int64]bool) map[string]models.Domain {
	byName := make(map[string]models.Domain)
	for _, cache := range cacheMap {
		if dnsheAccountIDs[cache.AccountID] && cache.DomainName != "" && (cache.RenewalDate != "" || cache.RenewalURL != "") {
			// 保留信息最全的那条（同名 DNSHE 域名理论上只有一条）
			existing, ok := byName[cache.DomainName]
			if !ok || (len(cache.RenewalDate) > len(existing.RenewalDate)) {
				byName[cache.DomainName] = models.Domain{
					Name:        cache.DomainName,
					RenewalDate: cache.RenewalDate,
					RenewalURL:  cache.RenewalURL,
				}
			}
		}
	}
	return byName
}

// backfillExpiryFromDNSHE copies renewal_date/renewal_url from the DNSHE expiry map onto
// non-DNSHE domains whose renewal_date and renewal_url are both empty. Callers must invoke
// this BEFORE the merge+upsert loop so the backfilled values are written to cache in one pass
// (otherwise the cache's stale/empty renewal_date would be preserved and the backfill skipped).
// 手动锁定（renewal_manual=1）的域名跳过回填，保留用户手动维护的值。
func backfillExpiryFromDNSHE(domains []models.Domain, dnsheAccountIDs map[int64]bool, dnsheExpiryByName map[string]models.Domain, cacheMap map[string]*models.DomainCache) {
	if len(dnsheExpiryByName) == 0 {
		return
	}
	for i := range domains {
		if dnsheAccountIDs[domains[i].AccountID] {
			continue
		}
		if domains[i].Name == "" || (domains[i].RenewalDate != "" && domains[i].RenewalURL != "") {
			continue
		}
		// 手动锁定的域名跳过回填，避免 DNSHE 过期信息覆盖用户手动维护的值
		if cache, ok := cacheMap[cacheKey(domains[i].AccountID, domains[i].ID)]; ok && cache.RenewalManual {
			continue
		}
		src, ok := dnsheExpiryByName[domains[i].Name]
		if !ok {
			continue
		}
		if domains[i].RenewalDate == "" {
			domains[i].RenewalDate = src.RenewalDate
		}
		if domains[i].RenewalURL == "" {
			domains[i].RenewalURL = src.RenewalURL
		}
	}
}

func (s *DNSService) ListDomains(ctx context.Context, userID, accountID int64) ([]models.Domain, error) {
	domains, err := s.ListDomainsFromCache(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	// 缓存为空时自动从服务商同步一次（首次访问新账户）
	if len(domains) == 0 {
		domains, _, err = s.ListDomainsFromProvider(ctx, userID, accountID)
		if err != nil {
			return nil, err
		}
	}
	return domains, nil
}

// ListDomainsFromCache returns domains from cache for a specific account.
// Returns empty list when cache is empty — does NOT call the provider.
// Use ListDomainsFromProvider (via refresh endpoint) to populate the cache.
func (s *DNSService) ListDomainsFromCache(ctx context.Context, userID, accountID int64) ([]models.Domain, error) {
	if s.domainCacheService == nil {
		return []models.Domain{}, nil
	}

	caches, err := s.domainCacheService.GetCacheByUser(userID)
	if err != nil {
		return nil, err
	}

	// Determine if this account is a DNSHE account for UsesDNSHEDNS backfill
	account, accErr := s.accountService.Get(userID, accountID)
	isDNSHE := accErr == nil && account.ProviderType == "dnshe"

	// 过滤出指定账户的域名
	domains := make([]models.Domain, 0)
	for _, cache := range caches {
		if cache.AccountID == accountID {
			domain := models.Domain{
				ID:          cache.DomainID,
				Name:        cache.DomainName,
				AccountID:   cache.AccountID,
				RenewalDate: cache.RenewalDate,
				RenewalURL:  cache.RenewalURL,
				CacheSynced: true,
			}
			// 将 provider_updated_on 映射到 updated_on
			if cache.ProviderUpdatedOn != nil {
				domain.UpdatedOn = cache.ProviderUpdatedOn.Format("2006-01-02T15:04:05Z")
			}
			if isDNSHE {
				uses := cache.UsesDNSHEDNS
				domain.UsesDNSHEDNS = &uses
			}
			domains = append(domains, domain)
		}
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

	// Set account info for all domains
	for i := range domains {
		domains[i].AccountID = accountID
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
				key := cacheKey(accountID, domain.ID)
				providerDomainIDs[key] = true
			}

			// Check which cached domains are not in provider response
			for key, cache := range cacheMap {
				if cache.AccountID == accountID && !providerDomainIDs[key] {
					domainsToDelete = append(domainsToDelete, cache.DomainID)
				}
			}

			// 跨账户回填（前置）：当前账户非 DNSHE 时，用 DNSHE 缓存中的 renewal_date + renewal_url
			// 回填到本账户同名域名（DNSHE 域名解析至第三方后，第三方域名无过期时间/续费地址）。
			// 先回填到内存域名，再走下面的合并+写缓存，避免缓存旧值盖住回填值。
			if account.ProviderType != "dnshe" {
				if dnsheAccounts, err := s.accountService.List(userID); err == nil {
					dnsheIDs := dnsheAccountIDsFrom(dnsheAccounts)
					if len(dnsheIDs) > 0 {
						dnsheExpiryByName := dnsheExpiryFromCache(cacheMap, dnsheIDs)
						backfillExpiryFromDNSHE(domains, dnsheIDs, dnsheExpiryByName, cacheMap)
					}
				}
			}

			// Merge domain cache data and save provider's renewal date to cache
			for i := range domains {
				key := cacheKey(domains[i].AccountID, domains[i].ID)
				if cache, ok := cacheMap[key]; ok {
					if cache.RenewalManual {
						// 用户已手动维护到期时间/续费地址，同步时保留手动值，不被 provider 返回值覆盖
						domains[i].RenewalDate = cache.RenewalDate
						domains[i].RenewalURL = cache.RenewalURL
					} else {
						// Provider returned empty renewal date - preserve cached value
						if domains[i].RenewalDate == "" {
							domains[i].RenewalDate = cache.RenewalDate
						}
						// Provider returned empty renewal URL - preserve cached value
						if domains[i].RenewalURL == "" {
							domains[i].RenewalURL = cache.RenewalURL
						}
					}
					domains[i].CacheSynced = true
					// 回填解析归属（DNSHE 账户下域名保留缓存中的 uses_dnshe_dns）
					if account.ProviderType == "dnshe" {
						uses := cache.UsesDNSHEDNS
						domains[i].UsesDNSHEDNS = &uses
					}
				}
				// Always save to cache (UpsertCache preserves existing renewal info when new values are empty;
				// RenewalManual 传 nil，同步不改动手动锁定状态)
				s.domainCacheService.UpsertCache(userID, domains[i].AccountID, domains[i].ID, domains[i].Name, &models.UpdateDomainCacheRequest{
					RenewalDate: domains[i].RenewalDate,
					RenewalURL:  domains[i].RenewalURL,
				})
			}
		}
	}

	return domains, domainsToDelete, nil
}

func (s *DNSService) GetDomain(ctx context.Context, userID, accountID int64, domainID string) (*models.Domain, error) {
	// 先获取账户判断是否 DNSHE
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, err
	}
	isDNSHE := account.ProviderType == "dnshe"

	// 优先从缓存读取
	if s.domainCacheService != nil {
		cache, err := s.domainCacheService.GetCache(userID, accountID, domainID)
		if err == nil && cache != nil {
			domain := &models.Domain{
				ID:          cache.DomainID,
				Name:        cache.DomainName,
				AccountID:   cache.AccountID,
				RenewalDate: cache.RenewalDate,
				RenewalURL:  cache.RenewalURL,
				CacheSynced: true,
			}
			if cache.ProviderUpdatedOn != nil {
				domain.UpdatedOn = cache.ProviderUpdatedOn.Format("2006-01-02T15:04:05Z")
			}
			// 补充账户名称
			domain.AccountName = account.Name
			if isDNSHE {
				uses := cache.UsesDNSHEDNS
				domain.UsesDNSHEDNS = &uses
			}
			return domain, nil
		}
	}

	// 缓存未命中，从供应商获取
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
			if isDNSHE {
				uses := cache.UsesDNSHEDNS
				domain.UsesDNSHEDNS = &uses
			}
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
		record.TTL = p.DefaultTTL()
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

	updatedRecord, err := p.UpdateRecord(ctx, account.APIKey, domainID, record)
	if err != nil {
		return nil, err
	}
	if updatedRecord != nil && updatedRecord.UpdatedOn == "" {
		updatedRecord.UpdatedOn = time.Now().Format(time.RFC3339)
	}

	return updatedRecord, nil
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

	// 手动编辑：用户提供了非空到期时间或续费地址时，锁定该字段，防止后续同步被 provider 覆盖；
	// 两者都为空时解锁（置 0），允许同步用 provider 数据回填。
	if req.RenewalManual == nil {
		flag := 0
		if req.RenewalDate != "" || req.RenewalURL != "" {
			flag = 1
		}
		req.RenewalManual = &flag
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

	// 与 UpdateDomainCache 一致：调用方未显式指定 renewal_manual 时，按是否提供非空
	// 到期时间/续费地址自动计算——非空锁定（1），都为空解锁（0），防止后续同步覆盖。
	for i := range items {
		if items[i].RenewalManual == nil {
			flag := 0
			if items[i].RenewalDate != "" || items[i].RenewalURL != "" {
				flag = 1
			}
			items[i].RenewalManual = &flag
		}
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
