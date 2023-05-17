// Package globals contains the globally shared variables like connection
// pointers and the environment mapping
package globals

import (
	"github.com/gchaincl/dotsql"
	"github.com/go-chi/httplog"
)

// ScopeInformation contains the information about the scope for this service
type ScopeInformation struct {
	JSONSchema       string `json:"$schema"`
	ScopeName        string `json:"name"`
	ScopeDescription string `json:"description"`
	ScopeValue       string `json:"scopeStringValue"`
}

// ServiceName sets the string by which the service is identified in the
// API gateway and in the logging
// TODO: Change this name to a appropriate one after inserting this template
var ServiceName = "consumers"

// Environment contains the environment variables that have been specified
// in the environment.json5 file
var Environment map[string]string = make(map[string]string)

// HttpLogger is the logger which is used by code interacting with the
// webserver
var HttpLogger = httplog.NewLogger(ServiceName, httplog.Options{JSON: true})

// ScopeConfiguration contains the scope needed to access this service
var ScopeConfiguration ScopeInformation

// Queries contains the sql queries loaded by the microservice
var Queries *dotsql.DotSql
