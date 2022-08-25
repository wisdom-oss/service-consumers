// This file contains all functions used to start the microservice. Put further prerequisites which may need to be
// initialized into this file
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"microservice/gateway"
	"microservice/helpers"
	"microservice/vars"
)

/*
Initialization Step 1 - Flag Creation

This initialization step will create a boolean flag which may trigger a healthcheck later on
*/
func init() {
	// Create a new boolean variable flag which uses an existing variable pointer for the value to be assigned
	flag.BoolVar(
		&vars.ExecuteHealthcheck,
		"healthcheck",
		false,
		"Run a healthcheck of the service which will check if the service can call itself and is correctly setup on the API gateway",
	)
	// Parse the created flags into their variables
	flag.Parse()
}

/*
Initialization Step 2 - Logger Configuration

This step will set up the logging library logrus for this microservice and set the correct logging level
*/
func init() {
	// Check if a logging level was set in the environment variables
	rawLoggingLevel, envFound := os.LookupEnv("CONFIG_LOGGING_LEVEL")
	// If the logging level was not set use info as default level
	if !envFound || (envFound && rawLoggingLevel == "") {
		rawLoggingLevel = "info"
	}
	// Parse the raw value to a logging level which is understood by logrus
	logrusLoggingLevel, err := log.ParseLevel(rawLoggingLevel)
	// If an unknown logging level was supplied use the Info level as default level
	if err != nil {
		logrusLoggingLevel = log.InfoLevel
	}
	// Set the level for the logging library
	log.SetLevel(logrusLoggingLevel)
	// Set the formatter for the logging library
	log.SetFormatter(&log.TextFormatter{
		// Display the full time stamp in the logs
		FullTimestamp: true,
		// Show the levels name fully, even though this may result in shifts between the log lines
		DisableLevelTruncation: true,
	})
}

/*
Initialization Step 3 - Required environment variable check

This initialization step will check the existence of the following variables and if the values are not empty strings:
	- CONFIG_API_GATEWAY_HOST
	- CONFIG_API_GATEWAY_ADMIN_PORT
	- CONFIG_API_GATEWAY_SERVICE_PATH

Furthermore, this step will use sensitive defaults on the following environment variables
	- CONFIG_HTTP_LISTEN_PORT = 8000

TODO: Add own required and optional variables to this function, if needed
*/
func init() {
	logger := log.WithFields(log.Fields{
		"initStep":     3,
		"initStepName": "CONFIGURATION_CHECK",
	})
	logger.Debug("Validating the required environment variables for their existence and if the variables are not empty")
	// Use os.LookupEnv to check if the variables are existent in the environment, but ignore their values since
	// they have already been read once
	var apiGatewayHostSet, apiGatewayAdminPortSet, apiGatewayServicePathSet, httpListenPortSet, scopeConfigFilePathSet bool
	vars.ApiGatewayHost, apiGatewayHostSet = os.LookupEnv("CONFIG_API_GATEWAY_HOST")
	vars.ApiGatewayAdminPort, apiGatewayAdminPortSet = os.LookupEnv("CONFIG_API_GATEWAY_ADMIN_PORT")
	vars.ApiGatewayServicePath, apiGatewayServicePathSet = os.LookupEnv("CONFIG_API_GATEWAY_SERVICE_PATH")
	// Now check the results of the environment variable lookup and check if the string did not only contain whitespaces
	if !apiGatewayHostSet || strings.TrimSpace(vars.ApiGatewayHost) == "" {
		logger.Fatal("The required environment variable 'CONFIG_API_GATEWAY_HOST' is not populated.")
	}
	if !apiGatewayAdminPortSet || strings.TrimSpace(vars.ApiGatewayAdminPort) == "" {
		logger.Fatal("The required environment variable 'CONFIG_API_GATEWAY_ADMIN_PORT' is not populated.")
	}
	if !apiGatewayServicePathSet || strings.TrimSpace(vars.ApiGatewayServicePath) == "" {
		logger.Fatal("The required environment variable 'CONFIG_API_GATEWAY_SERVICE_PATH' is not populated.")
	}
	// Now check if the optional variables have been set. If not set their respective default values
	// TODO: Add checks for own optional variables, if needed
	vars.HttpListenPort, httpListenPortSet = os.LookupEnv("CONFIG_HTTP_LISTEN_PORT")
	if !httpListenPortSet {
		vars.HttpListenPort = "8000"
	}
	if _, err := strconv.Atoi(vars.HttpListenPort); err != nil {
		logger.Warning("The http listen port which has been set is not a number. Defaulting to 8000")
		vars.HttpListenPort = "8000"
	}
	vars.ScopeConfigFilePath, scopeConfigFilePathSet = os.LookupEnv("CONFIG_SCOPE_FILE_PATH")
	if !scopeConfigFilePathSet {
		vars.ScopeConfigFilePath = "/microservice/res/scope.json"
	}
}

