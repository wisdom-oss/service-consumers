// This file contains all functions used to start the microservice. Put further prerequisites which may need to be
// initialized into this file
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	gateway "github.com/wisdom-oss/golang-kong-access"
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
	var apiGatewayHostSet, apiGatewayAdminPortSet, apiGatewayServicePathSet, httpListenPortSet,
		scopeConfigFilePathSet, postgresHostSet, postgresUserSet, postgresPasswordSet, postgresPortSet bool
	vars.APIGatewayHost, apiGatewayHostSet = os.LookupEnv("CONFIG_API_GATEWAY_HOST")
	tmpAdminPort, apiGatewayAdminPortSet := os.LookupEnv("CONFIG_API_GATEWAY_ADMIN_PORT")
	vars.ServiceRoutePath, apiGatewayServicePathSet = os.LookupEnv("CONFIG_API_GATEWAY_SERVICE_PATH")
	vars.ListenPort, httpListenPortSet = os.LookupEnv("CONFIG_HTTP_LISTEN_PORT")
	vars.DatabaseHost, postgresHostSet = os.LookupEnv("CONFIG_POSTGRES_HOST")
	vars.DatabaseUser, postgresUserSet = os.LookupEnv("CONFIG_POSTGRES_USER")
	vars.DatabaseUserPassword, postgresPasswordSet = os.LookupEnv("CONFIG_POSTGRES_PASSWORD")
	vars.DatabasePort, postgresPortSet = os.LookupEnv("CONFIG_POSTGRES_PORT")
	// Now check the results of the environment variable lookup and check if the string did not only contain whitespaces
	if !apiGatewayHostSet || strings.TrimSpace(vars.APIGatewayHost) == "" {
		logger.Fatal("The required environment variable 'CONFIG_API_GATEWAY_HOST' is not populated.")
	}
	if !apiGatewayAdminPortSet || strings.TrimSpace(tmpAdminPort) == "" {
		logger.Fatal("The required environment variable 'CONFIG_API_GATEWAY_ADMIN_PORT' is not populated.")
	}
	if !apiGatewayServicePathSet || strings.TrimSpace(vars.ServiceRoutePath) == "" {
		logger.Fatal("The required environment variable 'CONFIG_API_GATEWAY_SERVICE_PATH' is not populated.")
	}
	if !postgresHostSet || strings.TrimSpace(vars.DatabaseHost) == "" {
		logger.Fatal("The required environment variable 'CONFIG_POSTGRES_HOST' is not populated.")
	}
	if !postgresUserSet || strings.TrimSpace(vars.DatabaseUser) == "" {
		logger.Fatal("The required environment variable 'CONFIG_POSTGRES_USER' is not populated.")
	}
	if !postgresPasswordSet || strings.TrimSpace(vars.DatabaseUserPassword) == "" {
		logger.Fatal("The required environment variable 'CONFIG_POSTGRES_PASSWORD' is not populated.")
	}
	// Now check if the optional variables have been set. If not set their respective default values
	// TODO: Add checks for own optional variables, if needed
	if !httpListenPortSet {
		vars.ListenPort = "8000"
	}
	if _, err := strconv.Atoi(vars.ListenPort); err != nil {
		logger.Warning("The http listen port which has been set is not a number. Defaulting to 8000")
		vars.ListenPort = "8000"
	}
	if !postgresPortSet {
		vars.DatabasePort = "5432"
	}
	if _, err := strconv.Atoi(vars.DatabasePort); err != nil {
		logger.Warning("The postgres port which has been set is not a number. Defaulting to 5432")
		vars.DatabasePort = "5432"
	}
	if !apiGatewayAdminPortSet {
		vars.APIGatewayPort = 8001
	}
	tmpAdminPortInt, err := strconv.Atoi(tmpAdminPort)
	if err != nil {
		logger.Warning("The gateway admin api port has not been set to a number. Defaulting to 8001")
		vars.APIGatewayPort = 8001
	} else {
		vars.APIGatewayPort = tmpAdminPortInt
	}

	vars.ScopeConfigurationPath, scopeConfigFilePathSet = os.LookupEnv("CONFIG_SCOPE_FILE_PATH")
	if !scopeConfigFilePathSet {
		vars.ScopeConfigurationPath = "/microservice/res/scope.json"
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
	logger.Infof("Checking if the api gateway on the host '%s' is reachable on port '%s'", vars.APIGatewayHost,
		vars.APIGatewayPort)
	gatewayReachable := helpers.PingHost(vars.APIGatewayHost,
		vars.APIGatewayPort, 10)
	if !gatewayReachable {
		logger.Fatalf("The api gateway on the host '%s' is not reachable on port '%s'", vars.APIGatewayHost,
			vars.APIGatewayPort)
	} else {
		logger.Info("The api gateway is reachable via tcp")
	}
	// Check if a connection to the postgres database is possible
	logger.Info("Checking if the postgres database is reachable and the login data is valid")
	postgresConnectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=wisdom sslmode=disable",
		vars.DatabaseHost, vars.DatabasePort, vars.DatabaseUser, vars.DatabaseUserPassword)
	logger.Debugf("Built the follwoing connection string: '%s'", postgresConnectionString)
	// Create a possible error object
	var connectionError error
	logger.Info("Opening the connection to the consumer database")
	vars.PostgresConnection, connectionError = sql.Open("postgres", postgresConnectionString)
	if connectionError != nil {
		logger.WithError(connectionError).Fatal("Unable to connect to the consumer database.")
	}
	// Now ping the database to check if the connection is working
	databasePingError := vars.PostgresConnection.Ping()
	if databasePingError != nil {
		logger.WithError(databasePingError).Fatal("Unable to ping to the consumer database.")
	}
	logger.Info("The connection to the consumer database was successfully established")
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
	logger.Infof("Reading the scope configuration file from '%s'", vars.ScopeConfigurationPath)
	fileContents, err := ioutil.ReadFile(vars.ScopeConfigurationPath)
	if err != nil {
		logger.WithError(err).Fatal("Unable to read the contents of the scope configuration file")
	}
	logger.Debugf("Read the following file contents: %s", fileContents)
	logger.Debug("Parsing the file contents into the scope configuration for the service")

	parserError := json.Unmarshal(fileContents, &vars.ScopeConfiguration)
	if parserError != nil {
		logger.WithError(parserError).Fatalf("Unable to parse the contents of '%s'", vars.ScopeConfigurationPath)
	}
}

