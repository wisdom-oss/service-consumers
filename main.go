package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"microservice/internal/configuration"
	"microservice/internal/db"
	"microservice/router"
)

var headerReadTimeout = 10 * time.Second
var serverShutdownTimeout = 20 * time.Second

// the main function bootstraps the http server and handlers used for this
// microservice.
func main() {

	if err := configuration.Default.Initialize(); err != nil {
		slog.Error("unable to initialize configuration", "error", err)
		os.Exit(1)
	}

	if err := configuration.Default.Read(); err != nil {
		slog.Error("unable to parse configuration", "error", err)
		os.Exit(1)
	}

	// setting up the database connection
	err := db.Connect()
	if err != nil {
		slog.Error("unable to connect to the database", "error", err)
		os.Exit(1)
	}

	// running database migrations stored in resources/migrations
	err = db.MigrateDatabase()
	if err != nil {
		slog.Error("failed to execute database migrations", "error", err)
		os.Exit(1)
	}

	// configure your router
	r, err := router.Configure()
	if err != nil {
		slog.Error("unable to create router", "error", err)
		os.Exit(1)
	}

	c := configuration.Default.Viper()

	// create a http server to handle the requests
	server := http.Server{
		Addr:              net.JoinHostPort(c.GetString(configuration.ConfigurationKey_HttpHost), c.GetString(configuration.ConfigurationKey_HttpPort)), //nolint:lll
		Handler:           r.Handler(),
		ReadHeaderTimeout: headerReadTimeout,
	}

	// Start the server and log errors that happen while running it
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("unable to start http server", "error", err)
		}
	}()

	// Set up some the signal handling to allow the server to shut down gracefully
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// Block further code execution until the shutdown signal was received
	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		slog.Error("unable to shutdown api gracefully", "error", err)
		slog.Error("forcing shutdown...")
		return
	}

}
