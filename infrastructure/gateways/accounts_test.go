// +build !integration

package gateways

import (
	"context"
	"database/sql/driver"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/domain/repositories"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

func TestNewAccounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	newListAccountError := xerrors.New("new_list_error")
	newUpdateBalanceError := xerrors.New("new_update_balance_error")
	newFetchAccountError := xerrors.New("new_fetch_account_error")
	newLockQueryError := xerrors.New("new_lock_query_error")

	tests := []struct {
		name    string
		before  func()
		want    repositories.Accounts
		wantErr error
	}{
		{
			name: "newListAccountQuery returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts").
					WillReturnError(newListAccountError)
			},
			wantErr: newListAccountError,
		},
		{
			name: "newUpdateBalanceQuery returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .* WHERE id = .*").
					WillReturnError(newUpdateBalanceError)
			},
			wantErr: newUpdateBalanceError,
		},
		{
			name: "newFetchAccountQuery returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*").WillReturnError(newFetchAccountError)
			},
			wantErr: newFetchAccountError,
		},
		{
			name: "newLockQuery returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE").
					WillReturnError(newLockQueryError)
			},
			wantErr: newLockQueryError,
		},
		{
			name: "works well",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")
			},
			want: &Accounts{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			got, err := NewAccounts(db)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Errorf("NewAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				got.(*Accounts).listQuery = tt.want.(*Accounts).listQuery
				got.(*Accounts).balanceQuery = tt.want.(*Accounts).balanceQuery
				got.(*Accounts).fetchQuery = tt.want.(*Accounts).fetchQuery
				got.(*Accounts).lockQuery = tt.want.(*Accounts).lockQuery
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccounts() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccounts_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id", "user_name", "balance", "currency", "created_at", "updated_at"})
	var testAccount = &entities.Account{AccountId: "user_name", Balance: decimal.NewFromFloat(4.124), Currency: "USD"}

	queryContextError := xerrors.New("query_context_error")

	testAccount.SetId(1)

	testRow := []driver.Value{testAccount.GetId(), testAccount.AccountId, testAccount.Balance, testAccount.Currency, time.Now(), time.Now()}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		before  func()
		want    entities.Accounts
		wantErr error
	}{
		{
			name: "QueryContext returns error",
			args: args{ctx: context.Background()},
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")
				mock.ExpectQuery("SELECT .* FROM accounts").WillReturnError(queryContextError)
			},
			wantErr: queryContextError,
		},
		{
			name: "everything fine",
			args: args{ctx: context.Background()},
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")
				mock.ExpectQuery("SELECT .* FROM accounts").
					WillReturnRows(rows.AddRow(testRow...))

			},
			want: entities.Accounts{testAccount},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			a, err := NewAccounts(db)
			if err != nil {
				t.Fatalf("Error from NewAccounts: %+v", err)
			}
			got, err := a.List(tt.args.ctx)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("List() got = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("List() got = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}

func TestAccounts_GetByNameTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id", "user_name", "balance", "currency", "created_at", "updated_at"})
	var testAccount = &entities.Account{AccountId: "user_name", Balance: decimal.NewFromFloat(4.124), Currency: "USD"}
	testAccount.SetId(1)

	testRow := []driver.Value{testAccount.GetId(), testAccount.AccountId, testAccount.Balance, testAccount.Currency, time.Now(), time.Now()}

	queryRowContextError := xerrors.New("query_row_context_error")

	type args struct {
		ctx   context.Context
		accId entities.AccountId
	}
	tests := []struct {
		name    string
		args    args
		before  func(*args)
		want    *entities.Account
		wantErr error
	}{
		{
			name: "QueryContext returns error",
			args: args{
				ctx:   context.Background(),
				accId: "alice123",
			},
			before: func(a *args) {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")

				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .* WHERE .*").
					WithArgs(a.accId).
					WillReturnError(queryRowContextError)
			},
			wantErr: queryRowContextError,
		},
		{
			name: "working fine",
			args: args{
				ctx:   context.Background(),
				accId: "bob456",
			},
			before: func(a *args) {
				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")

				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .* WHERE .*").
					WithArgs(a.accId).
					WillReturnRows(rows.AddRow(testRow...))
			},
			want: testAccount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(&tt.args)

			a, err := NewAccounts(db)
			if err != nil {
				t.Fatalf("NewAccounts error: %+v", err)
			}

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("db.Begin error: %+v", err)
			}

			got, err := a.GetByNameTx(tx, tt.args.ctx, tt.args.accId)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Fatalf("GetByNameTx() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetByNameTx() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccounts_UpdateBalanceTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	execContextError := xerrors.New("exec_context_error")

	type args struct {
		ctx     context.Context
		account *entities.Account
		amount  decimal.Decimal
	}
	tests := []struct {
		name    string
		args    args
		before  func(*args)
		wantErr error
	}{
		{
			name: "ExecContext returns error",
			args: args{
				ctx: context.Background(),
				account: &entities.Account{
					AccountId: "alice123",
					Balance:   decimal.NewFromFloat(4.212),
					Currency:  "USD",
				},
				amount: decimal.NewFromFloat(1.212),
			},
			before: func(a *args) {
				a.account.SetId(1)

				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")

				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET .*").
					WithArgs(a.amount, a.account.GetId()).
					WillReturnError(execContextError).
					WillReturnResult(sqlmock.NewErrorResult(execContextError))
			},
			wantErr: execContextError,
		},
		{
			name: "working fine",
			args: args{
				ctx: context.Background(),
				account: &entities.Account{
					AccountId: "alice123",
					Balance:   decimal.NewFromFloat(4.212),
					Currency:  "USD",
				},
				amount: decimal.NewFromFloat(1.212),
			},
			before: func(a *args) {
				a.account.SetId(1)

				mock.ExpectPrepare("SELECT .* FROM accounts")
				mock.ExpectPrepare("UPDATE accounts SET .*")
				mock.ExpectPrepare("SELECT .* WHERE .*")
				mock.ExpectPrepare("LOCK TABLE accounts IN .* MODE")

				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET .*").
					WithArgs(a.amount, a.account.GetId()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(&tt.args)

			a, err := NewAccounts(db)
			if err != nil {
				t.Fatalf("NewAccounts error: %+v", err)
			}

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("db.Begin error: %+v", err)
			}

			err = a.UpdateBalanceTx(tx, tt.args.ctx, tt.args.account, tt.args.amount)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Errorf("UpdateBalanceTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
