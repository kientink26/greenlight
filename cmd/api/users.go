package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/kientink26/greenlight/internal/data"
	"github.com/kientink26/greenlight/internal/validator"
)

func (app *application) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password       string `json:"password"`
		TokenPlaintext string `json:"token"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidatePlaintextPassword(v, input.Password)
	data.ValidatePlaintextToken(v, input.TokenPlaintext)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetForToken(input.TokenPlaintext, data.ScopePasswordReset)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if err = user.Password.Set(input.Password); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err = app.models.Users.Update(user); err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"message": "your password was reset"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	if err := user.Password.Set(input.Password); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if err := app.models.Users.Insert(user); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.background(func() {
		data := map[string]interface{}{
			"userID":          user.ID,
			"activationToken": token.Plaintext,
		}
		if err := app.mailer.Send(user.Email, "user_welcome.tmpl", data); err != nil {
			app.logger.Println(err)
		}
	})
	if err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlainTextToken string `json:"token"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidatePlaintextToken(v, input.PlainTextToken); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetForToken(input.PlainTextToken, data.ScopeActivation)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	user.Activated = true
	if err = app.models.Users.Update(user); err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
