package service

import (
	"context"
	"dns-mng/models"
	"dns-mng/provider"
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

	var allDomains []models.Domain
	for _, account := range accounts {
		p, err := provider.Get(account.ProviderType)
		if err != nil {
			continue
		}
		domains, err := p.ListDomains(ctx, account.APIKey)
		if err != nil {
			continue
		}
		for i := range domains {
			domains[i].AccountID = account.ID
			domains[i].AccountName = account.Name
		}
		allDomains = append(allDomains, domains...)
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
