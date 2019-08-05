package services

import (
	"context"

	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination=../../internal/mock/wallet.go -package=mock github.com/NickRI/wallets-task/domain/services Wallet
type Wallet interface {
	LedgersList(ctx context.Context) (entities.Ledgers, error)
	AccountsList(ctx context.Context) (entities.Accounts, error)
	Send(ctx context.Context, creditId, debitId string, amount decimal.Decimal) (*entities.Ledger, error)
}
