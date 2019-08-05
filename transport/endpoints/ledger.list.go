package endpoints

import (
	"context"

	"github.com/NickRI/wallets-task/domain/services"
	"github.com/go-kit/kit/endpoint"
)

func LedgerList(ws services.Wallet) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return ws.LedgersList(ctx)
	}
}
