package entities

import (
	"strconv"

	"github.com/shopspring/decimal"
)

type AccountId string

func (aid *AccountId) String() string {
	return string(*aid)
}

func (aid *AccountId) Quoted() string {
	return strconv.Quote(aid.String())
}

func (aid *AccountId) MarshalJSON() ([]byte, error) {
	return []byte(aid.Quoted()), nil
}

type Account struct {
	id        int64
	AccountId AccountId       `json:"id"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  Currency        `json:"currency"`
}

func (a *Account) GetId() int64 {
	return a.id
}

func (a *Account) SetId(id int64) {
	a.id = id
}

func (a *Account) Equal(v *Account) bool {
	return a.id == v.id
}

type Accounts []*Account

func (as *Accounts) Add(a *Account) {
	*as = append(*as, a)
}

func (as Accounts) Equal(v Accounts) bool {
	if len(as) != len(v) {
		return false
	}

	for i := range as {
		if !as[i].Equal(v[i]) {
			return false
		}
	}

	return true
}
