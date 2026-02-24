package dynu

// Dynu API response types

type DynuDomain struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	UnicodeName       string `json:"unicodeName"`
	Token             string `json:"token"`
	State             string `json:"state"`
	Group             string `json:"group"`
	IPv4Address       string `json:"ipv4Address"`
	IPv6Address       string `json:"ipv6Address"`
	TTL               int    `json:"ttl"`
	IPv4              bool   `json:"ipv4"`
	IPv6              bool   `json:"ipv6"`
	IPv4WildcardAlias bool   `json:"ipv4WildcardAlias"`
	IPv6WildcardAlias bool   `json:"ipv6WildcardAlias"`
	CreatedOn         string `json:"createdOn"`
	UpdatedOn         string `json:"updatedOn"`
}

type DynuDomainsResponse struct {
	StatusCode int          `json:"statusCode"`
	Domains    []DynuDomain `json:"domains"`
}

type DynuRecord struct {
	ID         int64  `json:"id"`
	DomainID   int64  `json:"domainId"`
	DomainName string `json:"domainName"`
	NodeName   string `json:"nodeName"`
	Hostname   string `json:"hostname"`
	RecordType string `json:"recordType"`
	TTL        int    `json:"ttl"`
	State      bool   `json:"state"`
	Content    string `json:"content"`
	UpdatedOn  string `json:"updatedOn"`
	// Type-specific fields
	TextData    string `json:"textData,omitempty"`
	Group       string `json:"group,omitempty"`
	IPv4Address string `json:"ipv4Address,omitempty"`
	IPv6Address string `json:"ipv6Address,omitempty"`
	Host        string `json:"host,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Weight      int    `json:"weight,omitempty"`
	Port        int    `json:"port,omitempty"`
}

type DynuRecordsResponse struct {
	StatusCode int          `json:"statusCode"`
	DnsRecords []DynuRecord `json:"dnsRecords"`
}
