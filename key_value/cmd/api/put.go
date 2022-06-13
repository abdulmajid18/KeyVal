package main

import (
	"net/http"

	"github.com/abdulmajid18/keyVal/key_value/internal/data"
	"github.com/abdulmajid18/keyVal/key_value/internal/validator"
	"github.com/gorilla/mux"
)

func (app *application) PutHandler(w http.ResponseWriter, r *http.Request) {
	var input data.PutData
	vars := mux.Vars(r)
	secret_key := vars["secret_key"]

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidatePutData(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	state, err := app.models.Put.CheckExistenceDB(secret_key)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	if state {
		err = data.Insert(input)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.writeJSON(w, http.StatusCreated, "Created Successfully!", nil)
	}
}
