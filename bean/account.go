package bean

import ()

// Ccy is a currency like USD or GOOG or WHATEVER
type Ccy string

// AccountName is of the form Assets:Bob:Investing:Etc
type AccountName string

// Account is for now simply a string
type Account struct {
	Name AccountName
}

func (a Account) String() string {
	return string(a.Name)
}
