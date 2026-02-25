package tencentcloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

// parseAPIKey parses the API key in format "SecretId,SecretKey"
func (c *Client) parseAPIKey(apiKey string) (string, string, error) {
	parts := strings.SplitN(apiKey, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid API key format, expected 'SecretId,SecretKey'")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func (c *Client) newDNSPodClient(apiKey string) (*dnspod.Client, error) {
	secretId, secretKey, err := c.parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	return dnspod.NewClient(credential, "", cpf)
}

func (c *Client) ListDomains(ctx context.Context, apiKey string) ([]*dnspod.DomainListItem, error) {
	client, err := c.newDNSPodClient(apiKey)
	if err != nil {
		return nil, err
	}

	request := dnspod.NewDescribeDomainListRequest()
	response, err := client.DescribeDomainList(request)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}

	return response.Response.DomainList, nil
}

func (c *Client) GetDomain(ctx context.Context, apiKey string, domain string) (*dnspod.DomainInfo, error) {
	client, err := c.newDNSPodClient(apiKey)
	if err != nil {
		return nil, err
	}

	request := dnspod.NewDescribeDomainRequest()
	request.Domain = common.StringPtr(domain)
	response, err := client.DescribeDomain(request)
	if err != nil {
		return nil, fmt.Errorf("get domain: %w", err)
	}

	return response.Response.DomainInfo, nil
}

func (c *Client) ListRecords(ctx context.Context, apiKey string, domain string) ([]*dnspod.RecordListItem, error) {
	client, err := c.newDNSPodClient(apiKey)
	if err != nil {
		return nil, err
	}

	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)
	response, err := client.DescribeRecordList(request)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}

	return response.Response.RecordList, nil
}

func (c *Client) CreateRecord(ctx context.Context, apiKey string, domain string, recordType string, name string, value string, ttl uint64, mx uint64) (*dnspod.CreateRecordResponse, error) {
	client, err := c.newDNSPodClient(apiKey)
	if err != nil {
		return nil, err
	}

	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.RecordType = common.StringPtr(recordType)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.TTL = common.Uint64Ptr(ttl)
	
	if name != "" && name != "@" {
		request.SubDomain = common.StringPtr(name)
	} else {
		request.SubDomain = common.StringPtr("@")
	}
	
	if recordType == "MX" && mx > 0 {
		request.MX = common.Uint64Ptr(mx)
	}

	response, err := client.CreateRecord(request)
	if err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}

	return response, nil
}

func (c *Client) UpdateRecord(ctx context.Context, apiKey string, domain string, recordID uint64, recordType string, name string, value string, ttl uint64, mx uint64) error {
	client, err := c.newDNSPodClient(apiKey)
	if err != nil {
		return err
	}

	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.RecordId = common.Uint64Ptr(recordID)
	request.RecordType = common.StringPtr(recordType)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.TTL = common.Uint64Ptr(ttl)
	
	if name != "" && name != "@" {
		request.SubDomain = common.StringPtr(name)
	} else {
		request.SubDomain = common.StringPtr("@")
	}
	
	if recordType == "MX" && mx > 0 {
		request.MX = common.Uint64Ptr(mx)
	}

	_, err = client.ModifyRecord(request)
	if err != nil {
		return fmt.Errorf("update record: %w", err)
	}

	return nil
}

func (c *Client) DeleteRecord(ctx context.Context, apiKey string, domain string, recordID uint64) error {
	client, err := c.newDNSPodClient(apiKey)
	if err != nil {
		return err
	}

	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.RecordId = common.Uint64Ptr(recordID)

	_, err = client.DeleteRecord(request)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	return nil
}
