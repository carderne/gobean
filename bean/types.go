package bean

import (
	"fmt"
	"time"
)

// Ccy is a currency like USD or GOOG or WHATEVER
type Ccy string

// AccountName is of the form Assets:Bob:Investing:Etc
type AccountName string

// Token is raw token from input file with a bunch of flags
// quotes are removed, newlines inside quotes are maintained
type Token struct {
	Indent  bool
	Quote   bool
	Comment bool
	EOL     bool
	Text    string
}

// A Line from the beancount file
type Line struct {
	Blank  bool
	Tokens []Token
}

func (l Line) String() string {
	str := ""
	for _, t := range l.Tokens {
		str += fmt.Sprint(" | ", t.Text)
	}
	return str
}

// Directive is one or more lines that go together
type Directive struct {
	Lines []Line
}

func (d Directive) String() string {
	str := ""
	for _, l := range d.Lines {
		str += fmt.Sprint(l, "\n")
	}
	str += "\n"
	return str
}

// Amount is a number with a currency
type Amount struct {
	Number float64
	Ccy    Ccy
}

func (a Amount) String() string {
	return fmt.Sprintf("%.2f %s", a.Number, a.Ccy)
}

// Account is for now simply a string
type Account struct {
	Name AccountName
}

// Posting is an individual leg of a transaction
type Posting struct {
	Account     Account
	Amount      *Amount      // to allow nil
	Transaction *Transaction // nil until exctractPostings is run
}

func (p Posting) String() string {
	var amountStr string
	if p.Amount != nil {
		amountStr = fmt.Sprintf("%v", p.Amount)
	}
	return fmt.Sprintf("%v: %v", p.Account.Name, amountStr)
}

// Transaction must have at least two postings
type Transaction struct {
	Date      time.Time
	Type      string
	Payee     string
	Narration string
	Postings  []Posting
}

func (t Transaction) String() string {
	str := fmt.Sprintf("%s %s\n", t.Date.Format(dateLayout), t.Narration)
	for _, p := range t.Postings {
		str += fmt.Sprintf("  %v\n", p)
	}
	return str
}

// AccountEvent is opening/closing accounts
type AccountEvent struct {
	Date    time.Time
	Open    bool
	Account Account
	Ccy     Ccy
}

func (ae AccountEvent) String() string {
	openOrClose := "close"
	if ae.Open {
		openOrClose = "open"
	}
	return fmt.Sprintf("%s %s %s %s\n", ae.Date.Format(dateLayout), openOrClose, ae.Account.Name, ae.Ccy)
}

// Balance statement
type Balance struct {
	Date    time.Time
	Account Account
	Amount  Amount
}

func (b Balance) String() string {
	return fmt.Sprintf("%s balance %s %v\n", b.Date.Format(dateLayout), b.Account.Name, b.Amount)
}

// Ledger is the full view of the beancount file
type Ledger struct {
	AccountEvents []AccountEvent
	Balances      []Balance
	Transactions  []Transaction
}

// CcyBal is a map of Ccy -> number
type CcyBal = map[Ccy]float64

// AccBal is a map of Account -> CcyBal
type AccBal = map[AccountName]CcyBal
