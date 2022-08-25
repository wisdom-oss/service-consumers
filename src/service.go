package main

import (
	"net/http"
	"os"

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

	// Set up the HTTP server
	http.Handle("/ping", http.HandlerFunc(handlers.PingHandler))
	// Protect the request handler with authorization
	http.Handle("/", handlers.AuthorizationCheck(http.HandlerFunc(handlers.RequestHandler)))
	// Start the http server
	http.ListenAndServe(":"+vars.HttpListenPort, nil)
}
