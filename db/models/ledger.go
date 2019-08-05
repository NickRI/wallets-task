package models

import (
	"github.com/NickRI/wallets-task/db/common"
	"github.com/NickRI/wallets-task/domain/entities"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type Ledger struct {
	Pays  [2]*Payment
	NameA string
	NameB string
}

func NewLedgerFromAccount(credit, debit *entities.Account, amount decimal.Decimal) *Ledger {
	guid := uuid.NewV4()

	return &Ledger{
		Pays: [2]*Payment{
			{
				Guid:      guid,
				AccountId: credit.GetId(),
				Amount:    amount.Neg(),
			},
			{
				Guid:      guid,
				AccountId: debit.GetId(),
				Amount:    amount,
			},
		},
		NameA: string(credit.AccountId),
		NameB: string(debit.AccountId),
	}
}

func (l *Ledger) Bind() []interface{} {
	return []interface{}{
		&common.NullUUID{V: &l.Pays[0].Guid},
		&common.NullInt64{V: &l.Pays[0].AccountId},
		&common.NullDecimal{V: &l.Pays[0].Amount},
		&common.NullUUID{V: &l.Pays[1].Guid},
		&common.NullInt64{V: &l.Pays[1].AccountId},
		&common.NullDecimal{V: &l.Pays[1].Amount},
	}
}

func (l *Ledger) BindScan() []interface{} {
	bind := append(l.Pays[0].Bind(), l.Pays[1].Bind()...)
	return append(bind, &common.NullString{V: &l.NameA}, &common.NullString{V: &l.NameB})
}

func (l Ledger) ToDomain() *entities.Ledger {
	return &entities.Ledger{
		&entities.Payment{
			Account:   entities.AccountId(l.NameB),
			Amount:    l.Pays[1].Amount,
			Direction: entities.Outgoing,
			ToAccount: entities.AccountId(l.NameA),
		},
		&entities.Payment{
			Account:     entities.AccountId(l.NameA),
			Amount:      l.Pays[0].Amount,
			Direction:   entities.Incoming,
			FromAccount: entities.AccountId(l.NameB),
		},
	}
}

type LedgerList []*Ledger

func (ls *LedgerList) Add(l *Ledger) {
	*ls = append(*ls, l)
}

func (ls LedgerList) ToDomain() (v entities.Ledgers) {
	for _, l := range ls {
		v.Add(l.ToDomain())
	}

	return
}
