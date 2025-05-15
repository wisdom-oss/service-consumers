package configuration

// and necessary credentials.
const (
	EnvConfigurationType = "AWARD_CONFIG_TYPE"
)

// The supported configuration types of the microservice.
const (
	ConfigurationType_Vault = "vault"
	ConfigurationType_Local = "local"
)

const (
	DatabaseCredentialType_Static  = "static"
	DatabaseCredentialType_Dynamic = "dynamic"
)

// The most common configuration keys used in a basic microservice configuration.
// These keys should be extended if your service uses additional keys to ensure
// a consistent usage throughout your code.
const (
	ConfigurationKey_DatabaseUser            = "database.user"
	ConfigurationKey_DatabasePassword        = "database.password"
	ConfigurationKey_DatabaseHost            = "database.host"
	ConfigurationKey_DatabasePort            = "database.port"
	ConfigurationKey_DatabaseSSLMode         = "database.ssl-mode"
	ConfigurationKey_DatabaseName            = "database.name"             // translates to the database in postgres
	ConfigurationKey_DatabaseCredentialType  = "database.credential-type"  // used for vault reading
	ConfigurationKey_DatabaseCredentialRole  = "database.credential-role"  // used for vault reading
	ConfigurationKey_DatabaseCredentialMount = "database.credential-mount" // used for vault reading

	ConfigurationKey_HttpHost = "http.host"
	ConfigurationKey_HttpPort = "http.port"

	ConfigurationKey_OidcAuthority = "oidc.auhority"

	ConfigurationKey_AuthorizationRequired = "authorization.required"
)
