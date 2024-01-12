package bean

import (
	"fmt"
	"time"

	"github.com/cockroachdb/apd/v3"
)

// Ccy is a currency like USD or GOOG or WHATEVER
type Ccy string

// AccountName is of the form Assets:Bob:Investing:Etc
type AccountName string

// Token is raw token from input file with a bunch of flags
// quotes are removed, newlines inside quotes are maintained
type Token struct {
	LineNum int
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

// LineNum returns the source file line number of this Line
func (l Line) LineNum() int {
	if l.Blank {
		return -1
	}
	return l.Tokens[0].LineNum
}

func (l Line) String() string {
	str := fmt.Sprintf("line:%d", l.LineNum())
	for _, t := range l.Tokens {
		str += fmt.Sprint(" | ", t.Text)
	}
	return str
}

// Directive is one or more lines that go together
type Directive struct {
	Lines []Line
}

// LineNum returns the soruce file number of the
// _first line_ of this line
func (d Directive) LineNum() int {
	return d.Lines[0].LineNum()
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
	Number apd.Decimal
	Ccy    Ccy
}

// NewAmount creates an Amount
func NewAmount(num string, ccy string) (Amount, error) {
	d, _, err := apdCtx.NewFromString(num)
	if err != nil {
		return Amount{}, fmt.Errorf("in NewAmount: %w", err)
	}
	return Amount{*d, Ccy(ccy)}, nil
}

// MustNewAmount creates an Amount and panics! on underlying errors
func MustNewAmount(num string, ccy string) Amount {
	res, err := NewAmount(num, ccy)
	if err != nil {
		panic(err)
	}
	return res
}

// Eq returns true if the Amounts have same Number and Ccy
func (a Amount) Eq(other Amount) bool {
	if a.Ccy != other.Ccy {
		return false
	}

	equal := apd.New(0, 0)
	res := &apd.Decimal{}
	_, err := apdCtx.Cmp(res, &a.Number, &other.Number)
	if err != nil {
		return false
	}
	return *res == *equal
}

// Add adds other to the amount
func (a Amount) Add(other Amount) (Amount, error) {
	if a.Ccy != other.Ccy {
		return Amount{}, fmt.Errorf("cant add Amount with different currency: a: %s other: %s", a.Ccy, other.Ccy)
	}
	curVal := a.Number
	newVal := apd.Decimal{}
	apdCtx.Add(&newVal, &curVal, &other.Number)
	a.Number = newVal
	return a, nil
}

// MustAdd adds other to the Amount
func (a Amount) MustAdd(other Amount) Amount {
	res, err := a.Add(other)
	if err != nil {
		panic(err)
	}
	return res
}

// Neg returns the negated Amount
func (a Amount) Neg() Amount {
	neg := apd.Decimal{}
	apdCtx.Neg(&neg, &a.Number)
	a.Number = neg
	return a
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.Number.Text('f'), a.Ccy)
}

// Account is for now simply a string
type Account struct {
	Name AccountName
}

func (a Account) String() string {
	return string(a.Name)
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
	str := fmt.Sprintf("%s %s\n", t.Date.Format(time.DateOnly), t.Narration)
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
	return fmt.Sprintf("%s %s %s %s\n", ae.Date.Format(time.DateOnly), openOrClose, ae.Account.Name, ae.Ccy)
}

// Balance statement
type Balance struct {
	Date    time.Time
	Account Account
	Amount  Amount
}

func (b Balance) String() string {
	return fmt.Sprintf("%s balance %s %v\n", b.Date.Format(time.DateOnly), b.Account.Name, b.Amount)
}

// Ledger is the full view of the beancount file
type Ledger struct {
	AccountEvents []AccountEvent
	Balances      []Balance
	Transactions  []Transaction
	Prices        []Price
	Pads          []Pad
}

// CcyAmount is a map of Ccy -> number
type CcyAmount = map[Ccy]Amount

// NewCcyAmount converts a string map to a CcyAmount
func NewCcyAmount(bals map[string]string) (CcyAmount, error) {
	res := make(CcyAmount, len(bals))
	for ccy, num := range bals {
		val, err := NewAmount(num, ccy)
		if err != nil {
			return nil, fmt.Errorf("in NewCcyBal: %w", err)
		}
		res[Ccy(ccy)] = val
	}
	return res, nil
}

// MustNewCcyAmount converts a regular string map to a CcyAmount
func MustNewCcyAmount(bals map[string]string) CcyAmount {
	res := make(CcyAmount, len(bals))
	for ccy, num := range bals {
		res[Ccy(ccy)] = MustNewAmount(num, ccy)
	}
	return res
}

// AccBal is a map of Account -> CcyBal
type AccBal = map[AccountName]CcyAmount

// Price contains an exchange rate between commodities
type Price struct {
	Date   time.Time
	Ccy    Ccy
	Amount Amount
}

func (p Price) String() string {
	return fmt.Sprintf("%s price %s %v\n", p.Date.Format(time.DateOnly), p.Ccy, p.Amount)
}

// Pad is a pad directive
type Pad struct {
	Date    time.Time
	PadTo   Account
	PadFrom Account
}

func (p Pad) String() string {
	return fmt.Sprintf("%s pad %s %v\n", p.Date.Format(time.DateOnly), p.PadTo, p.PadFrom)
}
