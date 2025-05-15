package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"microservice/internal/configuration"
)

// This file contains the connection to the database which is automatically
// initialized on import/app startup

// Pool is not initialized at the app startup and needs to be initiatized by
// calling [Connect].
var pool *pgxpool.Pool

// Errors which are returned if the database configuration is not in order.
var (
	ErrNoDatabaseUser          = errors.New("no database user configured")
	ErrNoDatabasePassword      = errors.New("no database password configured")
	ErrNoDatabaseHost          = errors.New("no database host configured")
	ErrPoolConfigurationFailed = errors.New("unable to initialize database pool")
	ErrPoolPingFailed          = errors.New("unable to ping database via pool")
)

const (
	KeyUser     = configuration.ConfigurationKey_DatabaseUser
	KeyPassword = configuration.ConfigurationKey_DatabasePassword
	KeyHost     = configuration.ConfigurationKey_DatabaseHost
	KeyPort     = configuration.ConfigurationKey_DatabasePort
	KeySSLMode  = configuration.ConfigurationKey_DatabaseSSLMode
	KeyDatabase = configuration.ConfigurationKey_DatabaseName
)

const pgSqlConnString = `user=%s password=%s host=%s port=%d sslmode=%s database=%s`

func Pool() *pgxpool.Pool {
	return pool
}

func Connect() (err error) {
	slog.Info("initializing database connection")

	config := configuration.Default.Viper()

	if !config.IsSet(KeyHost) {
		return ErrNoDatabaseHost
	}
	if !config.IsSet(KeyUser) {
		return ErrNoDatabaseUser
	}
	if !config.IsSet(KeyPassword) {
		return ErrNoDatabasePassword
	}

	connectionString := fmt.Sprintf(pgSqlConnString,
		config.GetString(KeyUser), config.GetString(KeyPassword),
		config.GetString(KeyHost), config.GetInt(KeyPort),
		config.GetString(KeySSLMode), config.GetString(KeyDatabase),
	)
	slog.Debug("generated connection string", "connString", connectionString)

	pgConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return fmt.Errorf("unable to parse database configuration string: %w", err)
	}

	slog.Debug("initializing database pool with connection string", "connString", connectionString)
	pool, err = pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrPoolConfigurationFailed.Error(), err)
	}

	slog.Info("validating database connection")
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("%s: %w", ErrPoolPingFailed.Error(), err)
	}
	return nil
}
