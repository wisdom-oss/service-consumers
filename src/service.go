package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/wisdom-oss/microservice-middlewares"
	"microservice/request/routes"
	"microservice/vars/globals"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
This function is used to set up the http server for the microservice
*/
func main() {

	// Set up the routing of the different functions
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(httplog.Handler(globals.HttpLogger))
	router.Use(middleware.Heartbeat("/healthcheck"))
	router.Use(wisdomMiddleware.GatewayConfigInterceptor(globals.Environment["GATEWAY_CONFIG_LOCATION"], "/_gatewayConfig"))
	router.Use(wisdomMiddleware.Authorization([]string{"/healthcheck", "/_gatewayConfig"}, globals.ScopeConfiguration.ScopeValue))
	router.Get("/", routes.GetConsumers)
	router.Put("/", routes.CreateConsumer)
	router.Patch("/{consumerID}", routes.UpdateConsumer)
	// Configure the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", globals.Environment["LISTEN_PORT"]),
		WriteTimeout: time.Second * 600,
		ReadTimeout:  time.Second * 600,
		IdleTimeout:  time.Second * 600,
		Handler:      router,
	}

	// Start the server and log errors that happen while running it
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("An error occurred while starting the http server")
		}
	}()

	// Set up the signal handling to allow the server to shut down gracefully

	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, os.Interrupt)

	// Block further code execution until the shutdown signal was received
	<-cancelSignal

}
