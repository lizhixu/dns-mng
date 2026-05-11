package huaweicloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwdns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	hwdnsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
)

const (
	defaultRegionID = "cn-north-4"
	pageLimit       = int32(500)
)

type apiCredentials struct {
	AK        string
	SK        string
	ProjectID string
	RegionID  string
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func looksLikeRegionID(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	prefixes := []string{
		"cn-", "ap-", "af-", "eu-", "sa-", "la-", "na-", "me-", "ru-", "tr-", "ae-", "my-", "us-",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func parseAPIKey(apiKey string) (apiCredentials, error) {
	parts := strings.Split(apiKey, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	switch len(parts) {
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return apiCredentials{}, fmt.Errorf("invalid API key: empty AK or SK")
		}
		return apiCredentials{AK: parts[0], SK: parts[1], RegionID: defaultRegionID}, nil
	case 3:
		if parts[0] == "" || parts[1] == "" {
			return apiCredentials{}, fmt.Errorf("invalid API key: empty AK or SK")
		}
		if looksLikeRegionID(parts[2]) {
			return apiCredentials{AK: parts[0], SK: parts[1], RegionID: parts[2]}, nil
		}
		return apiCredentials{AK: parts[0], SK: parts[1], ProjectID: parts[2], RegionID: defaultRegionID}, nil
	case 4:
		if parts[0] == "" || parts[1] == "" {
			return apiCredentials{}, fmt.Errorf("invalid API key: empty AK or SK")
		}
		return apiCredentials{AK: parts[0], SK: parts[1], ProjectID: parts[2], RegionID: parts[3]}, nil
	default:
		return apiCredentials{}, fmt.Errorf("invalid API key format: expected 'AK,SK' or 'AK,SK,ProjectId' or 'AK,SK,RegionId' or 'AK,SK,ProjectId,RegionId'")
	}
}

func (c *Client) newDNSClient(apiKey string) (*hwdns.DnsClient, error) {
	cr, err := parseAPIKey(apiKey)
	if err != nil {
		return nil, err
	}
	reg, err := hwdnsregion.SafeValueOf(cr.RegionID)
	if err != nil {
		return nil, fmt.Errorf("dns region: %w", err)
	}
	b := basic.NewCredentialsBuilder().WithAk(cr.AK).WithSk(cr.SK)
	if cr.ProjectID != "" {
		b = b.WithProjectId(cr.ProjectID)
	}
	cred, err := b.SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("huawei credentials: %w", err)
	}
	hc, err := hwdns.DnsClientBuilder().
		WithRegion(reg).
		WithCredential(cred).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("dns client: %w", err)
	}
	return hwdns.NewDnsClient(hc), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func trimRootZone(name, zoneName string) string {
	n := strings.TrimSuffix(strings.TrimSpace(name), ".")
	z := strings.TrimSuffix(strings.TrimSpace(zoneName), ".")
	if n == z {
		return ""
	}
	suf := "." + z
	if strings.HasSuffix(n, suf) {
		return strings.TrimSuffix(n, suf)
	}
	return n
}

func fqdnRecordName(nodeName, zoneName string) string {
	z := strings.TrimSpace(zoneName)
	if !strings.HasSuffix(z, ".") {
		z += "."
	}
	node := strings.TrimSpace(nodeName)
	if node == "" || node == "@" {
		return z
	}
	return node + "." + z
}

// parseRecordRef splits "recordsetUUID|index" for multi-value record sets.
func parseRecordRef(id string) (recordSetID string, index int, indexed bool) {
	base, suf, ok := strings.Cut(id, "|")
	if !ok {
		return strings.TrimSpace(id), 0, false
	}
	idx, err := strconv.Atoi(suf)
	if err != nil || base == "" {
		return strings.TrimSpace(id), 0, false
	}
	return base, idx, true
}

func parseMXValue(s string) (priority int, host string) {
	s = strings.TrimSpace(s)
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return 0, strings.TrimSuffix(s, ".")
	}
	priority, _ = strconv.Atoi(fields[0])
	host = strings.TrimSpace(strings.Join(fields[1:], " "))
	host = strings.TrimSuffix(host, ".")
	return priority, host
}

func formatMXValue(priority int, host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return ""
	}
	return fmt.Sprintf("%d %s.", priority, host)
}

