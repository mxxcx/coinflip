package rest

import (
	"database/sql"
	"time"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/mtdx/ns-ga/coinflip"
	"github.com/mtdx/ns-ga/config"
	"github.com/mtdx/ns-ga/cors"
)

const (
	MAX_MULTIPLIER       = 10000 // max crash multiplier
	INSTANT_CRASH_CHANCE = 2     // instant crash chance,
)

var r *chi.Mux

func routes() {
	tokenAuth := jwtauth.New("HS256", []byte(config.JwtKey()), nil)
	r.Route("/api", func(r chi.Router) {
		r.Get("/coinflip", coinflip.GetGamesHandler)
		r.Get("/coinflip-top-players", coinflip.GetTopPlayersHandler)
		r.Get("/coinflipws", coinflip.WebsocketsHandler)

		r.Group(func(r chi.Router) { // Protected routes
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator)

			r.Post("/coinflip", coinflip.GameCreateHandler)
			r.Put("/coinflip/{gameID:[0-9]+}", coinflip.JoinGameHandler)
			r.Delete("/coinflip/{gameID:[0-9]+}", coinflip.DeleteGameHandler)
			r.Get("/coinflip-history", coinflip.GetGamesHistoryHandler)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
}

// Router create chi router & add the routes
func Router(dbconn *sql.DB) *chi.Mux {
	r = chi.NewRouter()

	r.Use(corsConfig().Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.WithValue("DBCONN", dbconn))

	routes()

	return r
}

func corsConfig() *cors.Cors {
	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	  })

	  return cors;
} 