package entities

import "github.com/shopspring/decimal"

type Payment struct {
	Account     AccountId       `json:"account"`
	Amount      decimal.Decimal `json:"amount"`
	ToAccount   AccountId       `json:"to_account,omitempty"`
	FromAccount AccountId       `json:"from_account,omitempty"`
	Direction   Direction       `json:"direction"`
}

func (l *Payment) Equal(v *Payment) bool {
	return l.Account == v.Account
}

type Payments []*Payment

func (ps *Payments) Add(p *Payment) {
	*ps = append(*ps, p)
}

func (ps *Payments) Append(p Payments) {
	*ps = append(*ps, p...)
}

func (ps Payments) Equal(v Payments) bool {
	if len(ps) != len(v) {
		return false
	}

	for i := range ps {
		if !ps[i].Equal(v[i]) {
			return false
		}
	}

	return true
}
