package main

import (
	"errors"
	"go-sveltekit/internal/database"
	"go-sveltekit/internal/password"
	"go-sveltekit/internal/request"
	"go-sveltekit/internal/response"
	"go-sveltekit/internal/validator"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (app *application) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string              `json:"email"`
		Password  string              `json:"password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	_, err = app.db.GetUserByEmail(r.Context(), input.Email)
	notExist := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !notExist {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "email", "Email is required")
	input.Validator.CheckField(validator.Matches(input.Email, validator.RgxEmail), "email", "Must be a valid email address")
	input.Validator.CheckField(notExist, "email", "Email is already in use")

	input.Validator.CheckField(input.Password != "", "password", "Password is required")
	input.Validator.CheckField(len(input.Password) >= 8, "password", "Password is too short")
	input.Validator.CheckField(len(input.Password) <= 72, "password", "Password is too long")
	input.Validator.CheckField(validator.NotIn(input.Password, password.CommonPasswords...), "password", "Password is too common")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_, err = app.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:    input.Email,
		Password: hashedPassword,
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handlerGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	err := response.JSON(w, http.StatusOK, user)
	if err != nil {
		app.serverError(w, r, err)
	}
}
