package service

import (
	"context"
	"dns-mng/models"
	"dns-mng/provider"
	"sync"
)

type DNSService struct {
	accountService *AccountService
}

func NewDNSService(accountService *AccountService) *DNSService {
	return &DNSService{accountService: accountService}
}

func (s *DNSService) ListAllDomains(ctx context.Context, userID int64) ([]models.Domain, error) {
	accounts, err := s.accountService.List(userID)
	if err != nil {
		return nil, err
	}

	// Use goroutines to fetch domains concurrently
	type result struct {
		domains []models.Domain
		err     error
	}

	results := make(chan result, len(accounts))
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(acc models.Account) {
			defer wg.Done()

			p, err := provider.Get(acc.ProviderType)
			if err != nil {
				results <- result{err: err}
				return
			}

			domains, err := p.ListDomains(ctx, acc.APIKey)
			if err != nil {
				results <- result{err: err}
				return
			}

			// Add account info to domains
			for i := range domains {
				domains[i].AccountID = acc.ID
				domains[i].AccountName = acc.Name
			}

			results <- result{domains: domains}
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

	return allDomains, nil
}

func (s *DNSService) ListDomains(ctx context.Context, userID, accountID int64) ([]models.Domain, error) {
	account, err := s.accountService.Get(userID, accountID)
	if err != nil {
		return nil, err
	}

	p, err := provider.Get(account.ProviderType)
	if err != nil {
		return nil, err
	}

	return p.ListDomains(ctx, account.APIKey)
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

	return p.GetDomain(ctx, account.APIKey, domainID)
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
