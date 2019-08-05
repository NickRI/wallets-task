package entities

import (
	"bytes"
	"strconv"

	"golang.org/x/xerrors"
)

type Direction int

const (
	Outgoing Direction = iota
	Incoming
)

func (d Direction) MarshalJSON() ([]byte, error) {
	switch d {
	case Outgoing:
		return []byte(strconv.Quote("outgoing")), nil
	case Incoming:
		return []byte(strconv.Quote("incoming")), nil
	default:
		return []byte("null"), nil
	}
}

func (d *Direction) UnmarshalJSON(v []byte) error {
	if v == nil {
		return xerrors.New("direction can't be nil")
	}

	if bytes.Compare(v[1:len(v)-1], []byte("outgoing")) == 0 {
		*d = Outgoing
		return nil
	}

	if bytes.Compare(v[1:len(v)-1], []byte("incoming")) == 0 {
		*d = Incoming
		return nil
	}

	return xerrors.New("wrong type of Direction only outgoing/incoming values allowed")
}
