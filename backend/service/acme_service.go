package service

import (
	"context"
	"errors"
	"strings"

	"dns-mng/models"
)

type AcmeService struct {
	dns *DNSService
}

func NewAcmeService(dns *DNSService) *AcmeService {
	return &AcmeService{dns: dns}
}

type acmeDomainMatch struct {
	accountID  int64
	domainID   string
	domainName string
	nodeName   string
}

func normalizeFQDN(fqdn string) string {
	s := strings.TrimSpace(strings.ToLower(fqdn))
	s = strings.TrimSuffix(s, ".")
	return s
}

func (s *AcmeService) matchDomain(ctx context.Context, userID int64, fqdn string) (*acmeDomainMatch, error) {
	fqdn = normalizeFQDN(fqdn)
	if fqdn == "" || !strings.Contains(fqdn, ".") {
		return nil, errors.New("invalid fqdn")
	}

	// Prefer cache, but will fallback to provider fetch if cache empty.
	domains, err := s.dns.ListAllDomainsFromCache(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(domains) == 0 {
		return nil, errors.New("no domains available for user")
	}

	var best *acmeDomainMatch
	bestLen := -1

	for _, d := range domains {
		root := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(d.Name), "."))
		if root == "" {
			continue
		}

		if fqdn == root || strings.HasSuffix(fqdn, "."+root) {
			if len(root) > bestLen {
				relative := ""
				if fqdn != root {
					relative = strings.TrimSuffix(fqdn, "."+root)
					relative = strings.TrimSuffix(relative, ".")
				}
				best = &acmeDomainMatch{
					accountID:  d.AccountID,
					domainID:   d.ID,
					domainName: root,
					nodeName:   relative,
				}
				bestLen = len(root)
			}
		}
	}

	if best == nil {
		return nil, errors.New("no matching domain found for fqdn")
	}
	return best, nil
}

func (s *AcmeService) Present(ctx context.Context, userID int64, req *models.AcmeDNS01Request) (*models.AcmeDNS01Response, error) {
	match, err := s.matchDomain(ctx, userID, req.FQDN)
	if err != nil {
		return nil, err
	}

	records, err := s.dns.ListRecords(ctx, userID, match.accountID, match.domainID)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if strings.EqualFold(r.RecordType, "TXT") &&
			strings.EqualFold(r.NodeName, match.nodeName) &&
			r.Content == req.Value {
			return &models.AcmeDNS01Response{
				Status:   "ok",
				Domain:   match.domainName,
				NodeName: match.nodeName,
			}, nil
		}
	}

	ttl := req.TTL
	if ttl <= 0 {
		ttl = 300
	}

	state := true
	_, err = s.dns.CreateRecord(ctx, userID, match.accountID, match.domainID, &models.CreateRecordRequest{
		NodeName:    match.nodeName,
		RecordType:  "TXT",
		Content:     req.Value,
		TTL:         ttl,
		State:       &state,
	})
	if err != nil {
		return nil, err
	}

	return &models.AcmeDNS01Response{
		Status:   "ok",
		Domain:   match.domainName,
		NodeName: match.nodeName,
	}, nil
}

func (s *AcmeService) Cleanup(ctx context.Context, userID int64, req *models.AcmeDNS01Request) (*models.AcmeDNS01Response, error) {
	match, err := s.matchDomain(ctx, userID, req.FQDN)
	if err != nil {
		return nil, err
	}

	records, err := s.dns.ListRecords(ctx, userID, match.accountID, match.domainID)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if strings.EqualFold(r.RecordType, "TXT") &&
			strings.EqualFold(r.NodeName, match.nodeName) &&
			r.Content == req.Value {
			_ = s.dns.DeleteRecord(ctx, userID, match.accountID, match.domainID, r.ID)
		}
	}

	return &models.AcmeDNS01Response{
		Status:   "ok",
		Domain:   match.domainName,
		NodeName: match.nodeName,
	}, nil
}

