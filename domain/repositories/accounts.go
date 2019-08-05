package repositories

import (
	"context"
	"database/sql"

	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination=../../internal/mock/accounts.go -package=mock github.com/NickRI/wallets-task/domain/repositories Accounts
type Accounts interface {
	List(context.Context) (entities.Accounts, error)
	GetByNameTx(*sql.Tx, context.Context, entities.AccountId) (*entities.Account, error)
	UpdateBalanceTx(*sql.Tx, context.Context, *entities.Account, decimal.Decimal) error
	LockTx(*sql.Tx, context.Context) error
}
