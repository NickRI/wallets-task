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

type Accounts struct {
	listQuery    *listAccountQuery
	balanceQuery *updateBalanceQuery
	fetchQuery   *fetchAccountQuery
	lockQuery    *lockQuery
}

func NewAccounts(d *sql.DB) (repositories.Accounts, error) {
	var err error
	table := &Accounts{}

	table.listQuery, err = newListAccountQuery(d)
	if err != nil {
		return nil, xerrors.Errorf("Error preparation listAccountQuery: %w", err)
	}

	table.balanceQuery, err = newUpdateBalanceQuery(d)
	if err != nil {
		return nil, xerrors.Errorf("Error preparation newUpdateBalanceQuery: %w", err)
	}

	table.fetchQuery, err = newFetchAccountQuery(d)
	if err != nil {
		return nil, xerrors.Errorf("Error preparation newFetchAccountQuery: %w", err)
	}

	table.lockQuery, err = newLockQuery(d)
	if err != nil {
		return nil, xerrors.Errorf("Error preparation newLockQuery: %w", err)
	}

	return table, nil
}

func (a *Accounts) List(ctx context.Context) (entities.Accounts, error) {
	acList := models.AccountList{}
	rows, err := a.listQuery.QueryContext(ctx)
	if err != nil {
		return nil, models.DBErrorWrapper{err}
	}
	defer rows.Close()

	for rows.Next() {
		account := &models.Account{}
		if err := rows.Scan(account.Bind()...); err != nil {
			return nil, models.DBErrorWrapper{err}
		}
		acList.Add(account)
	}

	return acList.ToDomain(), nil
}

func (a *Accounts) GetByNameTx(tx *sql.Tx, ctx context.Context, accId entities.AccountId) (*entities.Account, error) {
	row := tx.StmtContext(ctx, a.fetchQuery.Stmt).QueryRowContext(ctx, accId)

	account := &models.Account{}
	if err := row.Scan(account.Bind()...); err != nil {
		return nil, err
	}

	return account.ToDomain(), nil
}

func (a *Accounts) UpdateBalanceTx(tx *sql.Tx, ctx context.Context, account *entities.Account, amount decimal.Decimal) error {
	_, err := tx.StmtContext(ctx, a.balanceQuery.Stmt).ExecContext(ctx, amount, account.GetId())
	return err
}

func (a *Accounts) LockTx(tx *sql.Tx, ctx context.Context) error {
	_, err := tx.StmtContext(ctx, a.lockQuery.Stmt).ExecContext(ctx)
	return err
}

type updateBalanceQuery struct {
	*sql.Stmt
}

func newUpdateBalanceQuery(d *sql.DB) (*updateBalanceQuery, error) {
	stmt, err := d.Prepare("UPDATE accounts SET balance = balance + $1 WHERE id = $2")
	if err != nil {
		return nil, err
	}

	return &updateBalanceQuery{stmt}, nil
}

type listAccountQuery struct {
	*sql.Stmt
}

func newListAccountQuery(d *sql.DB) (*listAccountQuery, error) {
	stmt, err := d.Prepare("SELECT id, user_name, balance, currency, created_at, updated_at FROM accounts")
	if err != nil {
		return nil, err
	}

	return &listAccountQuery{stmt}, nil
}

type fetchAccountQuery struct {
	*sql.Stmt
}

func newFetchAccountQuery(d *sql.DB) (*fetchAccountQuery, error) {
	stmt, err := d.Prepare(`SELECT id, user_name, balance, currency, created_at, updated_at FROM accounts WHERE user_name = $1`)
	if err != nil {
		return nil, err
	}

	return &fetchAccountQuery{stmt}, nil
}

type lockQuery struct {
	*sql.Stmt
}

func newLockQuery(d *sql.DB) (*lockQuery, error) {
	stmt, err := d.Prepare(`LOCK TABLE accounts IN SHARE UPDATE EXCLUSIVE MODE`)
	if err != nil {
		return nil, err
	}

	return &lockQuery{stmt}, nil
}
