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

	// 极致优化延迟：跳过 ListRecords 检查，直接盲插记录
	// ACME TXT 值每次请求通常都是唯一的。省去一次完整的 API 查询，达到最低延迟。
	ttl := req.TTL
	if ttl < 0 {
		ttl = 0 // 传承到底层以使用 provider.DefaultTTL
	}

	state := true
	_, err = s.dns.CreateRecord(ctx, userID, match.accountID, match.domainID, &models.CreateRecordRequest{
		NodeName:   match.nodeName,
		RecordType: "TXT",
		Content:    req.Value,
		TTL:        ttl,
		State:      &state,
	})
	if err != nil {
		// 如果盲插失败（通常是因为并发重试导致 API 报“记录已存在”或者其他原因），
		// 降级查询真实记录，如果记录其实已经成功存在了，就忽略插入错误。
		records, listErr := s.dns.ListRecords(ctx, userID, match.accountID, match.domainID)
		if listErr == nil {
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
		}
		// 如果并没有存在，统一向上抛出原始插入失败的 error
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
