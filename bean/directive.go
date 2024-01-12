package bean

import (
	"fmt"
	"log"
	"time"
)

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

// newAccountEvent creates an AccountEvent from a Directive
func newAccountEvent(directive Directive) (AccountEvent, error) {
	line := directive.Lines[0] // TODO include metadata lines
	log.Println("newAccountEvent", line.Tokens[0].Text)
	tokens := line.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return AccountEvent{}, fmt.Errorf("in newAccountEvent: %w", err)
	}
	open := tokens[1].Text == "open"
	account := tokens[2].Text
	var ccy Ccy
	if len(tokens) >= 4 {
		// TODO should return err if ccy provided on close
		ccy = Ccy(tokens[3].Text)
	}
	if len(tokens) >= 5 {
		// TODO handle additional currencies
		log.Println("ignoring extra open/close tokens")
	}
	accountEvent := AccountEvent{
		Date:    date,
		Open:    open,
		Account: Account{AccountName(account)},
		Ccy:     ccy,
	}
	return accountEvent, nil
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

// newBalance creates a Balance from a Directive
func newBalance(directive Directive) (Balance, error) {
	line := directive.Lines[0] // TODO include metadata lines
	log.Println("newBalance", line.Tokens[0].Text)
	tokens := line.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Balance{}, fmt.Errorf("in newBalance: %w", err)
	}
	account := tokens[2].Text
	numberStr := tokens[3].Text
	ccy := tokens[4].Text
	if len(tokens) > 5 {
		return Balance{}, fmt.Errorf("too many balance tokens: %s", directive)
	}

	balance := Balance{
		Date:    date,
		Account: Account{AccountName(account)},
		Amount:  MustNewAmount(numberStr, ccy),
	}
	return balance, nil
}

// Price contains an exchange rate between commodities
type Price struct {
	Date   time.Time
	Ccy    Ccy
	Amount Amount
}

func (p Price) String() string {
	return fmt.Sprintf("%s price %s %v\n", p.Date.Format(time.DateOnly), p.Ccy, p.Amount)
}

// newPrice creates a Price
func newPrice(directive Directive) (Price, error) {
	tokens := directive.Lines[0].Tokens
	log.Println("newPrice", tokens[0])
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Price{}, fmt.Errorf("in newPrice: %w", err)
	}
	ccy := tokens[2].Text
	amtNum := tokens[3].Text
	amtCcy := tokens[4].Text
	amt := MustNewAmount(amtNum, amtCcy)
	price := Price{
		Date:   date,
		Ccy:    Ccy(ccy),
		Amount: amt,
	}
	return price, nil
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

// newPad creates a Pad
func newPad(directive Directive) (Pad, error) {
	tokens := directive.Lines[0].Tokens
	log.Println("newPad", tokens[0])
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Pad{}, fmt.Errorf("in newPrice: %w", err)
	}
	padTo := Account{AccountName(tokens[2].Text)}
	padFrom := Account{AccountName(tokens[3].Text)}
	pad := Pad{
		Date:    date,
		PadTo:   padTo,
		PadFrom: padFrom,
	}
	return pad, nil
}
