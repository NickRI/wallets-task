package models

import (
	"time"

	"github.com/NickRI/wallets-task/db/common"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type Payment struct {
	Id        int64
	Guid      uuid.UUID
	AccountId int64
	Amount    decimal.Decimal
	UpdatedAt time.Time
	CreatedAt time.Time
}

func (p *Payment) Bind() []interface{} {
	return []interface{}{
		&common.NullInt64{V: &p.Id},
		&common.NullUUID{V: &p.Guid},
		&common.NullInt64{V: &p.AccountId},
		&common.NullDecimal{V: &p.Amount},
		&common.NullTime{V: &p.UpdatedAt},
		&common.NullTime{V: &p.CreatedAt},
	}
}
