package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jcarloasilo/golang-rest-template/internal/database"
	"github.com/jcarloasilo/golang-rest-template/internal/request"
	"github.com/jcarloasilo/golang-rest-template/internal/response"
	"github.com/jcarloasilo/golang-rest-template/internal/validator"
)

func (app *application) handlerEmailConfirmation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code      string              `json:"code"`
		Validator validator.Validator `json:"-"`
	}

	user := contextGetAuthenticatedUser(r)

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	existingOTP, err := app.db.GetLatestOTP(r.Context(), database.GetLatestOTPParams{
		UserID: user.ID,
		Type:   database.OtpTypeEmailVerification,
	})
	if err != nil {
		app.notFound(w, r)
		return
	}

	if existingOTP.ExpiresAt.Before(time.Now()) {
		app.badRequest(w, r, errors.New("expired otp"))
		return
	}

	if existingOTP.Attempts >= existingOTP.MaxAttempts {
		app.badRequest(w, r, errors.New("too many failed attempts"))
		return
	}

	if input.Code != existingOTP.Code {
		input.Validator.CheckField(false, "code", "Invalid OTP")

		err = app.db.IncrementOTPAttempts(r.Context(), existingOTP.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	err = app.db.VerifyUser(r.Context(), database.VerifyUserParams{
		UserID:     user.ID,
		VerifiedAt: time.Now(),
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = response.JSON(w, http.StatusNoContent, nil)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) handlerNewEmailConfirmation(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	existingOTP, err := app.db.GetLatestOTP(r.Context(), database.GetLatestOTPParams{
		UserID: user.ID,
		Type:   database.OtpTypeEmailVerification,
	})
	noExistingOTP := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !noExistingOTP {
		app.serverError(w, r, err)
		return
	}

	if !noExistingOTP && !existingOTP.ExpiresAt.Before(time.Now()) {
		app.badRequest(w, r, errors.New("a valid otp is already active"))
		return
	}

	if !noExistingOTP {
		err = app.db.InvalidateExistingOTP(r.Context(), database.InvalidateExistingOTPParams{
			UserID: user.ID,
			Type:   database.OtpTypeEmailVerification,
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	otp, err := app.generateOTP(6)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	now := time.Now()

	err = app.db.CreateOTP(r.Context(), database.CreateOTPParams{
		Code:      otp,
		Type:      database.OtpTypeEmailVerification,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Minute * 5),
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.backgroundTask(r, func() error {
		type EmailData struct {
			Name string
			Code string
		}

		err = app.mailer.Send(user.Email, EmailData{
			Name: user.Name,
			Code: otp,
		}, "email_confirmation.tmpl")

		if err != nil {
			return err
		}

		return nil
	})

	err = response.JSON(w, http.StatusNoContent, nil)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
