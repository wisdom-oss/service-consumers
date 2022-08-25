package gateway

type TargetInformation struct {
	Id        string  `json:"id"`
	CreatedAt float64 `json:"created_at"`
	Upstream  struct {
		Id string `json:"id"`
	} `json:"upstream"`
	Address string   `json:"target"`
	Weight  int      `json:"weight"`
	Tags    []string `json:"tags"`
}

type ServiceConfiguration struct {
	Id                string   `json:"id"`
	CreatedAt         float64  `json:"created_at"`
	UpdatedAt         float64  `json:"updated_at"`
	Name              string   `json:"name"`
	Retries           int      `json:"retries"`
	Protocol          string   `json:"protocol"`
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	Path              string   `json:"path"`
	ConnectTimeout    int      `json:"connect_timeout"`
	WriteTimeout      int      `json:"write_timeout"`
	ReadTimeout       int      `json:"read_timeout"`
	Tags              []string `json:"tags"`
	ClientCertificate struct {
		Id string `json:"id"`
	} `json:"client_certificate"`
	TlsVerify      bool        `json:"tls_verify"`
	TlsVerifyDepth interface{} `json:"tls_verify_depth"`
	CaCertificates []string    `json:"ca_certificates"`
	Enabled        bool        `json:"enabled"`
}

type RouteInformation struct {
	Id                      string            `json:"id"`
	CreatedAt               float64           `json:"created_at"`
	UpdatedAt               float64           `json:"updated_at"`
	Name                    string            `json:"name"`
	Protocols               []string          `json:"protocols"`
	Methods                 []string          `json:"methods"`
	Hosts                   []string          `json:"hosts"`
	Paths                   []string          `json:"paths"`
	Headers                 map[string]string `json:"headers"`
	HttpsRedirectStatusCode int               `json:"https_redirect_status_code"`
	RegexPriority           int               `json:"regex_priority"`
	StripPath               bool              `json:"strip_path"`
	PathHandling            string            `json:"path_handling"`
	PreserveHost            bool              `json:"preserve_host"`
	RequestBuffering        bool              `json:"request_buffering"`
	ResponseBuffering       bool              `json:"response_buffering"`
	Tags                    []string          `json:"tags"`
	Service                 struct {
		Id string `json:"id"`
	} `json:"service"`
}
