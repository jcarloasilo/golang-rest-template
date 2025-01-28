package main

import (
	"fmt"
	"go-sveltekit/cmd/api/docs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) routes() http.Handler {
	docs.SwaggerInfo.Host = app.config.swagger.host
	docs.SwaggerInfo.Schemes = []string{app.config.swagger.scheme}

	mux := chi.NewRouter()

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)
	mux.Use(app.authenticate)

	mux.Get("/status", app.status)
	mux.Post("/users", app.createUser)
	mux.Post("/authentication-tokens", app.createAuthenticationToken)

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)

		mux.Get("/protected", app.protected)
	})

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireBasicAuthentication)

		mux.Get("/basic-auth-protected", app.protected)
	})

	mux.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s://%s/swagger/doc.json", app.config.swagger.scheme, app.config.swagger.host)),
	))

	return mux
}
