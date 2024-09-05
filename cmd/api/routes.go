package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/movies", app.requireActivatedUser(app.listMovieHandler))
	router.HandlerFunc(http.MethodPost, "/movies", app.requireActivatedUser(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/movies/:id", app.requireActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/movies/:id", app.requireActivatedUser(app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationToken)
	router.HandlerFunc(http.MethodPost, "/tokens/password-reset", app.createPasswordResetTokenHandler)

	return app.recoverPanic(app.enableCORS(app.authenticate(router)))
}
