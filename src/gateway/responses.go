package gateway

type UpstreamConfiguration struct {
	Id               string  `json:"id"`
	CreatedAt        float64 `json:"created_at"`
	Name             string  `json:"name"`
	Algorithm        string  `json:"algorithm"`
	HashOn           string  `json:"hash_on"`
	HashFallback     string  `json:"hash_fallback"`
	HashOnCookiePath string  `json:"hash_on_cookie_path"`
	Slots            int     `json:"slots"`
	HealthChecks     struct {
		Active struct {
			Timeout   int `json:"timeout"`
			Unhealthy struct {
				Interval     int   `json:"interval"`
				TcpFailures  int   `json:"tcp_failures"`
				Timeouts     int   `json:"timeouts"`
				HttpFailures int   `json:"http_failures"`
				HttpStatuses []int `json:"http_statuses"`
			} `json:"unhealthy"`
			Type        string `json:"type"`
			Concurrency int    `json:"concurrency"`
			Headers     []struct {
				XAnotherHeader []string `json:"x-another-header"`
				XMyHeader      []string `json:"x-my-header"`
			} `json:"headers"`
			Healthy struct {
				Interval     int   `json:"interval"`
				Successes    int   `json:"successes"`
				HttpStatuses []int `json:"http_statuses"`
			} `json:"healthy"`
			HttpPath               string `json:"http_path"`
			HttpsSni               string `json:"https_sni"`
			HttpsVerifyCertificate bool   `json:"https_verify_certificate"`
		} `json:"active"`
		Passive struct {
			Type      string `json:"type"`
			Unhealthy struct {
				HttpStatuses []int `json:"http_statuses"`
				HttpFailures int   `json:"http_failures"`
				Timeouts     int   `json:"timeouts"`
				TcpFailures  int   `json:"tcp_failures"`
			} `json:"unhealthy"`
			Healthy struct {
				HttpStatuses []int `json:"http_statuses"`
				Successes    int   `json:"successes"`
			} `json:"healthy"`
		} `json:"passive"`
		Threshold int `json:"threshold"`
	} `json:"healthchecks"`
	Tags              []string `json:"tags"`
	HostHeader        string   `json:"host_header"`
	ClientCertificate struct {
		Id string `json:"id"`
	} `json:"client_certificate"`
}

type TargetList struct {
	Targets []TargetInformation `json:"data"`
	Next    string              `json:"next"`
}

type RouteList struct {
	Routes []RouteInformation `json:"data"`
	Next   string             `json:"next"`
}