/*
Initialization Step 4 - Check the dependency connections

This initialization step will check if all dependency containers are reachable.

TODO: Add checks for new dependencies
*/
func init() {
	// Create a logger for this step
	logger := log.WithFields(log.Fields{
		"initStep":     4,
		"initStepName": "DEPENDENCY_CONNECTION_CHECK",
	})
	// Check if the kong admin api is reachable
	logger.Infof("Checking if the api gateway on the host '%s' is reachable on port '%s'", vars.ApiGatewayHost,
		vars.ApiGatewayAdminPort)
	gatewayReachable := helpers.PingHost(vars.ApiGatewayHost,
		vars.ApiGatewayAdminPort, 10)
	if !gatewayReachable {
		logger.Fatalf("The api gateway on the host '%s' is not reachable on port '%s'", vars.ApiGatewayHost,
			vars.ApiGatewayAdminPort)
	} else {
		logger.Info("The api gateway is reachable via tcp")
	}
}

/**
Initialization Step 5 - Load the scope setup for this service

This initialization step will load the supplied scope.json file to get the information needed for checking the incoming
requests for the correct scope
*/
func init() {
	logger := log.WithFields(log.Fields{
		"initStep":     5,
		"initStepName": "OAUTH2_SCOPE_CONFIGURATION",
	})
	logger.Infof("Reading the scope configuration file from '%s'", vars.ScopeConfigFilePath)
	fileContents, err := ioutil.ReadFile(vars.ScopeConfigFilePath)
	if err != nil {
		logger.WithError(err).Fatal("Unable to read the contents of the scope configuration file")
	}
	logger.Debugf("Read the following file contents: %s", fileContents)
	logger.Debug("Parsing the file contents into the scope configuration for the service")

	parserError := json.Unmarshal(fileContents, &vars.Scope)
	if parserError != nil {
		logger.WithError(parserError).Fatalf("Unable to parse the contents of '%s'", vars.ScopeConfigFilePath)
	}
}

/**
Initialization Step 6 - Register service in upstream of the microservice and setup routing

This initialization step will use the admin api of the api gateway to add itself to the upstream for the service
instances. If no upstream is set up, one will be created automatically
*/
func init() {
	// Since this is the fist call to the api gateway we need to prepare the calls to the gateway
	gateway.PrepareGatewayConnections()
	// Now check if the upstream is already set up
	if !gateway.IsUpstreamSetUp() {
		gateway.CreateUpstream()
	}
	// Now check if this service instance is listed in the upstreams targets
	if !gateway.IsIPAddressInUpstreamTargets() {
		gateway.AddServiceToUpstreamTargets()
	}
	// Now check if a service entry exists for this service
	if !gateway.IsServiceSetUp() {
		gateway.CreateServiceEntry()
	}
	// Now check if the service entry has the upstream already configured as host
	if !gateway.IsUpstreamSetInServiceEntry() {
		gateway.SetUpstreamAsServiceEntryHost()
	}
	// Now check if the service entry has a route matching the configuration
	if !gateway.IsRouteConfigured() {
		gateway.ConfigureRoute()
	}
}
