// This file contains all globally used variables and their default values
package vars

import (
	"microservice/structs"
)

// TODO: Change me
const ServiceName = "change-me"

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
)
