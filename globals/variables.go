package globals

import (
	"github.com/qustavo/dotsql"
	wisdomType "github.com/wisdom-oss/commonTypes"
)

// This file contains globally shared variables (e.g., service name, sql queries)

// ServiceName contains the global identifier for the service
const ServiceName = "consumer-management"

// SqlQueries contains the prepared sql queries from the resources folder
var SqlQueries *dotsql.DotSql

// AuthorizationConfiguration contains the configuration of the Authorization
// middleware for this microservice
var AuthorizationConfiguration wisdomType.AuthorizationConfiguration

// Environment contains a mapping between the environment variables and the values
// they were set to. However, this variable only contains the configured environment
// variables
var Environment map[string]string = make(map[string]string)

// Errors contains all errors that have been predefined in the "errors.json" file.
var Errors map[string]wisdomType.WISdoMError = make(map[string]wisdomType.WISdoMError)
