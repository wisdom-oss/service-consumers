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

	"microservice/utils"
	"microservice/vars"
)

// RequiredSettings associates the name of an environment variable with a pointer to the storage location of the value
var RequiredSettings = map[string]*string{
	"API_GATEWAY_HOST": &vars.APIGatewayHost,
	"PG_HOST":          &vars.DatabaseHost,
	"PG_USER":          &vars.DatabaseUser,
	"PG_PASS":          &vars.DatabaseUserPassword,
	"SERVICE_PATH":     &vars.ServiceRoutePath,
	// TODO: Add own required settings
}

// OptionalIntSettings associates the name of an environment variable with a pointer to the storage location of the
// value. If the value is not found a default value will be loaded
var OptionalIntSettings = map[string]*int{
	"LISTEN_PORT":      &vars.ListenPort,
	"PG_PORT":          &vars.DatabasePort,
	"API_GATEWAY_PORT": &vars.APIGatewayPort,
}

// OptionalStringSettings associates the name of an environment variable with a pointer to the storage location of the
// value. If the value is not found a default value will be loaded
var OptionalStringSettings = map[string]*string{
	"SCOPE_FILE_LOCATION": &vars.ScopeConfigurationPath,
}

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
	log.SetFormatter(
		&log.TextFormatter{
			// Display the full time stamp in the logs
			FullTimestamp: true,
			// Show the levels name fully, even though this may result in shifts between the log lines
			DisableLevelTruncation: true,
		},
	)
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
	logger := log.WithFields(
		log.Fields{
			"initStep":     3,
			"initStepName": "CONFIGURATION_CHECK",
		},
	)
	logger.Debug("Validating the required environment variables for their existence and if the variables are not empty")
	// Check the required variables for their values
	for envName, valuePointer := range RequiredSettings {
		var err error
		*valuePointer, err = utils.ReadEnvironmentVariable(envName)
		if err != nil {
			logger.WithError(err).Fatalf("The required environment variable '%s' is not set", envName)
		}
	}

	// Now check the default integer variables if they exist and are convertible
	for envName, valuePointer := range OptionalIntSettings {
		stringValue, err := utils.ReadEnvironmentVariable(envName)
		if err != nil || strings.TrimSpace(stringValue) == "" {
			logger.Infof("Using default value '%d' for environment variable '%s'", *valuePointer, envName)
		} else {
			intValue, conversionError := strconv.Atoi(stringValue)
			if conversionError != nil {
				logger.WithError(conversionError).Warningf(
					"Using default value '%d' for environment variable '%s'",
					*valuePointer, envName,
				)
			} else {
				*valuePointer = intValue
			}
		}
	}

	// Now check for the optional setting strings
	for envName, valuePointer := range OptionalStringSettings {
		stringValue, err := utils.ReadEnvironmentVariable(envName)
		if err != nil || strings.TrimSpace(stringValue) == "" {
			logger.Infof("Using default value '%s' for environment variavble '%s'", *valuePointer, envName)
		} else {
			*valuePointer = stringValue
		}
	}

}

/*
Initialization Step 4 - Check the dependency connections

This initialization step will check if all dependency containers are reachable.

TODO: Add checks for new dependencies
*/
func init() {
	// Create a logger for this step
	logger := log.WithFields(
		log.Fields{
			"initStep":     4,
			"initStepName": "DEPENDENCY_CONNECTION_CHECK",
		},
	)
	// Check if the kong admin api is reachable
	logger.Infof(
		"Checking if the api gateway on the host '%s' is reachable on port '%d'", vars.APIGatewayHost,
		vars.APIGatewayPort,
	)
	gatewayReachable := utils.PingHost(
		vars.APIGatewayHost,
		vars.APIGatewayPort, 10,
	)
	if !gatewayReachable {
		logger.Fatalf(
			"The api gateway on the host '%s' is not reachable on port '%d'", vars.APIGatewayHost,
			vars.APIGatewayPort,
		)
	} else {
		logger.Info("The api gateway is reachable via tcp")
	}
	// Check if a connection to the postgres database is possible
	logger.Info("Checking if the postgres database is reachable and the login data is valid")
	postgresConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=wisdom sslmode=disable",
		vars.DatabaseHost, vars.DatabasePort, vars.DatabaseUser, vars.DatabaseUserPassword,
	)
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

