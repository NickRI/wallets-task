package restapi

import (
	"github.com/NickRI/wallets-task/domain/services"
	"github.com/NickRI/wallets-task/transport/endpoints"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

func MakeRoutes(w services.Wallet, l log.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&KitLoggerWrapper{l}))

	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(endpoints.ErrorEncoder),
	}

	handlers := MakeHandlers(w, options...)

	// RESTy routes for "wallet" resource
	r.Route("/wallet/", func(r chi.Router) {
		r.Post("/pay/{sender}/{receiver}", handlers.Send.ServeHTTP)
		r.Get("/ledgers", handlers.ListLedgers.ServeHTTP)
		r.Get("/accounts", handlers.ListAccounts.ServeHTTP)
	})

	return r
}