/*
Initialization Step 6 - Register service in upstream of the microservice and setup routing

This initialization step will use the admin api of the api gateway to add itself to the upstream for the service
instances. If no upstream is set up, one will be created automatically
*/
func init() {
	if !vars.ExecuteHealthcheck {
		setupErr := gateway.SetUpGatewayConnection(vars.APIGatewayHost, vars.APIGatewayPort, false)
		if setupErr != nil {
			log.WithError(setupErr).Fatal("Unable to set up the connection to the api gateway")
		}
		upstreamSetUp, err := gateway.IsUpstreamSetUp(vars.ServiceName)
		if err != nil {
			log.WithError(err).Fatal("Unable to check if the service already has a upstream set up")
		}
		if !upstreamSetUp {
			upstreamCreated, err := gateway.CreateNewUpstream(vars.ServiceName)
			if err != nil {
				log.WithError(err).Fatal("Unable to create a new upstream for this microservice")
			}
			if !upstreamCreated {
				log.Fatal("The upstream was not created even though no error occurred")
			} else {
				log.Info("Successfully created a new upstream for the microservice")
			}
		}

		// Get the local ip address to add it to the upstream targets
		localIPAddress := helpers.GetLocalIP()
		targetAddress := fmt.Sprintf("%s:%s", localIPAddress, vars.ListenPort)

		targetInUpstream, err := gateway.IsAddressInUpstreamTargetList(targetAddress, vars.ServiceName)
		if err != nil {
			log.WithError(err).Fatal("Unable to check if the address of the container is listed in the upstream of" +
				" the microservice")
		}
		if !targetInUpstream {
			// Build the target address

			targetAdded, err := gateway.CreateTargetInUpstream(targetAddress, vars.ServiceName)
			if err != nil {
				log.WithError(err).Fatal("Unable to add the address of the container to the upstream of the microservice")
			}
			if !targetAdded {
				log.Fatal("The target address was not added to the upstream of the service")
			}
		}

		serviceSetUp, err := gateway.IsServiceSetUp(vars.ServiceName)
		if err != nil {
			log.WithError(err).Fatal("Unable to check if the microservice already has a service configured on the" +
				" gateway")
		}
		if !serviceSetUp {
			log.Warning("No service was previously set up for this microservice. " +
				"Creating a new service on the api gateway")

			// Create a new service using the previously created/existing upstream as target of the service
			serviceCreated, err := gateway.CreateService(vars.ServiceName, vars.ServiceName)
			if err != nil {
				log.WithError(err).Fatal("Unable to create a new service for the microservice")
			}
			if !serviceCreated {
				log.Fatal("The service has not been created due to an unknown error")
			}
		}

		routeSetUp, err := gateway.ServiceHasRouteSetUp(vars.ServiceName)
		if err != nil {
			log.WithError(err).Fatal("Unable to check if the service of the microservice has any routes configured")
		}
		if !routeSetUp {
			routeCreated, err := gateway.CreateNewRoute(vars.ServiceName, vars.ServiceRoutePath)
			if err != nil {
				log.WithError(err).Fatal("Unable to create a route for the service")
			}
			if !routeCreated {
				log.Fatal("The route was not created due to an unknown reason")
			}

		} else {
			routeWithPathExists, err := gateway.ServiceHasRouteWithPathSetUp(vars.ServiceName, vars.ServiceRoutePath)
			if err != nil {
				log.WithError(err).Fatal("Unable to check if the service of the microservice has a route configured" +
					" matching the path supplied by the environment")
			}
			if !routeWithPathExists {
				routeCreated, err := gateway.CreateNewRoute(vars.ServiceName, vars.ServiceRoutePath)
				if err != nil {
					log.WithError(err).Fatal("Unable to create a route for the service")
				}
				if !routeCreated {
					log.Fatal("The route was not created due to an unknown reason")
				}
			}
		}
	}
}