func (c *Client) ListPublicZones(ctx context.Context, apiKey string) ([]model.PublicZoneResp, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	pub := "public"
	var out []model.PublicZoneResp
	var offset int32
	for {
		req := &model.ListPublicZonesRequest{
			Type:   &pub,
			Limit:  ptrInt32(pageLimit),
			Offset: ptrInt32(offset),
		}
		resp, err := client.ListPublicZones(req)
		if err != nil {
			return nil, fmt.Errorf("list public zones: %w", err)
		}
		if resp.Zones == nil {
			break
		}
		zones := *resp.Zones
		out = append(out, zones...)
		if int32(len(zones)) < pageLimit {
			break
		}
		offset += int32(len(zones))
		if resp.Metadata != nil && resp.Metadata.TotalCount != nil && offset >= *resp.Metadata.TotalCount {
			break
		}
	}
	return out, nil
}

func (c *Client) ShowPublicZone(ctx context.Context, apiKey, zoneID string) (*model.ShowPublicZoneResponse, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	return client.ShowPublicZone(&model.ShowPublicZoneRequest{ZoneId: zoneID})
}

func (c *Client) ListRecordSetsByZone(ctx context.Context, apiKey, zoneID string) ([]model.ListRecordSets, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	var out []model.ListRecordSets
	var offset int32
	for {
		req := &model.ListRecordSetsByZoneRequest{
			ZoneId: zoneID,
			Limit:  ptrInt32(pageLimit),
			Offset: ptrInt32(offset),
		}
		resp, err := client.ListRecordSetsByZone(req)
		if err != nil {
			return nil, fmt.Errorf("list record sets: %w", err)
		}
		if resp.Recordsets == nil {
			break
		}
		rs := *resp.Recordsets
		out = append(out, rs...)
		if int32(len(rs)) < pageLimit {
			break
		}
		offset += int32(len(rs))
		if resp.Metadata != nil && resp.Metadata.TotalCount != nil && offset >= *resp.Metadata.TotalCount {
			break
		}
	}
	return out, nil
}

func (c *Client) ShowRecordSet(ctx context.Context, apiKey, zoneID, recordSetID string) (*model.ShowRecordSetResponse, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	return client.ShowRecordSet(&model.ShowRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: recordSetID,
	})
}

func (c *Client) CreateRecordSet(ctx context.Context, apiKey string, body *model.CreateRecordSetRequest) (*model.CreateRecordSetResponse, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	return client.CreateRecordSet(body)
}

func (c *Client) UpdateRecordSet(ctx context.Context, apiKey string, req *model.UpdateRecordSetRequest) (*model.UpdateRecordSetResponse, error) {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return nil, err
	}
	return client.UpdateRecordSet(req)
}

func (c *Client) DeleteRecordSet(ctx context.Context, apiKey, zoneID, recordSetID string) error {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return err
	}
	_, err = client.DeleteRecordSet(&model.DeleteRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: recordSetID,
	})
	return err
}

func (c *Client) SetRecordSetStatus(ctx context.Context, apiKey, recordSetID, status string) error {
	_ = ctx
	client, err := c.newDNSClient(apiKey)
	if err != nil {
		return err
	}
	_, err = client.SetRecordSetsStatus(&model.SetRecordSetsStatusRequest{
		RecordsetId: recordSetID,
		Body: &model.SetRecordSetsStatusRequestBody{
			Status: status,
		},
	})
	return err
}

func ptrInt32(i int32) *int32 {
	return &i
}
