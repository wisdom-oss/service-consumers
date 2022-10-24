package main

import (
	context2 "context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"microservice/handlers"
	"microservice/vars"
)

/*
This function is used to set up the http server for the microservice
*/
func main() {
	if vars.ExecuteHealthcheck {
		response, err := http.Get("http://localhost:" + vars.ListenPort + "/ping")
		if err != nil {
			os.Exit(1)
		}
		if response.StatusCode != 204 {
			os.Exit(1)
		}
		return
	}

	// Set up the routing of the different functions
	router := mux.NewRouter()
	router.HandleFunc("/ping", handlers.PingHandler)
	router.Handle("/{consumer_id}", handlers.AuthorizationCheck(
		http.HandlerFunc(handlers.UpdateConsumerInformation),
	)).Methods("PATCH")
	router.Handle("/{consumer_id}", handlers.AuthorizationCheck(
		http.HandlerFunc(handlers.DeleteConsumerFromDatabase),
	)).Methods("DELETE")
	router.Handle("/", handlers.AuthorizationCheck(
		http.HandlerFunc(handlers.GetConsumers),
	)).Methods("GET")
	router.Handle("/", handlers.AuthorizationCheck(
		http.HandlerFunc(handlers.CreateNewConsumer),
	)).Methods("PUT")

	// Configure the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", vars.ListenPort),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	// Start the server and log errors that happen while running it
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("An error occurred while starting the http server")
		}
	}()

	// Set up the signal handling to allow the server to shut down gracefully
	log.Info("The microservice is now able to accept and handle requests")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, os.Interrupt)

	// Block further code execution until the shutdown signal was received
	<-cancelSignal

	context, cancel := context2.WithTimeout(context2.Background(), time.Second*15)
	defer cancel()

	go func() {
		err := server.Shutdown(context)
		if err != nil {
			log.WithError(err).Fatal("An error occurred while stopping the http server")
		}
	}()
	log.Info("Shutting down the microservice...")
}
