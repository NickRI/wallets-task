package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NickRI/wallets-task/db/models"
	"golang.org/x/xerrors"
)

type Response struct {
	Error errorWrapper `json:"error"`
	Data  interface{}  `json:"data"`
}

type errorWrapper struct {
	Err error
}

func (ew *errorWrapper) MarshalJSON() ([]byte, error) {
	if ew.Err != nil {
		return json.Marshal(ew.Err.Error())
	}

	return []byte("null"), nil
}

func NopDecoder(context.Context, *http.Request) (interface{}, error) {
	return nil, nil
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(&Response{Data: response})
}

func ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if xerrors.As(err, &models.NotFoundWrapper{}) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&Response{Error: errorWrapper{err}})
		return
	}

	if xerrors.As(err, &models.DBErrorWrapper{}) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Response{Error: errorWrapper{err}})
		return
	}

	if xerrors.As(err, &models.LowBalanceWrapper{}) {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(&Response{Error: errorWrapper{err}})
		return
	}

	if xerrors.As(err, &models.ValidationError{}) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&Response{Error: errorWrapper{err}})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(&Response{Error: errorWrapper{err}})
}