This initialization step will load the supplied scope-example.json file to get the information needed for checking the incoming
requests for the correct scope
*/
func init() {
	logger := log.WithFields(
		log.Fields{
			"initStep":     5,
			"initStepName": "OAUTH2_SCOPE_CONFIGURATION",
		},
	)
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
	logger := log.WithFields(
		log.Fields{
			"initStep":     6,
			"initStepName": "GATEWAY_SET_UP",
		},
	)
	if !vars.ExecuteHealthcheck {
		setupErr := gateway.SetUpGatewayConnection(vars.APIGatewayHost, vars.APIGatewayPort, false)
		if setupErr != nil {
			logger.WithError(setupErr).Fatal("Unable to set up the connection to the api gateway")
		}
		upstreamSetUp, err := gateway.IsUpstreamSetUp(vars.ServiceName)
		if err != nil {
			logger.WithError(err).Fatal("Unable to check if the service already has a upstream set up")
		}
		if !upstreamSetUp {
			upstreamCreated, err := gateway.CreateNewUpstream(vars.ServiceName)
			if err != nil {
				logger.WithError(err).Fatal("Unable to create a new upstream for this microservice")
			}
			if !upstreamCreated {
				logger.Fatal("The upstream was not created even though no error occurred")
			} else {
				logger.Info("Successfully created a new upstream for the microservice")
			}
		} else {
			logger.Info("The service already has a upstream entry in the database")
		}

		// Get the local ip address to add it to the upstream targets
		localIPAddress, _ := utils.LocalIPv4Address()
		targetAddress := fmt.Sprintf("%s:%d", localIPAddress, vars.ListenPort)

		targetInUpstream, err := gateway.IsAddressInUpstreamTargetList(targetAddress, vars.ServiceName)
		if err != nil {
			logger.WithError(err).Fatal(
				"Unable to check if the address of the container is listed in the upstream of" +
					" the microservice",
			)
		}
		if !targetInUpstream {
			// Build the target address

			targetAdded, err := gateway.CreateTargetInUpstream(targetAddress, vars.ServiceName)
			if err != nil {
				logger.WithError(err).Fatal("Unable to add the address of the container to the upstream of the microservice")
			}
			if !targetAdded {
				logger.Fatal("The target address was not added to the upstream of the service")
			} else {
				logger.Infof("Added the microservices ip address and listen port to the upstream targets")
			}
		} else {
			logger.Info("The microservices ip address and listen port are already listed as upstream targets")
		}

		serviceSetUp, err := gateway.IsServiceSetUp(vars.ServiceName)
		if err != nil {
			logger.WithError(err).Fatal(
				"Unable to check if the microservice already has a service configured on the" +
					" gateway",
			)
		}
		if !serviceSetUp {
			logger.Warning(
				"No service was previously set up for this microservice. " +
					"Creating a new service on the api gateway",
			)

			// Create a new service using the previously created/existing upstream as target of the service
			serviceCreated, err := gateway.CreateService(vars.ServiceName, vars.ServiceName)
			if err != nil {
				logger.WithError(err).Fatal("Unable to create a new service for the microservice")
			}
			if !serviceCreated {
				logger.Fatal("The service has not been created due to an unknown error")
			} else {
				logger.Info("Successfully created a new service entry for the microservice")
			}
		} else {
			logger.Info("The microservice already has a service entry set up")
		}

		routeSetUp, err := gateway.ServiceHasRouteSetUp(vars.ServiceName)
		if err != nil {
			logger.WithError(err).Fatal("Unable to check if the service of the microservice has any routes configured")
		}
		if !routeSetUp {
			routeCreated, err := gateway.CreateNewRoute(vars.ServiceName, vars.ServiceRoutePath)
			if err != nil {
				logger.WithError(err).Fatal("Unable to create a route for the service")
			}
			if !routeCreated {
				logger.Fatal("The route was not created due to an unknown reason")
			} else {
				logger.Info("The route was successfully created for this microservice")
			}
		} else {
			routeWithPathExists, err := gateway.ServiceHasRouteWithPathSetUp(vars.ServiceName, vars.ServiceRoutePath)
			if err != nil {
				logger.WithError(err).Fatal(
					"Unable to check if the service of the microservice has a route configured" +
						" matching the path supplied by the environment",
				)
			}
			if !routeWithPathExists {
				routeCreated, err := gateway.CreateNewRoute(vars.ServiceName, vars.ServiceRoutePath)
				if err != nil {
					logger.WithError(err).Fatal("Unable to create a route for the service")
				}
				if !routeCreated {
					logger.Fatal("The route was not created due to an unknown reason")
				}
			} else {
				logger.Info("The requested route already exists")
			}
		}
	}
}
