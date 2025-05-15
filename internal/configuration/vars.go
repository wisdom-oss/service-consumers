package configuration

// This file contains the sensible default values that are used in the
// configuration as well as other variables used for the configuration and the
// reading of it.

var environmentVariables = map[string][]string{
	ConfigurationKey_DatabaseHost: {
		"PGHOST", "PG_HOST", "POSTGRES_HOST", "DB_HOST", "DATABASE_HOST",
	},
	ConfigurationKey_DatabasePort: {
		"PGPORT", "PG_PORT", "POSTGRES_PORT", "DB_PORT", "DATABASE_PORT",
	},
	ConfigurationKey_DatabaseUser: {
		"PGUSER", "PG_USER", "POSTGRES_USER", "DB_USER", "DATABASE_USER",
	},
	ConfigurationKey_DatabasePassword: {
		"PGPASSWORD", "PG_PASSWORD", "PGPASS", "PG_PASS", "POSTGRES_PASS",
		"POSTGRES_PASSWORD", "DB_PASS", "DB_PASSWORD", "DATABASE_PASSWORD",
	},
	ConfigurationKey_DatabaseSSLMode: {
		"PGSSLMODE", "PG_SSLMODE", "PG_SSL_MODE", "POSTGRES_SSLMODE",
		"POSTGRES_SSL_MODE", "DB_SSLMODE", "DB_SSL_MODE", "DATABASE_SSLMODE",
		"DATABASE_SSL_MODE",
	},
	ConfigurationKey_HttpPort:              {"HTTP_PORT"},
	ConfigurationKey_AuthorizationRequired: {"AUTH_REQUIRED", "AUTHORIZATION_REQUIRED"},
	ConfigurationKey_OidcAuthority:         {"OIDC_AUTHORITY", "OIDC_ISSUER"},
}

var defaults = map[string]any{
	ConfigurationKey_DatabasePort:    5432, //nolint:mnd
	ConfigurationKey_DatabaseSSLMode: "disable",
	ConfigurationKey_DatabaseName:    "wisdom",
	ConfigurationKey_HttpPort:        8000, //nolint:mnd
}
