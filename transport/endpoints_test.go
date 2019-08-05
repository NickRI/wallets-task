// +build !integration

package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/NickRI/wallets-task/db/models"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/internal/mock"
	"github.com/NickRI/wallets-task/transport/endpoints"
	"github.com/NickRI/wallets-task/transport/restapi"
	"github.com/go-chi/chi"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"

	"net/http/httptest"
	"testing"
)

func Test_AccountListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWallet := mock.NewMockWallet(ctrl)

	testError := xerrors.New("some_error")

	tests := []struct {
		name     string
		before   func(entities.Accounts)
		want     entities.Accounts
		wantCode int
		wantErr  string
	}{
		{
			name: "wallet returns error",
			before: func(want entities.Accounts) {
				mockWallet.EXPECT().AccountsList(gomock.Any()).
					Return(want, testError)
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  testError.Error(),
		},
		{
			name: "wallet returns db-error",
			before: func(want entities.Accounts) {
				mockWallet.EXPECT().AccountsList(gomock.Any()).
					Return(want, models.DBErrorWrapper{testError})
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  testError.Error(),
		},
		{
			name: "wallet returns accounts normally",
			before: func(want entities.Accounts) {
				mockWallet.EXPECT().AccountsList(gomock.Any()).
					Return(want, nil)
			},
			wantCode: http.StatusOK,
			want: entities.Accounts{
				&entities.Account{
					AccountId: "alice123",
					Balance:   decimal.NewFromFloat(2.54),
					Currency:  "USD",
				},
				&entities.Account{
					AccountId: "bob456",
					Balance:   decimal.NewFromFloat(10.54),
					Currency:  "USD",
				},
			},
		},
	}

	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(endpoints.ErrorEncoder),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.want)
			h := restapi.MakeHandlers(mockWallet, options...)

			req := httptest.NewRequest("GET", "restapi://localhost/wallet/accounts", nil)
			w := httptest.NewRecorder()

			h.ListAccounts.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != tt.wantCode {
				t.Fatalf("ListAccountHandler() StatusCode = %v, wantCode = %v", resp.StatusCode, tt.wantCode)
			}

			respBody := struct {
				Err  string            `json:"error"`
				Data entities.Accounts `json:"data"`
			}{}

			if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
				t.Fatal(err)
			}

			if respBody.Err != tt.wantErr {
				t.Fatalf("ListAccountHandler() error = %v, wantErr = %v", respBody.Err, tt.wantErr)
			}

			if len(respBody.Data) != len(tt.want) {
				t.Fatalf("ListAccountHandler() got = %v, want %v length", len(respBody.Data), len(tt.want))

			}

			for i := range respBody.Data {
				if !reflect.DeepEqual(respBody.Data[i], tt.want[i]) {
					t.Fatalf("ListAccountHandler() got = %v, want %v", respBody.Data[i], tt.want[i])
				}
			}
		})
	}
}

func Test_LedgerListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWallet := mock.NewMockWallet(ctrl)

	testError := xerrors.New("some_error")

	tests := []struct {
		name     string
		before   func(entities.Ledgers)
		want     entities.Ledgers
		wantCode int
		wantErr  string
	}{
		{
			name: "wallet returns error",
			before: func(want entities.Ledgers) {
				mockWallet.EXPECT().LedgersList(gomock.Any()).
					Return(want, testError)
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  testError.Error(),
		},
		{
			name: "wallet returns db-error",
			before: func(want entities.Ledgers) {
				mockWallet.EXPECT().LedgersList(gomock.Any()).
					Return(want, models.DBErrorWrapper{testError})
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  testError.Error(),
		},
		{
			name: "wallet returns ledgers normally",
			before: func(want entities.Ledgers) {
				mockWallet.EXPECT().LedgersList(gomock.Any()).
					Return(want, nil)
			},
			wantCode: http.StatusOK,
			want: entities.Ledgers{
				&entities.Ledger{
					&entities.Payment{
						Account:   "alice123",
						Amount:    decimal.NewFromFloat(2.54),
						ToAccount: "bob456",
						Direction: entities.Outgoing,
					},
					&entities.Payment{
						Account:     "bob456",
						Amount:      decimal.NewFromFloat(2.54),
						FromAccount: "alice123",
						Direction:   entities.Incoming,
					},
				},
				&entities.Ledger{
					&entities.Payment{
						Account:   "bob456",
						Amount:    decimal.NewFromFloat(22.54),
						ToAccount: "alice123",
						Direction: entities.Outgoing,
					},
					&entities.Payment{
						Account:     "alice123",
						Amount:      decimal.NewFromFloat(22.54),
						FromAccount: "bob456",
						Direction:   entities.Incoming,
					},
				},
			},
		},
	}

	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(endpoints.ErrorEncoder),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.want)
			h := restapi.MakeHandlers(mockWallet, options...)

			req := httptest.NewRequest("GET", "restapi://localhost/wallet/ledgers", nil)
			w := httptest.NewRecorder()

			h.ListLedgers.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != tt.wantCode {
				t.Fatalf("LedgerListHandler() StatusCode = %v, wantCode = %v", resp.StatusCode, tt.wantCode)
			}

			respBody := struct {
				Err  string           `json:"error"`
				Data entities.Ledgers `json:"data"`
			}{}

			if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
				t.Fatal(err)
			}

			if respBody.Err != tt.wantErr {
				t.Fatalf("LedgerListHandler() error = %v, wantErr = %v", respBody.Err, tt.wantErr)
			}

			if len(respBody.Data) != len(tt.want) {
				t.Fatalf("LedgerListHandler() got = %v, want %v length", len(respBody.Data), len(tt.want))

			}

			for i := range respBody.Data {
				if !reflect.DeepEqual(respBody.Data[i], tt.want[i]) {
					t.Fatalf("LedgerListHandler() got = %v, want %v", respBody.Data[i], tt.want[i])
				}
			}
		})
	}
}

