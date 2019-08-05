package models

import (
	"time"

	"github.com/NickRI/wallets-task/domain/entities"

	"github.com/NickRI/wallets-task/db/common"

	"github.com/shopspring/decimal"
)

type Account struct {
	Id        int64
	UserName  string
	Balance   decimal.Decimal
	Currency  string
	UpdatedAt time.Time
	CreatedAt time.Time
}

func (a *Account) Bind() []interface{} {
	return []interface{}{
		&common.NullInt64{V: &a.Id},
		&common.NullString{V: &a.UserName},
		&common.NullDecimal{V: &a.Balance},
		&common.NullString{V: &a.Currency},
		&common.NullTime{V: &a.UpdatedAt},
		&common.NullTime{V: &a.CreatedAt},
	}
}

func (a *Account) ToDomain() *entities.Account {
	acc := &entities.Account{
		AccountId: entities.AccountId(a.UserName),
		Balance:   a.Balance,
		Currency:  entities.Currency(a.Currency),
	}
	acc.SetId(a.Id)
	return acc
}

type AccountList []*Account

func (al *AccountList) Add(a *Account) {
	*al = append(*al, a)
}

//ToDomain hydrate list to domain
func (al AccountList) ToDomain() (accs entities.Accounts) {
	for _, a := range al {
		accs.Add(a.ToDomain())
	}
	return
}
