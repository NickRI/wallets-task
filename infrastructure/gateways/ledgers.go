package gateways

import (
	"context"
	"database/sql"

	"github.com/NickRI/wallets-task/db/models"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/domain/repositories"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

type Ledgers struct {
	createQuery *createLedgerQuery
	listQuery   *listLedgersQuery
}

func NewLedgers(d *sql.DB) (repositories.Ledgers, error) {
	var err error
	table := new(Ledgers)

	table.listQuery, err = newListLedgersQuery(d)
	if err != nil {
		return nil, xerrors.Errorf("Error preparation listLedgersQuery: %w", err)
	}

	table.createQuery, err = newCreatePaymentQuery(d)
	if err != nil {
		return nil, xerrors.Errorf("Error preparation createPaymentQuery: %w", err)
	}

	return table, nil
}

func (p *Ledgers) AddTx(tx *sql.Tx, ctx context.Context, credit *entities.Account, debit *entities.Account, amount decimal.Decimal) (*entities.Ledger, error) {
	pt := models.NewLedgerFromAccount(credit, debit, amount)

	_, err := tx.StmtContext(ctx, p.createQuery.Stmt).ExecContext(ctx, pt.Bind()...)
	if err != nil {
		return nil, models.DBErrorWrapper{err}
	}

	return pt.ToDomain(), nil
}

func (p *Ledgers) List(ctx context.Context) (entities.Ledgers, error) {
	acList := models.LedgerList{}
	rows, err := p.listQuery.QueryContext(ctx)
	if err != nil {
		return acList.ToDomain(), models.DBErrorWrapper{err}
	}
	defer rows.Close()

	for rows.Next() {
		ledger := &models.Ledger{Pays: [2]*models.Payment{&models.Payment{}, &models.Payment{}}}
		if err := rows.Scan(ledger.BindScan()...); err != nil {
			return acList.ToDomain(), models.DBErrorWrapper{err}
		}
		acList.Add(ledger)
	}

	return acList.ToDomain(), nil
}

type listLedgersQuery struct {
	*sql.Stmt
}

func newListLedgersQuery(d *sql.DB) (*listLedgersQuery, error) {
	stmt, err := d.Prepare(`SELECT p1.*, p2.*, a1.user_name, a2.user_name
		FROM payments p1
		JOIN payments p2 ON p1.guid = p2.guid AND p1.account_id != p2.account_id
		JOIN accounts a1 ON p1.account_id = a1.id
		JOIN accounts a2 ON p2.account_id = a2.id
		WHERE p1.amount > 0`,
	)
	if err != nil {
		return nil, err
	}

	return &listLedgersQuery{stmt}, nil
}

type createLedgerQuery struct {
	*sql.Stmt
}

func newCreatePaymentQuery(d *sql.DB) (*createLedgerQuery, error) {
	stmt, err := d.Prepare(`INSERT INTO payments (id, guid, account_id, amount, created_at, updated_at)
		VALUES (DEFAULT, $1, $2, $3, DEFAULT, DEFAULT), (DEFAULT, $4, $5, $6, DEFAULT, DEFAULT);
	`)
	if err != nil {
		return nil, err
	}

	return &createLedgerQuery{stmt}, err
}
