package bean

import (
	"fmt"
	"log"
	"sort"
	"time"
)

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

// AccountTimeLine is a map of account events, sorted ascending by date
// well it should be...
type AccountTimeLine = map[AccountName][]AccountEvent

// NewAccountTimeLine maps the AccountEvents by AccountName
func NewAccountTimeLine(aes []AccountEvent) (AccountTimeLine, error) {
	atl := make(AccountTimeLine, len(aes))
	sort.Slice(aes, func(i, j int) bool {
		return aes[i].Date.Before(aes[j].Date)
	})
	for _, ae := range aes {
		accName := ae.Account.Name
		if atl[accName] == nil {
			atl[accName] = make([]AccountEvent, 0, 4)
		}
		atl[accName] = append(atl[accName], ae)
	}
	return atl, nil
}

func openAtDate(atl AccountTimeLine, posting Posting) bool {
	at := atl[posting.Account.Name]
	open := false
	for _, ae := range at {
		if ae.Date.After(posting.Transaction.Date) {
			break
		}
		open = ae.Open
	}
	return open
}
