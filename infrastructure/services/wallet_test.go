// +build !integration

package services

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/domain/services"
	"github.com/NickRI/wallets-task/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

func TestNewWalletService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	newAccountsError := xerrors.New("new_accounts_error")
	newLedgersError := xerrors.New("new_ledgers_error")

	tests := []struct {
		name    string
		before  func()
		want    services.Wallet
		wantErr error
	}{
		{
			name: "NewAccounts returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE").
					WillReturnError(newAccountsError)
			},
			wantErr: newAccountsError,
		},
		{
			name: "NewLedgers returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .* ")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")

				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)").
					WillReturnError(newLedgersError)
			},
			wantErr: newLedgersError,
		},
		{
			name: "works fine",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")

				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)")
			},
			want: &WalletService{db: db},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			got, err := NewWalletService(db)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Errorf("NewWalletService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want.(*WalletService).Accounts = got.(*WalletService).Accounts
				tt.want.(*WalletService).Ledgers = got.(*WalletService).Ledgers
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWalletService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWalletService_AccountsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccounts := mock.NewMockAccounts(ctrl)
	testError := xerrors.New("test_error")
	testAccounts := entities.Accounts{
		&entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(4.124), Currency: "USD"},
		&entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(3.124), Currency: "USD"},
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		before  func(*args)
		want    entities.Accounts
		wantErr error
	}{
		{
			name: "return error",
			args: args{ctx: context.Background()},
			before: func(a *args) {
				mockAccounts.EXPECT().List(a.ctx).Return(nil, testError)
			},
			wantErr: testError,
		},
		{
			name: "works fine",
			args: args{ctx: context.Background()},
			before: func(a *args) {
				mockAccounts.EXPECT().List(a.ctx).Return(testAccounts, nil)
			},
			want: testAccounts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WalletService{Accounts: mockAccounts}
			tt.before(&tt.args)

			got, err := w.AccountsList(tt.args.ctx)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Fatalf("AccountsList() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("AccountsList() got = %d, want %d length", len(got), len(tt.want))
			}

			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Fatalf("AccountsList() got = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWalletService_LedgersList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLedgers := mock.NewMockLedgers(ctrl)
	testError := xerrors.New("test_error")
	testAmount := decimal.NewFromFloat(4.124)
	testLedgers := entities.Ledgers{
		&entities.Ledger{
			&entities.Payment{Account: "alice123", Amount: testAmount, ToAccount: "bob456", Direction: entities.Outgoing},
			&entities.Payment{Account: "bob456", Amount: testAmount, FromAccount: "alice123", Direction: entities.Incoming},
		},
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		before  func(*args)
		args    args
		want    entities.Ledgers
		wantErr error
	}{
		{
			name: "return error",
			args: args{ctx: context.Background()},
			before: func(a *args) {
				mockLedgers.EXPECT().List(a.ctx).Return(nil, testError)
			},
			wantErr: testError,
		},
		{
			name: "works fine",
			args: args{ctx: context.Background()},
			before: func(a *args) {
				mockLedgers.EXPECT().List(a.ctx).Return(testLedgers, nil)
			},
			want: testLedgers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WalletService{Ledgers: mockLedgers}
			tt.before(&tt.args)
			got, err := w.LedgersList(tt.args.ctx)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Errorf("PaymentsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("PaymentsList() got = %d, want %d length", len(got), len(tt.want))
			}

			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Fatalf("PaymentsList() got = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWalletService_Send(t *testing.T) {
	db, dbmock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccounts := mock.NewMockAccounts(ctrl)
	mockLedgers := mock.NewMockLedgers(ctrl)

	beginError := xerrors.New("begin_tx_error")

	getByNameTxError := xerrors.New("get_by_name_error")
	getByNameTxError2 := xerrors.New("get_by_name_error_2")

	lockError := xerrors.New("lock_error")

	updateBalanceError := xerrors.New("update_balance_error")
	updateBalanceError2 := xerrors.New("update_balance_error_2")

	addLedgerError := xerrors.New("add_ledger_error")
	commitError := &pq.Error{}

	type args struct {
		ctx      context.Context
		creditId string
		debitId  string
		amount   decimal.Decimal
	}
	tests := []struct {
		name    string
		args    args
		before  func(*args)
		want    *entities.Ledger
		wantErr error
	}{
		{
			name: "BeginTx returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin().WillReturnError(beginError)
			},
			wantErr: beginError,
		},
		{
			name: "creditId not found",
			args: args{
				ctx:      context.Background(),
				creditId: "alice1234",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(nil, sql.ErrNoRows)

				dbmock.ExpectRollback()
			},
			wantErr: xerrors.Errorf("account %s not found", "alice1234"),
		},
		{
			name: "get creditId returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(nil, getByNameTxError)

				dbmock.ExpectRollback()
			},
			wantErr: getByNameTxError,
		},
		{
			name: "credit's balance less than amount",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).
					Return(&entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(1.21), Currency: "USD"}, nil)

				dbmock.ExpectRollback()
			},
			wantErr: xerrors.Errorf("%s: don't have enough balance", "alice123"),
		},
		{
			name: "debitId not found",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob4567",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).
					Return(&entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).
					Return(nil, sql.ErrNoRows)

				dbmock.ExpectRollback()
			},
			wantErr: xerrors.Errorf("account %s not found", "bob4567"),
		},
		{
			name: "get debitId returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).
					Return(&entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).
					Return(nil, getByNameTxError2)

				dbmock.ExpectRollback()
			},
			wantErr: getByNameTxError2,
		},
		{
			name: "currencies is not equal",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).
					Return(&entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).
					Return(&entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(10.21), Currency: "EUR"}, nil)

				dbmock.ExpectRollback()
			},
			wantErr: xerrors.New("currencies for accounts is not equal"),
		},
		{
			name: "account table lock returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				credit := &entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}
				debit := &entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(12.21), Currency: "USD"}

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(credit, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).Return(debit, nil)

				mockAccounts.EXPECT().LockTx(gomock.Any(), a.ctx).Return(lockError)

				dbmock.ExpectRollback()
			},
			wantErr: lockError,
		},
		{
			name: "credit's balance update returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				credit := &entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}
				debit := &entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(12.21), Currency: "USD"}

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(credit, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).Return(debit, nil)

				mockAccounts.EXPECT().LockTx(gomock.Any(), a.ctx).Return(nil)

				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, credit, a.amount.Neg()).Return(updateBalanceError)

				dbmock.ExpectRollback()
			},
			wantErr: updateBalanceError,
		},
		{
			name: "debits's balance update returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				credit := &entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}
				debit := &entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(12.21), Currency: "USD"}

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(credit, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).Return(debit, nil)

				mockAccounts.EXPECT().LockTx(gomock.Any(), a.ctx).Return(nil)

				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, credit, a.amount.Neg()).Return(nil)
				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, debit, a.amount).Return(updateBalanceError2)

				dbmock.ExpectRollback()
			},
			wantErr: updateBalanceError2,
		},
		{
			name: "add ledger returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				credit := &entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}
				debit := &entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(12.21), Currency: "USD"}

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(credit, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).Return(debit, nil)

				mockAccounts.EXPECT().LockTx(gomock.Any(), a.ctx).Return(nil)

				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, credit, a.amount.Neg()).Return(nil)
				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, debit, a.amount).Return(nil)

				mockLedgers.EXPECT().AddTx(gomock.Any(), a.ctx, credit, debit, a.amount).Return(nil, addLedgerError)

				dbmock.ExpectRollback()

			},
			wantErr: addLedgerError,
		},
		{
			name: "commit returns error",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				credit := &entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}
				debit := &entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(12.21), Currency: "USD"}

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(credit, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).Return(debit, nil)

				mockAccounts.EXPECT().LockTx(gomock.Any(), a.ctx).Return(nil)

				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, credit, a.amount.Neg()).Return(nil)
				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, debit, a.amount).Return(nil)

				mockLedgers.EXPECT().AddTx(gomock.Any(), a.ctx, credit, debit, a.amount).Return(nil, nil)

				dbmock.ExpectCommit().WillReturnError(commitError)
			},
			wantErr: commitError,
		},
		{
			name: "commit works fine",
			args: args{
				ctx:      context.Background(),
				creditId: "alice123",
				debitId:  "bob456",
				amount:   decimal.NewFromFloat(3.21),
			},
			before: func(a *args) {
				dbmock.ExpectBegin()

				credit := &entities.Account{AccountId: "alice123", Balance: decimal.NewFromFloat(10.21), Currency: "USD"}
				debit := &entities.Account{AccountId: "bob456", Balance: decimal.NewFromFloat(12.21), Currency: "USD"}

				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.creditId)).Return(credit, nil)
				mockAccounts.EXPECT().GetByNameTx(gomock.Any(), a.ctx, entities.AccountId(a.debitId)).Return(debit, nil)

				mockAccounts.EXPECT().LockTx(gomock.Any(), a.ctx).Return(nil)

				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, credit, a.amount.Neg()).Return(nil)
				mockAccounts.EXPECT().UpdateBalanceTx(gomock.Any(), a.ctx, debit, a.amount).Return(nil)

				mockLedgers.EXPECT().AddTx(gomock.Any(), a.ctx, credit, debit, a.amount).Return(nil, nil)

				dbmock.ExpectCommit()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WalletService{
				db:       db,
				Accounts: mockAccounts,
				Ledgers:  mockLedgers,
			}
			tt.before(&tt.args)
			got, err := w.Send(tt.args.ctx, tt.args.creditId, tt.args.debitId, tt.args.amount)
			if err != nil && (!xerrors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error()) || tt.wantErr != nil && err == nil {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("PaymentsList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
