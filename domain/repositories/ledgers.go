package repositories

import (
	"context"
	"database/sql"

	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination=../../internal/mock/ledgers.go -package=mock github.com/NickRI/wallets-task/domain/repositories Ledgers
type Ledgers interface {
	List(context.Context) (entities.Ledgers, error)
	AddTx(*sql.Tx, context.Context, *entities.Account, *entities.Account, decimal.Decimal) (*entities.Ledger, error)
}
