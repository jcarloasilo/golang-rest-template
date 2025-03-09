package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
)

func (app *application) newEmailData() map[string]any {
	data := map[string]any{
		"BaseURL": app.config.baseURL,
	}

	return data
}

func (app *application) backgroundTask(r *http.Request, fn func() error) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				app.reportServerError(r, fmt.Errorf("%s", err))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	}()
}

func (app *application) generateOTP(length int) (string, error) {
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charSetLen := len(charSet)
	otp := make([]byte, length)

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(charSetLen)))
		if err != nil {
			return "", err // Return error if random number generation fails
		}
		otp[i] = charSet[randomIndex.Int64()]
	}

	return string(otp), nil
}
