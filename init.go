package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	wisdomType "github.com/wisdom-oss/commonTypes"

	"github.com/qustavo/dotsql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/wisdom-oss/service-consumers/globals"

	_ "github.com/lib/pq"
)

var l zerolog.Logger

// defaultAuth contains the default authentication configuration if no file
// is present (which shouldn't be the case). it only allows named users
// access to this service who use the same group as the service name
var defaultAuth = wisdomType.AuthorizationConfiguration{
	Enabled:                   true,
	RequireUserIdentification: true,
	RequiredUserGroup:         globals.ServiceName,
}

// this init functions sets up the logger which is used for this microservice
func init() {
	// set the time format to unix timestamps to allow easier machine handling
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// allow the logger to create an error stack for the logs
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// now use the environment variable `LOG_LEVEL` to determine the logging
	// level for the microservice.
	rawLoggingLevel, isSet := os.LookupEnv("LOG_LEVEL")

	// if the value is not set, use the info level as default.
	var loggingLevel zerolog.Level
	if !isSet {
		loggingLevel = zerolog.InfoLevel
	} else {
		// now try to parse the value of the raw logging level to a logging
		// level for the zerolog package
		var err error
		loggingLevel, err = zerolog.ParseLevel(rawLoggingLevel)
		if err != nil {
			// since an error occurred while parsing the logging level, use info
			loggingLevel = zerolog.InfoLevel
			log.Warn().Msg("unable to parse value from environment. using info")
		}
	}
	// since now a logging level is set, configure the logger
	zerolog.SetGlobalLevel(loggingLevel)
	l = log.With().Str("step", "init").Logger()
}

// this function initializes the environment variables used in this microservice
// and validates that the configured variables are present.
func init() {
	l.Info().Msg("loading environment for microservice")

	// now check if the default location for the environment configuration
	// was changed via the `ENV_CONFIG_LOCATION` variable
	location, locationChanged := os.LookupEnv("ENV_CONFIG_LOCATION")
	if !locationChanged {
		// since the location has not changed, set the default value
		location = "./environment.json"
		l.Debug().Msg("location for environment config not changed")
	}
	l.Info().Str("path", location).Msg("loading environment configuration file")
	var c wisdomType.EnvironmentConfiguration
	err := c.PopulateFromFilePath(location)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to load environment configuration")
	}
	l.Info().Msg("successfully loaded environment configuration")

	// since the configuration was successfully loaded, check the required
	// environment variables
	l.Info().Msg("validating configuration against current environment")
	globals.Environment, err = c.ParseEnvironment()
	if err != nil {
		l.Fatal().Err(err).Msg("error while parsing environment")
	}
}

// this function now loads the prepared errors from the error file and parses
// them into wisdom errors
func init() {
	l.Info().Msg("loading predefined errors")
	// check if the error file location was set
	filePath, isSet := globals.Environment["ERROR_FILE_LOCATION"]
	if !isSet {
		l.Fatal().Msg("no error file location set in environment")
	}
	// now check if the path is not empty
	if filePath == "" || strings.TrimSpace(filePath) == "" {
		l.Fatal().Msg("empty path supplied for error file location")
	}

	// since the path is not empty, try to open it
	file, err := os.Open(filePath)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to open error configuration file")
	}

	var errors []wisdomType.WISdoMError
	err = json.NewDecoder(file).Decode(&errors)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to load error configuration file")
	}
	for _, e := range errors {
		e.InferHttpStatusText()
		globals.Errors[e.ErrorCode] = e
	}
	l.Info().Msg("loaded predefined errors")
}

// this function loads the externally defined authorization configuration
// and overwrites the default options laid out here
func init() {
	l.Info().Msg("loading authorization configuration")
	filePath, isSet := globals.Environment["AUTH_CONFIG_FILE_LOCATION"]
	if !isSet {
		l.Warn().Msg("no auth file location set in environment. using default")
		globals.AuthorizationConfiguration = defaultAuth
		return
	}
	// now check if the path is not empty
	if filePath == "" || strings.TrimSpace(filePath) == "" {
		l.Warn().Msg("empty path supplied for error file location. using default")
		globals.AuthorizationConfiguration = defaultAuth
		return
	}

	// since a file was found, read from the file path
	var authConfig wisdomType.AuthorizationConfiguration
	err := authConfig.PopulateFromFilePath(filePath)
	if err != nil {
		l.Error().Err(err).Msg("unable to parse authorization configuration. ussing default")
		globals.AuthorizationConfiguration = defaultAuth
		return
	}

	globals.AuthorizationConfiguration = authConfig
	l.Info().Msg("loaded authorization configuration")
}

// this function opens a global connection to the postgres database used for
// this microservice and loads the prepared sql queries.
func init() {
	l.Info().Msg("preparing global database connection")
	// build a dsn from the environment variables
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=wisdom sslmode=disable",
		globals.Environment["PG_HOST"], globals.Environment["PG_PORT"], globals.Environment["PG_USER"],
		globals.Environment["PG_PASS"])

	// now open the connection to the database
	var err error
	globals.Db, err = sql.Open("postgres", dsn)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to open database connection")
	}
	l.Info().Msg("opened database connection")

	// now ping the database to check the connectivity
	l.Info().Msg("pinging the database to verify connectivity")
	err = globals.Db.Ping()
	if err != nil {
		l.Fatal().Err(err).Msg("connectivity verification failed")
	}
	l.Info().Msg("database connection verified. open and working")

	// now load the prepared sql queries
	l.Info().Msg("loading sql queries")
	globals.SqlQueries, err = dotsql.LoadFromFile(globals.Environment["QUERY_FILE_LOCATION"])
	if err != nil {
		l.Fatal().Err(err).Msg("unable to load queries used by the service")
	}
}

// this function just logs that the init process is finished
func init() {
	l.Info().Msg("finished initialization")
}
