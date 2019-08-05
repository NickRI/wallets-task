package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NickRI/wallets-task/db/models"
	"github.com/NickRI/wallets-task/domain/services"
	"github.com/go-chi/chi"
	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

type SendRequest struct {
	Sender   string          `json:"-"`
	Receiver string          `json:"-"`
	Amount   decimal.Decimal `json:"amount"`
}

func PaymentDecoder(ctx context.Context, r *http.Request) (interface{}, error) {
	var req SendRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, models.ValidationError{xerrors.Errorf("error while json decoding: %w", e)}
	}

	req.Sender = chi.URLParam(r, "sender")
	req.Receiver = chi.URLParam(r, "receiver")

	return req, nil
}

func PaymentSend(ws services.Wallet) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SendRequest)
		return ws.Send(ctx, req.Sender, req.Receiver, req.Amount)
	}
}
