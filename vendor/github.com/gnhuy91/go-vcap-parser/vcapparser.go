package vcapparser

import "encoding/json"

type Credentials struct {
	// Genenal field
	URI string `json:"uri,omitempty"`
	// Postgres service
	ID         int    `json:"ID,omitempty"`
	BindingID  string `json:"binding_id,omitempty"`
	Database   string `json:"database,omitempty"`
	DSN        string `json:"dsn,omitempty"`
	Host       string `json:"host,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
	JdbcURI    string `json:"jdbc_uri,omitempty"`
	Password   string `json:"password,omitempty"`
	Port       string `json:"port,omitempty"`
	Username   string `json:"username,omitempty"`
	// UAA service
	IssuerID  string            `json:"issuerId,omitempty"`
	Subdomain string            `json:"subdomain,omitempty"`
	Zone      map[string]string `json:"zone,omitempty"`
}

type VcapService struct {
	Credentials    Credentials `json:"credentials"`
	Label          string      `json:"label"`
	Name           string      `json:"name"`
	Plan           string      `json:"plan"`
	Provider       string      `json:"provider"`
	SyslogDrainURL string      `json:"syslog_drain_url"`
	Tags           []string    `json:"tags"`
}

// VcapServices is a map of services detail
type VcapServices map[string][]VcapService

// ParseVcapServices parse string provided from VCAP_SERVICES environment var
// to VcapServices struct.
func ParseVcapServices(vcapStr string) (VcapServices, error) {
	var vcapServices VcapServices

	if err := json.Unmarshal([]byte(vcapStr), &vcapServices); err != nil {
		return vcapServices, err
	}

	return vcapServices, nil
}
