package main

import "net/http"

func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	if err := app.writeJSON(w, status, envelope{"error": message}, nil); err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	msg := "server error"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "resource not found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := "method not allowed"
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusBadRequest, errors)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "edit conflict"
	app.errorResponse(w, r, http.StatusConflict, msg)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) requireAuthenticationResponse(w http.ResponseWriter, r *http.Request) {
	msg := "you must be authenticated"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	msg := "account must be active"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}

func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	msg := "doesn't have the necessary permissions"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}
