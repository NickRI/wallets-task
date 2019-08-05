package services

import (
	"context"
	"database/sql"

	"github.com/NickRI/wallets-task/db/models"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/domain/repositories"
	"github.com/NickRI/wallets-task/domain/services"
	"github.com/NickRI/wallets-task/infrastructure/gateways"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

type WalletService struct {
	db       *sql.DB
	Accounts repositories.Accounts
	Ledgers  repositories.Ledgers
}

func NewWalletService(d *sql.DB) (services.Wallet, error) {

	AccountTable, err := gateways.NewAccounts(d)
	if err != nil {
		return nil, xerrors.Errorf("Error in accounts table: %w", err)
	}

	LedgersTable, err := gateways.NewLedgers(d)
	if err != nil {
		return nil, xerrors.Errorf("Error in payments table: %w", err)
	}

	return &WalletService{db: d, Accounts: AccountTable, Ledgers: LedgersTable}, nil
}

func (w *WalletService) LedgersList(ctx context.Context) (entities.Ledgers, error) {
	return w.Ledgers.List(ctx)
}

func (w *WalletService) AccountsList(ctx context.Context) (entities.Accounts, error) {
	return w.Accounts.List(ctx)
}

func (w *WalletService) Send(ctx context.Context, creditId, debitId string, amount decimal.Decimal) (l *entities.Ledger, err error) {
	for {
		l, err = w.trySend(ctx, creditId, debitId, amount)
		if err != nil {
			var pgError *pq.Error
			if xerrors.As(err, &pgError) {
				tpe := xerrors.Unwrap(err).(*pq.Error)
				if tpe.Code == "40001" {
					continue
				}
			}
		}
		break
	}

	return
}

func (w *WalletService) trySend(ctx context.Context, creditId, debitId string, amount decimal.Decimal) (l *entities.Ledger, err error) {
	var credit, debit *entities.Account

	tx, err := w.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return l, xerrors.Errorf("begin transaction error: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Commit(); err != nil {
			err = xerrors.Errorf("error during commit: %w", err)
		}
	}()

	credit, err = w.Accounts.GetByNameTx(tx, ctx, entities.AccountId(creditId))
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			err = models.NotFoundWrapper{xerrors.Errorf("account %s not found", creditId)}
		} else {
			err = xerrors.Errorf("get by name error: %w", err)
		}

		return
	}

	if credit.Balance.LessThan(amount) {
		err = models.LowBalanceWrapper{xerrors.Errorf("%s: don't have enough balance", credit.AccountId)}
		return
	}

	debit, err = w.Accounts.GetByNameTx(tx, ctx, entities.AccountId(debitId))
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			err = models.NotFoundWrapper{xerrors.Errorf("account %s not found", debitId)}
		} else {
			err = xerrors.Errorf("get by name error: %w", err)
		}
		return
	}

	if credit.Currency != debit.Currency {
		err = xerrors.New("currencies for accounts is not equal")
		return
	}

	if err = w.Accounts.LockTx(tx, ctx); err != nil {
		err = xerrors.Errorf("lock accounts table error: %w", err)
		return
	}

	if err = w.Accounts.UpdateBalanceTx(tx, ctx, credit, amount.Neg()); err != nil {
		err = xerrors.Errorf("%s: error during decrease balance: %w", credit.AccountId, err)
		return
	}

	if err = w.Accounts.UpdateBalanceTx(tx, ctx, debit, amount); err != nil {
		err = xerrors.Errorf("%s: error during increase balance: %w", credit.AccountId, err)
		return
	}

	l, err = w.Ledgers.AddTx(tx, ctx, credit, debit, amount)
	if err != nil {
		err = xerrors.Errorf("error during add ledger: %w", err)
	}
	return
}
