package main

import (
	"net/http"

	"github.com/jcarloasilo/golang-rest-template/internal/response"
)

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a protected handler"))
}

func (app *application) verified(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a verified handler"))
}
