// +build !integration

package gateways

import (
	"context"
	"database/sql/driver"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NickRI/wallets-task/db/models"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/domain/repositories"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

func TestNewLedgers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	newListLedgersError := xerrors.New("new_list_ledgers_error")
	newCreatePaymentError := xerrors.New("new_create_payment_error")

	tests := []struct {
		name    string
		want    repositories.Ledgers
		before  func()
		wantErr error
	}{
		{
			name: "newListLedgersQuery returns error",
			before: func() {
				mock.ExpectPrepare(`SELECT .* FROM payments .* WHERE .*`).
					WillReturnError(newListLedgersError)
			},
			wantErr: newListLedgersError,
		},
		{
			name: "newCreatePaymentQuery returns error",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)").
					WillReturnError(newCreatePaymentError)
			},
			wantErr: newCreatePaymentError,
		},
		{
			name: "works well",
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)")
			},
			want: &Ledgers{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			got, err := NewLedgers(db)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Errorf("NewPayments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				got.(*Ledgers).listQuery = tt.want.(*Ledgers).listQuery
				got.(*Ledgers).createQuery = tt.want.(*Ledgers).createQuery
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPayments() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLedgers_AddTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	execContextError := xerrors.New("exec_context_error")

	type args struct {
		ctx    context.Context
		credit *entities.Account
		debit  *entities.Account
		amount decimal.Decimal
	}
	tests := []struct {
		name    string
		args    args
		before  func(*args, *entities.Ledger)
		want    *entities.Ledger
		wantErr error
	}{
		{
			name: "ExecContext returns error",
			args: args{
				ctx: context.Background(),
				credit: &entities.Account{
					AccountId: "alice123",
					Balance:   decimal.NewFromFloat(4.212),
					Currency:  "USD",
				},
				debit: &entities.Account{
					AccountId: "bob456",
					Balance:   decimal.NewFromFloat(1.212),
					Currency:  "USD",
				},
				amount: decimal.NewFromFloat(1.212),
			},
			before: func(a *args, l *entities.Ledger) {
				a.credit.SetId(1)
				a.debit.SetId(1)

				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)")

				pt := models.NewLedgerFromAccount(a.credit, a.debit, a.amount)

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO payments (.*) VALUES (.*), (.*)").
					WithArgs(
						sqlmock.AnyArg(), pt.Pays[0].AccountId, pt.Pays[0].Amount,
						sqlmock.AnyArg(), pt.Pays[1].AccountId, pt.Pays[1].Amount,
					).
					WillReturnError(execContextError).
					WillReturnResult(sqlmock.NewErrorResult(execContextError))
			},
			wantErr: execContextError,
		},
		{
			name: "works well",
			args: args{
				ctx: context.Background(),
				credit: &entities.Account{
					AccountId: "alice123",
					Balance:   decimal.NewFromFloat(4.212),
					Currency:  "USD",
				},
				debit: &entities.Account{
					AccountId: "bob456",
					Balance:   decimal.NewFromFloat(1.212),
					Currency:  "USD",
				},
				amount: decimal.NewFromFloat(1.212),
			},
			before: func(a *args, l *entities.Ledger) {
				a.credit.SetId(1)
				a.debit.SetId(1)

				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)")

				pt := models.NewLedgerFromAccount(a.credit, a.debit, a.amount)
				*l = *pt.ToDomain()

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO payments (.*) VALUES (.*), (.*)").
					WithArgs(
						sqlmock.AnyArg(), pt.Pays[0].AccountId, pt.Pays[0].Amount,
						sqlmock.AnyArg(), pt.Pays[1].AccountId, pt.Pays[1].Amount,
					).
					WillReturnResult(sqlmock.NewResult(1, 2))
			},
			want: &entities.Ledger{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(&tt.args, tt.want)

			p, err := NewLedgers(db)
			if err != nil {
				t.Fatalf("NewAccounts error: %+v", err)
			}

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("db.Begin error: %+v", err)
			}

			got, err := p.AddTx(tx, tt.args.ctx, tt.args.credit, tt.args.debit, tt.args.amount)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Fatalf("AddTx() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("AddTx() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLedgers_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	queryContextError := xerrors.New("query_context_error")

	var rows = sqlmock.NewRows([]string{"p1.id", "p1.guid", "p1.account_id", "p1.amount", "p1.created_at", "p1.updated_at",
		"p2.id", "p2.guid", "p2.account_id", "p2.amount", "p2.created_at", "p2.updated_at",
		"a1.user_name", "a2.user_name"})

	testAmount := decimal.NewFromFloat(4.124)

	outgoingPayment := &entities.Payment{
		Account:   "alice123",
		Amount:    testAmount,
		ToAccount: "bob456",
		Direction: entities.Outgoing,
	}

	incomingPayment := &entities.Payment{
		Account:     "bob456",
		Amount:      testAmount,
		FromAccount: "alice123",
		Direction:   entities.Incoming,
	}

	guidBytes := uuid.FromStringOrNil("c5417ca1-c06b-4a45-9cd9-85936d4b9665").Bytes()

	testRow := []driver.Value{1, guidBytes, 1, testAmount, time.Now(), time.Now(),
		2, guidBytes, 1, testAmount, time.Now(), time.Now(),
		"bob456", "alice123",
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		before  func()
		want    entities.Ledgers
		wantErr error
	}{
		{
			name: "QueryContext returns error",
			args: args{ctx: context.Background()},
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)")

				mock.ExpectQuery("SELECT .* FROM payments .* WHERE .*").
					WillReturnError(queryContextError).
					RowsWillBeClosed()
			},
			wantErr: queryContextError,
		},
		{
			name: "everything fine",
			args: args{ctx: context.Background()},
			before: func() {
				mock.ExpectPrepare("SELECT .* FROM payments .* WHERE .*")
				mock.ExpectPrepare("INSERT INTO payments (.*) VALUES (.*), (.*)")

				mock.ExpectQuery("SELECT .* FROM payments .* WHERE .*").
					WillReturnRows(rows.AddRow(testRow...)).
					RowsWillBeClosed()
			},
			want: entities.Ledgers{&entities.Ledger{outgoingPayment, incomingPayment}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()

			p, err := NewLedgers(db)
			if err != nil {
				t.Fatalf("NewAccounts error: %+v", err)
			}

			got, err := p.List(tt.args.ctx)
			if err != nil && !xerrors.Is(err, tt.wantErr) || tt.wantErr != nil && err == nil {
				t.Fatalf("List() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("List() got = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Fatalf("List() got = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}
