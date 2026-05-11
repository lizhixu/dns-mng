package aliyun

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

const (
	aliyunRegion        = "cn-hangzhou"
	defaultRecordLine   = "default"
	domainRecordsPageSz = 500
	domainsPageSz       = 100
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) parseAPIKey(apiKey string) (accessKeyID, accessKeySecret string, err error) {
	parts := strings.SplitN(apiKey, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid API key format, expected 'AccessKeyId,AccessKeySecret'")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func (c *Client) newDNSClient(apiKey string) (*alidns.Client, error) {
	ak, sk, err := c.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}
	return alidns.NewClientWithAccessKey(aliyunRegion, ak, sk)
}

func rrToNodeName(rr string) string {
	rr = strings.TrimSpace(rr)
	if rr == "" || rr == "@" {
		return ""
	}
	return rr
}

func nodeNameToRR(node string) string {
	node = strings.TrimSpace(node)
	if node == "" || node == "@" {
		return "@"
	}
	return node
}

func recordLineFromRaw(raw map[string]interface{}) string {
	if raw == nil {
		return defaultRecordLine
	}
	v, ok := raw["line"].(string)
	if !ok || strings.TrimSpace(v) == "" {
		return defaultRecordLine
	}
	return v
}

// ListDomains returns all domains (paginated).
func (c *Client) ListDomains(ctx context.Context, apiKey string) ([]alidns.DomainInDescribeDomains, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	var out []alidns.DomainInDescribeDomains
	page := int64(1)
	for {
		req := alidns.CreateDescribeDomainsRequest()
		req.PageNumber = requests.NewInteger64(page)
		req.PageSize = requests.NewInteger64(domainsPageSz)
		resp, err := client.DescribeDomains(req)
		if err != nil {
			return nil, fmt.Errorf("list domains: %w", err)
		}
		out = append(out, resp.Domains.Domain...)
		if int64(len(resp.Domains.Domain)) < domainsPageSz ||
			page*domainsPageSz >= resp.TotalCount {
			break
		}
		page++
	}
	return out, nil
}

// AllAliDnsDelegation returns true when DescribeDomainNs reports that every
// nameserver at the registrar is Alibaba Cloud DNS (解析仅在阿里云生效).
func (c *Client) AllAliDnsDelegation(ctx context.Context, apiKey, domainName string) (bool, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return false, err
	}
	req := alidns.CreateDescribeDomainNsRequest()
	req.DomainName = domainName
	resp, err := client.DescribeDomainNs(req)
	if err != nil {
		return false, fmt.Errorf("describe domain ns: %w", err)
	}
	return resp.AllAliDns, nil
}

// DescribeDomain wraps DescribeDomainInfo.
func (c *Client) DescribeDomain(ctx context.Context, apiKey, domainName string) (*alidns.DescribeDomainInfoResponse, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	req := alidns.CreateDescribeDomainInfoRequest()
	req.DomainName = domainName
	resp, err := client.DescribeDomainInfo(req)
	if err != nil {
		return nil, fmt.Errorf("describe domain: %w", err)
	}
	return resp, nil
}

// ListAllRecords returns all DNS records for a domain (paginated).
func (c *Client) ListAllRecords(ctx context.Context, apiKey, domainName string) ([]alidns.Record, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	var out []alidns.Record
	page := int64(1)
	for {
		req := alidns.CreateDescribeDomainRecordsRequest()
		req.DomainName = domainName
		req.PageNumber = requests.NewInteger64(page)
		req.PageSize = requests.NewInteger64(domainRecordsPageSz)
		resp, err := client.DescribeDomainRecords(req)
		if err != nil {
			return nil, fmt.Errorf("list records: %w", err)
		}
		out = append(out, resp.DomainRecords.Record...)
		if int64(len(resp.DomainRecords.Record)) < domainRecordsPageSz ||
			page*domainRecordsPageSz >= resp.TotalCount {
			break
		}
		page++
	}
	return out, nil
}

func (c *Client) AddDomainRecord(ctx context.Context, apiKey, domainName, rr, line, recordType, value string, ttl int64, priority int64) (string, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return "", err
	}
	req := alidns.CreateAddDomainRecordRequest()
	req.DomainName = domainName
	req.RR = rr
	req.Type = recordType
	req.Value = value
	req.Line = line
	req.TTL = requests.NewInteger64(ttl)
	if strings.EqualFold(recordType, "MX") && priority > 0 {
		req.Priority = requests.NewInteger64(priority)
	}
	resp, err := client.AddDomainRecord(req)
	if err != nil {
		return "", fmt.Errorf("add record: %w", err)
	}
	if resp.RecordId == "" {
		return "", fmt.Errorf("add record: empty RecordId")
	}
	return resp.RecordId, nil
}

func (c *Client) UpdateDomainRecord(ctx context.Context, apiKey, recordID, rr, line, recordType, value string, ttl int64, priority int64) error {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return err
	}
	req := alidns.CreateUpdateDomainRecordRequest()
	req.RecordId = recordID
	req.RR = rr
	req.Type = recordType
	req.Value = value
	req.Line = line
	req.TTL = requests.NewInteger64(ttl)
	if strings.EqualFold(recordType, "MX") && priority > 0 {
		req.Priority = requests.NewInteger64(priority)
	}
	_, err = client.UpdateDomainRecord(req)
	if err != nil {
		return fmt.Errorf("update record: %w", err)
	}
	return nil
}

func (c *Client) SetRecordStatus(ctx context.Context, apiKey, recordID, status string) error {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return err
	}
	req := alidns.CreateSetDomainRecordStatusRequest()
	req.RecordId = recordID
	req.Status = status
	_, err = client.SetDomainRecordStatus(req)
	if err != nil {
		return fmt.Errorf("set record status: %w", err)
	}
	return nil
}

func (c *Client) DeleteDomainRecord(ctx context.Context, apiKey, recordID string) error {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return err
	}
	req := alidns.CreateDeleteDomainRecordRequest()
	req.RecordId = recordID
	_, err = client.DeleteDomainRecord(req)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}
	return nil
}
