package main

import (
	"errors"
	"net/http"

	"github.com/abdulmajid18/keyVal/key_value/internal/data"
	"github.com/abdulmajid18/keyVal/key_value/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected data from the request body
	var input struct {
		UserName string `json:"username"`
		Email    string `json:"email"`
		DbName   string `json:"dbname"`
		Password string `json:"password"`
	}

	// Parse the request body into he anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data from the request body into a new User struct. Notice also that we
	// set the Activated field to false, which isn't strictly necessary because the
	// Activated field will have the zero-value of false by default. But setting this
	// explicitly helps to make our intentions clear to anyone reading the code.
	user := &data.User{
		UserName:  input.UserName,
		Email:     input.Email,
		DbName:    input.DbName,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = user.GenerateKey()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	// Validate the user struct and return the error messages to the client if any of
	// the checks fail.
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually
		// add a message to the validator instance, and then call our
		// failedValidationResponse() helper.
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Use the background helper to execute an anonymous function that sends the welcome
	// email.
	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.Println(err, nil)
		}
	})

	// Write a JSON response containing the user data along with a 201 Created status
	// code.
	// err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// }

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
