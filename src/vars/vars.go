// Package vars contains all globally used variables and their default values. Furthermore,
// the package also contains all internal errors since they are just variables to golang
package vars

import (
	"database/sql"

	"microservice/structs"
)

// ServiceName is the name of the service which is used for identifying it in the gateway
// TODO: Change the service name and remove the TODO comment
const ServiceName = "template-service"

// ===== Required Setting Variables =====
var (
	// APIGatewayHost contains the IP address or hostname of the used Kong API Gateway
	APIGatewayHost string

	// ServiceRoutePath is the path under which the instance of the microservice shall be reachable via the Kong API
	// Gateway
	ServiceRoutePath string

	// DatabaseHost specifies the host on which the postgres database runs on
	DatabaseHost string

	// DatabaseUser is the username of the postgres user accessing the database
	DatabaseUser string

	// DatabaseUserPassword is the password of the user accessing the database
	DatabaseUserPassword string
)

// ===== Optional Setting Variables =====
var (
	// ListenPort is the port this microservice will listen on. It defaults to 8000
	ListenPort int = 8000

	// DatabasePort specifies on which port the database used listens on
	DatabasePort int = 5432

	// ScopeConfigurationPath specifies from where the service should read the configuration of the needed access scope
	ScopeConfigurationPath string = "/res/scope.json"

	// APIGatewayPort contains the port on which the admin api of the Kong API Gateway listens
	APIGatewayPort int = 8001
)

// ===== Globally used variables =====

// PostgresConnection is as connection shared throughout the service
var PostgresConnection *sql.DB

// ScopeConfiguration containing the information about the scope needed to access this service
var ScopeConfiguration *structs.ScopeInformation

// ExecuteHealthcheck is an indicator for the microservice if the service shall execute a healthcheck.
// You can trigger a health check by starting the executable with -healthcheck
var ExecuteHealthcheck bool
