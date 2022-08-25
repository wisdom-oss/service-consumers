// This file contains all globally used variables and their default values
package vars

import (
	"database/sql"

	"microservice/structs"
)

const ServiceName = "consumer-management"

var (
	// The host on which the API gateway runs and this service shall be registered on
	ApiGatewayHost string
	// The administration port of the api gateway on the host
	ApiGatewayAdminPort string
	// The path on which the service shall be reachable
	ApiGatewayServicePath string
	// The http port on which the service will listen for new requests
	HttpListenPort string
	// Indicator if a health check shall be executed instead of the main() function
	ExecuteHealthcheck bool
	// ScopeConfigFilePath
	// The path to the location where the scope configuration is stored.
	// Default: /microservice/res/scope.json
	ScopeConfigFilePath string
	// The scope the service was configured with
	Scope *structs.ScopeInformation
	/*
		PostgresHost

		The host on which the consumer database is running on
	*/
	PostgresHost string
	/*
		PostgresUser

		The user used to access the postgres database
	*/
	PostgresUser string
	/*
		PostgresPassword

		The password of the user used to access the postgres database
	*/
	PostgresPassword string
	/*
		PostgresPort

		The port on which the postgres database listens on for new connections. Default value: 5432
	*/
	PostgresPort string
	/*
		PostgresConnection

		The postgres connection which has been made during the initialization.
		This connection is shared throughout the microservice
	*/
	PostgresConnection *sql.DB
)
