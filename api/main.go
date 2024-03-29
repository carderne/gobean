// Package api is a very basic HTTP API to gobean
package api

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/unrolled/render"

	"github.com/carderne/gobean/bean"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var re *render.Render

var path string

func init() {
	re = render.New()
	if os.Getenv("ENV") == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

// API runs the the HTTP API serving from the provided beancount file
func API(newPath string) {
	path = newPath

	port := os.Getenv("PORT")
	if port == "" {
		port = "6767"
	}

	log.Debug().Str("Port", port).Msg("Starting up on http://localhost:" + port)

	r := chi.NewRouter()

	r.Get("/", health)
	r.Get("/health", health)
	r.Get("/balance", balance)

	log.Fatal().Err(http.ListenAndServe(":"+port, r))
}

func balance(w http.ResponseWriter, r *http.Request) {
	log.Debug().Str("Method", r.Method).Str("URL", r.URL.String()).Msg("Request")
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	ledger := bean.NewLedger(false)
	_, err = ledger.Load(file)
	if err != nil {
		panic(err)
	}
	date := time.Now()
	bals, err := ledger.GetBalances(date)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	re.JSON(w, http.StatusOK, bals)
}

func health(w http.ResponseWriter, r *http.Request) {
	log.Debug().Str("Method", r.Method).Str("URL", r.URL.String()).Msg("Request")
	re.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
