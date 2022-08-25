package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"microservice/handlers"
	"microservice/vars"
)

/*
This function is used to set up the http server for the microservice
*/
func main() {
	if vars.ExecuteHealthcheck {
		response, err := http.Get("http://localhost:" + vars.HttpListenPort + "/ping")
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
	router.Handle(
		"/{consumer_id}",
		handlers.AuthorizationCheck(
			http.HandlerFunc(handlers.UpdateConsumerInformation),
		),
	).Methods("PATCH")
	router.Handle(
		"/{consumer_id}",
		handlers.AuthorizationCheck(
			http.HandlerFunc(handlers.DeleteConsumerFromDatabase),
		),
	).Methods("DELETE")
	router.Handle(
		"/",
		handlers.AuthorizationCheck(
			http.HandlerFunc(handlers.GetConsumers),
		),
	).Methods("GET")
	router.Handle(
		"/",
		handlers.AuthorizationCheck(
			http.HandlerFunc(handlers.CreateNewConsumer),
		),
	).Methods("PUT")
}
