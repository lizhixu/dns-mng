package desec

// deSEC API response types

// Domain represents a domain in deSEC
type Domain struct {
	Name       string `json:"name"`
	Created    string `json:"created"`
	Published  string `json:"published,omitempty"`
	MinimumTTL int    `json:"minimum_ttl,omitempty"`
	Touched    string `json:"touched,omitempty"`
}

// RRSet represents a resource record set
type RRSet struct {
	Domain  string   `json:"domain"`
	Subname string   `json:"subname"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
	TTL     int      `json:"ttl"`
	Created string   `json:"created,omitempty"`
	Touched string   `json:"touched,omitempty"`
}

// RRSetRequest represents a request to create/update RRSet
type RRSetRequest struct {
	Subname string   `json:"subname"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
	TTL     int      `json:"ttl"`
}
