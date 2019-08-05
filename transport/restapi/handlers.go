package restapi

import (
	"net/http"

	"github.com/NickRI/wallets-task/domain/services"
	"github.com/NickRI/wallets-task/transport/endpoints"
	kithttp "github.com/go-kit/kit/transport/http"
)

// Handlers holds all go-kit handlers for the service.
type Handlers struct {
	Send         http.Handler
	ListLedgers  http.Handler
	ListAccounts http.Handler
}

// MakeHandlers initializes all go-kit handlers for the service.
func MakeHandlers(ws services.Wallet, options ...kithttp.ServerOption) Handlers {
	return Handlers{
		Send:         kithttp.NewServer(endpoints.PaymentSend(ws), endpoints.PaymentDecoder, endpoints.EncodeResponse, options...),
		ListLedgers:  kithttp.NewServer(endpoints.LedgerList(ws), endpoints.NopDecoder, endpoints.EncodeResponse, options...),
		ListAccounts: kithttp.NewServer(endpoints.AccountList(ws), endpoints.NopDecoder, endpoints.EncodeResponse, options...),
	}
}
