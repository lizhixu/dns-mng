package huaweicloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dns-mng/models"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
)

// Provider implements DNSProvider for Huawei Cloud public DNS (云解析服务).
type Provider struct {
	client *Client
}

func New() *Provider {
	return &Provider{client: NewClient()}
}

func (p *Provider) Name() string {
	return "huaweicloud"
}

func (p *Provider) DisplayName() string {
	return "华为云云解析 DNS"
}

func (p *Provider) WebsiteURL() string {
	return "https://console.huaweicloud.com/dns"
}

func (p *Provider) DefaultTTL() int {
	return 300
}

func (p *Provider) ListDomains(ctx context.Context, apiKey string) ([]models.Domain, error) {
	zones, err := p.client.ListPublicZones(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	out := make([]models.Domain, 0, len(zones))
	for _, z := range zones {
		zid := deref(z.Id)
		zname := strings.TrimSuffix(deref(z.Name), ".")
		if zid == "" || zname == "" {
			continue
		}
		st := "Active"
		if s := deref(z.Status); s != "" && s != "ACTIVE" {
			st = "Inactive"
		}
		out = append(out, models.Domain{
			ID:          zid,
			Name:        zname,
			UnicodeName: zname,
			State:       st,
			CreatedOn:   deref(z.CreatedAt),
			UpdatedOn:   deref(z.UpdatedAt),
		})
	}
	return out, nil
}

func (p *Provider) GetDomain(ctx context.Context, apiKey string, domainID string) (*models.Domain, error) {
	resp, err := p.client.ShowPublicZone(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}
	zname := strings.TrimSuffix(deref(resp.Name), ".")
	if zname == "" {
		zname = domainID
	}
	st := "Active"
	if s := deref(resp.Status); s != "" && s != "ACTIVE" {
		st = "Inactive"
	}
	return &models.Domain{
		ID:          domainID,
		Name:        zname,
		UnicodeName: zname,
		State:       st,
		CreatedOn:   deref(resp.CreatedAt),
		UpdatedOn:   deref(resp.UpdatedAt),
	}, nil
}

func recordSetToModels(zoneID, zoneName string, rs model.ListRecordSets) []models.Record {
	t := strings.ToUpper(strings.TrimSpace(deref(rs.Type)))
	if t == "NS" || t == "SOA" {
		return nil
	}
	if rs.Default != nil && *rs.Default && (t == "NS" || t == "SOA") {
		return nil
	}
	if rs.Records == nil || len(*rs.Records) == 0 {
		return nil
	}
	rid := deref(rs.Id)
	if rid == "" {
		return nil
	}
	ttl := int32(300)
	if rs.Ttl != nil {
		ttl = *rs.Ttl
	}
	active := !strings.EqualFold(deref(rs.Status), "DISABLE")
	node := trimRootZone(deref(rs.Name), zoneName)
	updated := deref(rs.UpdateAt)
	if updated == "" {
		updated = deref(rs.CreateAt)
	}

	recs := *rs.Records
	out := make([]models.Record, 0, len(recs))
	for i, raw := range recs {
		id := rid
		if len(recs) > 1 {
			id = fmt.Sprintf("%s|%d", rid, i)
		}
		content := strings.TrimSpace(raw)
		priority := 0
		if t == "MX" {
			priority, content = parseMXValue(content)
		} else {
			content = strings.TrimSuffix(content, ".")
		}
		out = append(out, models.Record{
			ID:         id,
			DomainID:   zoneID,
			DomainName: zoneName,
			NodeName:   node,
			RecordType: t,
			TTL:        int(ttl),
			State:      active,
			Content:    content,
			Priority:   priority,
			UpdatedOn:  updated,
		})
	}
	return out
}

func (p *Provider) ListRecords(ctx context.Context, apiKey string, domainID string) ([]models.Record, error) {
	zr, err := p.client.ShowPublicZone(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}
	zoneName := strings.TrimSuffix(deref(zr.Name), ".")
	list, err := p.client.ListRecordSetsByZone(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}
	var out []models.Record
	for _, rs := range list {
		out = append(out, recordSetToModels(domainID, zoneName, rs)...)
	}
	return out, nil
}

func buildRecordValue(recType, content string, priority int) (string, error) {
	t := strings.ToUpper(strings.TrimSpace(recType))
	switch t {
	case "MX":
		if content == "" {
			return "", fmt.Errorf("empty MX target")
		}
		return formatMXValue(priority, content), nil
	case "CNAME", "NS", "PTR":
		s := strings.TrimSpace(content)
		s = strings.TrimSuffix(s, ".")
		if s == "" {
			return "", fmt.Errorf("empty record value")
		}
		return s + ".", nil
	default:
		return strings.TrimSpace(content), nil
	}
}

func (p *Provider) CreateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	zr, err := p.client.ShowPublicZone(ctx, apiKey, domainID)
	if err != nil {
		return nil, err
	}
	zoneName := deref(zr.Name)
	if zoneName == "" {
		return nil, fmt.Errorf("empty zone name")
	}
	val, err := buildRecordValue(record.RecordType, record.Content, record.Priority)
	if err != nil {
		return nil, err
	}
	ttl := int32(record.TTL)
	if ttl <= 0 {
		ttl = 300
	}
	fqdn := fqdnRecordName(record.NodeName, zoneName)
	st := "ENABLE"
	if !record.State {
		st = "DISABLE"
	}
	req := &model.CreateRecordSetRequest{
		ZoneId: domainID,
		Body: &model.CreateRecordSetRequestBody{
			Name:    fqdn,
			Type:    strings.ToUpper(strings.TrimSpace(record.RecordType)),
			Records: []string{val},
			Ttl:     ptrInt32(ttl),
			Status:  &st,
		},
	}
	resp, err := p.client.CreateRecordSet(ctx, apiKey, req)
	if err != nil {
		return nil, err
	}
	rid := deref(resp.Id)
	if rid == "" {
		return nil, fmt.Errorf("create record: empty id")
	}
	return &models.Record{
		ID:         rid,
		DomainID:   domainID,
		DomainName: strings.TrimSuffix(zoneName, "."),
		NodeName:   record.NodeName,
		RecordType: strings.ToUpper(strings.TrimSpace(record.RecordType)),
		TTL:        int(ttl),
		State:      record.State,
		Content:    record.Content,
		Priority:   record.Priority,
		UpdatedOn:  time.Now().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UpdateRecord(ctx context.Context, apiKey string, domainID string, record *models.Record) (*models.Record, error) {
	baseID, idx, indexed := parseRecordRef(record.ID)
	sh, err := p.client.ShowRecordSet(ctx, apiKey, domainID, baseID)
	if err != nil {
		return nil, err
	}
	if sh.Records == nil || len(*sh.Records) == 0 {
		return nil, fmt.Errorf("record set has no values")
	}
	records := append([]string(nil), *sh.Records...)
	val, err := buildRecordValue(record.RecordType, record.Content, record.Priority)
	if err != nil {
		return nil, err
	}
	if indexed {
		if idx < 0 || idx >= len(records) {
			return nil, fmt.Errorf("record index out of range")
		}
		records[idx] = val
	} else {
		if len(records) == 1 {
			records[0] = val
		} else {
			return nil, fmt.Errorf("multi-value record set: use indexed id from list")
		}
	}
	ttl := int32(record.TTL)
	if ttl <= 0 {
		ttl = 300
	}
	up := &model.UpdateRecordSetRequest{
		ZoneId:      domainID,
		RecordsetId: baseID,
		Body: &model.UpdateRecordSetReq{
			Name:    sh.Name,
			Type:    sh.Type,
			Ttl:     ptrInt32(ttl),
			Records: &records,
		},
	}
	if _, err := p.client.UpdateRecordSet(ctx, apiKey, up); err != nil {
		return nil, err
	}
	status := "ENABLE"
	if !record.State {
		status = "DISABLE"
	}
	if err := p.client.SetRecordSetStatus(ctx, apiKey, baseID, status); err != nil {
		return nil, err
	}
	return record, nil
}

func (p *Provider) DeleteRecord(ctx context.Context, apiKey string, domainID string, recordID string) error {
	baseID, idx, indexed := parseRecordRef(recordID)
	if !indexed {
		return p.client.DeleteRecordSet(ctx, apiKey, domainID, baseID)
	}
	sh, err := p.client.ShowRecordSet(ctx, apiKey, domainID, baseID)
	if err != nil {
		return err
	}
	if sh.Records == nil {
		return fmt.Errorf("empty record set")
	}
	records := append([]string(nil), *sh.Records...)
	if idx < 0 || idx >= len(records) {
		return fmt.Errorf("record index out of range")
	}
	records = append(records[:idx], records[idx+1:]...)
	if len(records) == 0 {
		return p.client.DeleteRecordSet(ctx, apiKey, domainID, baseID)
	}
	ttl := sh.Ttl
	if ttl == nil {
		ttl = ptrInt32(300)
	}
	up := &model.UpdateRecordSetRequest{
		ZoneId:      domainID,
		RecordsetId: baseID,
		Body: &model.UpdateRecordSetReq{
			Name:    sh.Name,
			Type:    sh.Type,
			Ttl:     ttl,
			Records: &records,
		},
	}
	_, err = p.client.UpdateRecordSet(ctx, apiKey, up)
	return err
}