func Test_PaymentSendHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWallet := mock.NewMockWallet(ctrl)

	testDbError := xerrors.New("some_db_error")
	testNotFoundError := xerrors.New("some_error_not_found")
	testLowBalanceError := xerrors.New("some_low_balance_error")
	testSomeError := xerrors.New("some_error")

	type args struct {
		from   string
		to     string
		amount string
	}

	tests := []struct {
		name     string
		args     args
		before   func(args, *entities.Ledger)
		want     entities.Ledger
		wantCode int
		wantErr  string
	}{
		{
			name: "wallet returns wrong body",
			args: args{
				from:   "alice123",
				to:     "bob456",
				amount: `"sfefwe"`,
			},
			before:   func(a args, l *entities.Ledger) {},
			wantCode: http.StatusBadRequest,
			wantErr:  "error while json decoding: Error decoding string 'sfefwe': can't convert sfefwe to decimal: exponent is not numeric",
		},
		{
			name: "wallet returns db-error",
			args: args{
				from:   "alice123",
				to:     "bob456",
				amount: "13.5455",
			},
			before: func(a args, l *entities.Ledger) {
				mockWallet.EXPECT().Send(gomock.Any(), a.from, a.to, decimal.RequireFromString(a.amount)).
					Return(l, models.DBErrorWrapper{testDbError})
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  testDbError.Error(),
		},
		{
			name: "wallet returns not-found",
			args: args{
				from:   "alice123",
				to:     "bob456",
				amount: "7.395",
			},
			before: func(a args, l *entities.Ledger) {
				mockWallet.EXPECT().Send(gomock.Any(), a.from, a.to, decimal.RequireFromString(a.amount)).
					Return(l, models.NotFoundWrapper{testNotFoundError})
			},
			wantCode: http.StatusNotFound,
			wantErr:  testNotFoundError.Error(),
		},
		{
			name: "wallet returns low balance",
			args: args{
				from:   "alice123",
				to:     "bob456",
				amount: "7.395",
			},
			before: func(a args, l *entities.Ledger) {
				mockWallet.EXPECT().Send(gomock.Any(), a.from, a.to, decimal.RequireFromString(a.amount)).
					Return(l, models.LowBalanceWrapper{testLowBalanceError})
			},
			wantCode: http.StatusPaymentRequired,
			wantErr:  testLowBalanceError.Error(),
		},
		{
			name: "wallet returns some error",
			args: args{
				from:   "bob456",
				to:     "alice123",
				amount: "7.395",
			},
			before: func(a args, l *entities.Ledger) {
				mockWallet.EXPECT().Send(gomock.Any(), a.from, a.to, decimal.RequireFromString(a.amount)).
					Return(l, testSomeError)
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  testSomeError.Error(),
		},
		{
			name: "wallet run payment normally",
			args: args{
				from:   "bob456",
				to:     "alice123",
				amount: "17.395",
			},
			before: func(a args, l *entities.Ledger) {
				mockWallet.EXPECT().Send(gomock.Any(), a.from, a.to, decimal.RequireFromString(a.amount)).
					Return(l, nil)
			},
			wantCode: http.StatusOK,
			want: entities.Ledger{
				&entities.Payment{
					Account:   "alice123",
					Amount:    decimal.NewFromFloat(2.54),
					ToAccount: "bob456",
					Direction: entities.Outgoing,
				},
				&entities.Payment{
					Account:     "bob456",
					Amount:      decimal.NewFromFloat(2.54),
					FromAccount: "alice123",
					Direction:   entities.Incoming,
				},
			},
		},
	}

	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(endpoints.ErrorEncoder),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.args, &tt.want)
			h := restapi.MakeHandlers(mockWallet, options...)

			body := bytes.NewBufferString(`{"amount": ` + tt.args.amount + `}`)

			req := httptest.NewRequest("POST", "restapi://localhost/"+tt.args.from+"/"+tt.args.to, body)
			w := httptest.NewRecorder()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"sender", "receiver"},
					Values: []string{tt.args.from, tt.args.to},
				},
			}))

			h.Send.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != tt.wantCode {
				t.Fatalf("PaymentSendHandler() StatusCode = %v, wantCode = %v", resp.StatusCode, tt.wantCode)
			}

			respBody := struct {
				Err  string          `json:"error"`
				Data entities.Ledger `json:"data"`
			}{}

			if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
				t.Fatal(err)
			}

			if respBody.Err != tt.wantErr {
				t.Fatalf("PaymentSendHandler() error = %v, wantErr = %v", respBody.Err, tt.wantErr)
			}

			if len(respBody.Data) != len(tt.want) {
				t.Fatalf("PaymentSendHandler() got = %v, want %v length", len(respBody.Data), len(tt.want))
			}

			for i := range respBody.Data {
				if !reflect.DeepEqual(respBody.Data[i], tt.want[i]) {
					t.Fatalf("PaymentSendHandler() got = %v, want %v", respBody.Data[i], tt.want[i])
				}
			}
		})
	}
}
