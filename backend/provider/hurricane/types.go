package hurricane

// Hurricane Electric (dns.he.net) types
// Note: HE doesn't have an official API, this uses their web interface

// Domain represents a DNS zone in Hurricane Electric
type Domain struct {
	ID          string
	Name        string
	Status      string
	Serial      string
	MasterIP    string
	RecordCount int
}

// Record represents a DNS record
type Record struct {
	ID       string
	ZoneID   string
	Name     string
	Type     string
	Content  string
	TTL      int
	Priority int
	Dynamic  bool
}

// LoginResponse represents the login session
type LoginResponse struct {
	SessionID string
	Cookies   map[string]string
}
