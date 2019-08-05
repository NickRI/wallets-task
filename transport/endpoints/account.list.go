package endpoints

import (
	"context"

	"github.com/NickRI/wallets-task/domain/services"
	"github.com/go-kit/kit/endpoint"
)

func AccountList(ws services.Wallet) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return ws.AccountsList(ctx)
	}
}
