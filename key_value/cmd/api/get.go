package main

import (
	"fmt"
	"net/http"

	"github.com/abdulmajid18/keyVal/key_value/internal/data"
	"github.com/abdulmajid18/keyVal/key_value/internal/validator"
	"github.com/gorilla/mux"
)

func (app *application) GetHandler(w http.ResponseWriter, r *http.Request) {
	var input data.GetData
	vars := mux.Vars(r)
	secret_key := vars["secret_key"]

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidateGetData(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	state, err := app.models.Put.CheckExistenceDB(secret_key)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	if state {
		value, state, err := data.Get(input)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		switch {
		case !state:
			app.writeJSON(w, http.StatusNotFound, "Value for entered Key not Avialable", nil)
		case state:
			data := fmt.Sprintf("Key:%s    Value:%s ", input.Key, value)
			app.writeJSON(w, http.StatusFound, data, nil)
		}
	}

}
