package models

type DBErrorWrapper struct {
	Err error
}

func (dbw DBErrorWrapper) Error() string {
	return dbw.Err.Error()
}

func (dbw DBErrorWrapper) Unwrap() error {
	return dbw.Err
}

type LowBalanceWrapper struct {
	Err error
}

func (lbw LowBalanceWrapper) Error() string {
	return lbw.Err.Error()
}

func (lbw LowBalanceWrapper) Unwrap() error {
	return lbw.Err
}

type NotFoundWrapper struct {
	Err error
}

func (nfe NotFoundWrapper) Error() string {
	return nfe.Err.Error()
}

func (nfe NotFoundWrapper) Unwrap() error {
	return nfe.Err
}

type ValidationError struct {
	Err error
}

func (ve ValidationError) Error() string {
	return ve.Err.Error()
}

func (ve ValidationError) Unwrap() error {
	return ve.Err
}
