// Package vars contains all globally used variables and their default values
package vars

import (
	"database/sql"

	"microservice/structs"
)

const ServiceName = "consumer-management"

var (

	// APIGatewayHost contains the IP address or hostname of the used Kong API Gateway
	APIGatewayHost string

	// APIGatewayPort contains the port on which the admin api of the Kong API Gateway listens
	APIGatewayPort int = 8001

	// ServiceRoutePath is the path under which the instance of the microservice shall be reachable via the Kong API
	// Gateway
	ServiceRoutePath string

	// ListenPort is the port this microservice will listen on. It defaults to 8000
	ListenPort string = "8000"

	// ExecuteHealthcheck is an indicator for the microservice if the service shall execute a healthcheck.
	//You can trigger a health check by starting the executable with -healthcheck
	ExecuteHealthcheck bool

	// ScopeConfigurationPath specifies from where the service should read the configuration of the needed access scope
	ScopeConfigurationPath string

	// ScopeConfiguration containing the information about the scope needed to access this service
	ScopeConfiguration *structs.ScopeInformation

	// DatabaseHost specifies the host on which the postgres database runs on
	DatabaseHost string

	// DatabaseUser is the username of the postgres user accessing the database
	DatabaseUser string

	// DatabaseUserPassword is the password of the user accessing the database
	DatabaseUserPassword string

	// DatabasePort specifies on which port the database used listens on
	DatabasePort string

	// PostgresConnection is as connection shared throughout the service
	PostgresConnection *sql.DB
)
