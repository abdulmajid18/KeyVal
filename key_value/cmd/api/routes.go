package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandleFunc("/v1/healthcheck", app.healthcheckHandler).Methods("GET")
	router.HandleFunc("/v1/register_user", app.registerUserHandler).Methods("POST")
	router.HandleFunc("/v1/register_user", app.registerUserHandler).Methods("POST")
	router.HandleFunc("/v1/put/{secret_key}", app.PutHandler).Methods("POST")
	router.HandleFunc("/v1/get/{secret_key}", app.GetHandler).Methods("POST")
	return app.recoverPanic(router)
}
