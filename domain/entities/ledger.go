package entities

type Ledger [2]*Payment

type Ledgers []*Ledger

func (ls *Ledgers) Add(l *Ledger) {
	*ls = append(*ls, l)
}

func (ls *Ledgers) Append(v Ledgers) {
	*ls = append(*ls, v...)
}
