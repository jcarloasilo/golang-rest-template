package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/jcarloasilo/golang-rest-template/internal/password"
	"github.com/jcarloasilo/golang-rest-template/internal/request"
	"github.com/jcarloasilo/golang-rest-template/internal/response"
	"github.com/jcarloasilo/golang-rest-template/internal/validator"

	"github.com/pascaldekloe/jwt"
)

func (app *application) handlerLogin(w http.ResponseWriter, r *http.Request) {
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

	user, err := app.db.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			input.Validator.CheckField(false, "email", "Email address could not be found")
		default:
			app.serverError(w, r, err)
			return
		}
	}

	input.Validator.CheckField(input.Email != "", "email", "Email is required")

	passwordMatches, err := password.Matches(input.Password, user.HashedPassword)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Password != "", "password", "Password is required")
	input.Validator.CheckField(passwordMatches, "password", "Password is incorrect")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	var claims jwt.Claims
	claims.Subject = user.ID.String()

	expiry := time.Now().Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.config.baseURL
	claims.Audiences = []string{app.config.baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"authentication_token":        string(jwtBytes),
		"authentication_token_expiry": expiry.Format(time.RFC3339),
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}
