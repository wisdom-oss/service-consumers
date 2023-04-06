package main

import (
	"database/sql"
	"fmt"
	"github.com/gchaincl/dotsql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/titanous/json5"
	"github.com/wisdom-oss/microservice-utils"
	requestErrors "microservice/request/error"
	"microservice/structs"
	"microservice/vars"
	"microservice/vars/globals"
	"microservice/vars/globals/connections"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// Initialization: configure the logger level and format
func init() {
	// set up the time format and the error logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// now read the environment variable loglevel
	logLevel, _ := os.LookupEnv("LOG_LEVEL")
	logLevel = strings.ToLower(logLevel)
	switch logLevel {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.PanicLevel)
		log.Log().Str("level", logLevel).Str("init-step", "configure-logger").Msg("configured global logger")
		break
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.FatalLevel)
		log.Log().Str("level", logLevel).Str("initStep", "configure-logger").Msg("configured global logger")
		break
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.ErrorLevel)
		log.Log().Str("level", logLevel).Str("initStep", "configure-logger").Msg("configured global logger")
		break
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.WarnLevel)
		log.Log().Str("level", logLevel).Str("initStep", "configure-logger").Msg("configured global logger")
		break
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.InfoLevel)
		log.Log().Str("level", logLevel).Str("initStep", "configure-logger").Msg("configured global logger")
		break
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.DebugLevel)
		log.Log().Str("level", logLevel).Str("initStep", "configure-logger").Msg("configured global logger")
		break
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.TraceLevel)
		log.Log().Str("level", logLevel).Str("initStep", "configure-logger").Msg("configured global logger")
		break
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		globals.HttpLogger = globals.HttpLogger.Level(zerolog.InfoLevel)
		log.Warn().Str("level", "info").Str("initStep", "configure-logger").Msg("configured global logger with default level `info`")
		break
	}
}

// initialization: environment variables as specified from the given path
func init() {
	l := log.With().Str("initStep", "load-environment").Logger()
	// check if the environment variables set a different location for the config
	// file
	l.Info().Msg("loading environment configuration")
	envFileLocation, envSet := os.LookupEnv("ENVIRONMENT_CONFIGURATION")
	var filePath string
	if envSet {
		filePath = envFileLocation
	} else {
		filePath = "/res/environment.json5"
	}
	var environmentConfiguration structs.EnvironmentConfiguration
	configurationContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read environment configuration")
	}
	err = json5.Unmarshal(configurationContent, &environmentConfiguration)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to unmarshal the environment configuration")
	}
	l.Info().Msg("successfully parsed environment configuration")
	l.Info().Msg("loading required environment variables")
	// now iterate through the required environment variables
	for _, key := range environmentConfiguration.RequiredEnvironmentVariables {
		l.Debug().Str("env", key).Msg("reading required environment variable")
		value, isSet := os.LookupEnv(key)
		if !isSet {
			l.Fatal().Err(vars.ErrEnvironmentVariableNotFound).Msgf(
				"the environment variable '%s' is required but not set")
		} else {
			l.Debug().Str("env", key).Msg("found value for environment variable")
			globals.Environment[key] = value
		}
	}
	l.Info().Msg("successfully loaded required environment variables")

	// now iterate through the optional environment variables
	for _, optionalEnvironmentVariable := range environmentConfiguration.OptionalEnvironmentVariables {
		l.Debug().Str("env", optionalEnvironmentVariable.EnvironmentKey).Msg("reading optional environment variable")
		value, isSet := os.LookupEnv(optionalEnvironmentVariable.EnvironmentKey)
		if !isSet {
			l.Debug().Str("env", optionalEnvironmentVariable.EnvironmentKey).Msg("environment variable not found")
			l.Info().Str("env", optionalEnvironmentVariable.EnvironmentKey).Msg("using default value")
			globals.Environment[optionalEnvironmentVariable.EnvironmentKey] = optionalEnvironmentVariable.DefaultValue
		} else {
			l.Debug().Str("env", optionalEnvironmentVariable.EnvironmentKey).Msg("found value for environment variable")
			globals.Environment[optionalEnvironmentVariable.EnvironmentKey] = value
		}
	}

	l.Info().Msg("finished loading of the optional environment variables")
}

// initialization: load the http errors
func init() {
	l := log.With().Str("initStep", "parse-http-errors").Logger()
	l.Info().Msg("reading http error file")
	errorFilePath := globals.Environment["ERROR_FILE_LOCATION"]
	fileContents, err := os.ReadFile(errorFilePath)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to read http error configuration")
	}
	l.Info().Msg("successfully read the http error file contents")
	l.Info().Msg("unmarshalling http error file contents")
	var availableErrors []structs.RequestError
	err = json5.Unmarshal(fileContents, &availableErrors)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to unmarshal the request errors")
	}
	l.Info().Msg("successfully unmarshalled http error file contents")

	// now iterate through the errors and add them to the mapping
	l.Info().Msg("creating request error mapping")
	for _, requestError := range availableErrors {
		l.Debug().Str("error", requestError.ErrorCode).Msg("creating map entry")
		requestErrors.RequestErrors[requestError.ErrorCode] = requestError
	}
	l.Info().Msg("request errors successfully mapped")
}

// initialization: load the scope configuration file
func init() {
	l := log.With().Str("initStep", "parse-oauth-scope").Logger()
	l.Info().Msg("reading file contents of scope file")
	scopeFilePath := globals.Environment["SCOPE_FILE_LOCATION"]
	fileContents, err := os.ReadFile(scopeFilePath)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to read scope file configuration")
	}
	l.Info().Msg("successfully read the scope file contents")
	l.Info().Msg("unmarshalling the scope file contents")
	err = json5.Unmarshal(fileContents, &globals.ScopeConfiguration)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to unmarshal the scope file")
	}
	l.Info().Msg("scope configuration successfully unmarshalled")
}

// initialization: connect to the database
func init() {
	l := log.With().Str("initStep", "connect-postgres").Logger()

	// try to ping the database server to ensure a connection can be opened
	l.Info().Msg("pinging the database host")
	reachable := wisdomUtils.PingHost(globals.Environment["PG_HOST"], globals.Environment["PG_PORT"], 5)
	if !reachable {
		l.Fatal().Msg("unable to ping the postgres database")
	}
	l.Info().Msg("successfully pinged the database host")
	// now try to build the connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=wisdom sslmode=disable",
		globals.Environment["PG_HOST"], globals.Environment["PG_PORT"], globals.Environment["PG_USER"],
		globals.Environment["PG_PASS"])

	// now try to connect to the database
	l.Info().Msg("opening connection to the database")
	var err error
	connections.DbConnection, err = sql.Open("postgres", dsn)
	if err != nil {
		l.Fatal().Err(err).Msg("error during database connection")
	}

	// to test if the connection was successful, ping the database using the function provided by sql
	err = connections.DbConnection.Ping()
	if err != nil {
		l.Fatal().Err(err).Msg("error during database pings")
	}
	l.Info().Msg("database connection established")
}

// initialization: load sql queries
func init() {
	l := log.With().Str("initStep", "load-queries").Logger()
	l.Info().Msg("loading sql queries")
	l.Debug().Msgf("using path: %s", globals.Environment["QUERY_FILE_LOCATION"])
	var err error
	globals.Queries, err = dotsql.LoadFromFile(globals.Environment["QUERY_FILE_LOCATION"])
	if err != nil {
		l.Fatal().Err(err).Msg("unable to load queries used by the service")
	}
}
