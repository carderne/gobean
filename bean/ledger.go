package bean

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/cockroachdb/apd/v3"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(io.Discard)
}

// debug indicates whether DEBUG env var is set to 1
var debug bool

// apd Decimal context
var apdCtx = apd.BaseContext

// Ledger is the full view of the beancount file
type Ledger struct {
	AccountEvents   []AccountEvent
	AccountTimeLine AccountTimeLine
	Balances        []Balance
	Transactions    []Transaction
	Postings        []Posting
	Prices          []Price
	Pads            []Pad
}

// NewLedger parses the supplied file and creates a ledger
// ready for calculations
func NewLedger(dbg bool) *Ledger {
	debug = dbg
	if debug {
		log.SetOutput(os.Stderr)
	}

	l := Ledger{}
	return &l
}

// Load a beancount file/string into the Ledger
func (l *Ledger) Load(rc io.ReadCloser) (*Ledger, error) {
	var err error
	l, err = l.parse(rc)
	if err != nil {
		return l, fmt.Errorf("in GetBalances: %w", err)
	}
	l.Transactions, err = balanceTransactions(l.Transactions)
	if err != nil {
		return l, fmt.Errorf("in GetBalances: %w", err)
	}
	debugSlice(l.Transactions, "ledger.Transactions")

	// extractPostings never errors currently
	postings, _ := extractPostings(l.Transactions)
	postings, _ = sortPostings(postings)
	l.Postings = postings
	debugSlice(l.Postings, "ledger.Postings")

	accountTimeLine, err := NewAccountTimeLine(l.AccountEvents)
	l.AccountTimeLine = accountTimeLine

	return l, nil
}

// newLedger creates the basic ledger with
// accountEvents (open/close), balance directives and transactions.
// These are not yet logically validated, only checked semantically
func (l *Ledger) fill(directives []Directive) (*Ledger, error) {
	var accountEvents []AccountEvent
	var balances []Balance
	var transactions []Transaction
	var prices []Price
	var pads []Pad

	for _, directive := range directives {
		if len(directive.Lines) == 0 {
			continue
		}
		switch typeStr := dirType(directive.Lines[0].Tokens[1].Text); typeStr {
		case dirBalance:
			d, err := newBalance(directive)
			if err != nil {
				return l, fmt.Errorf("in newLedger: %w", err)
			}
			balances = append(balances, d)
		case dirOpen, dirClose:
			d, err := newAccountEvent(directive)
			if err != nil {
				return l, fmt.Errorf("in newLedger: %w", err)
			}
			accountEvents = append(accountEvents, d)
		case dirTxn, dirStar, dirBang:
			d, err := newTransaction(directive)
			if err != nil {
				return l, fmt.Errorf("in newLedger: %w", err)
			}
			transactions = append(transactions, d)
		case dirPrice:
			d, err := newPrice(directive)
			if err != nil {
				return l, fmt.Errorf("in newLedger: %w", err)
			}
			prices = append(prices, d)
		case dirPad:
			d, err := newPad(directive)
			if err != nil {
				return l, fmt.Errorf("in newLedger: %w", err)
			}
			pads = append(pads, d)
		case dirNote, dirCommodity, dirQuery, dirCustom:
		default:
			return l, fmt.Errorf("in NewLedger: found unrecognised directive: %s", typeStr)
		}
	}
	debugSlice(transactions, "transactions")
	debugSlice(accountEvents, "accountEvents")
	debugSlice(balances, "balances")
	debugSlice(prices, "prices")
	debugSlice(pads, "pads")
	l.AccountEvents = accountEvents
	l.Balances = balances
	l.Transactions = transactions
	l.Prices = prices
	l.Pads = pads
	return l, nil
}

// parse does all the work of loading a file path
// and returning a Ledger
func (l *Ledger) parse(rc io.ReadCloser) (*Ledger, error) {
	// getTokens never errors currently
	tokens, _ := getTokens(rc)
	// makeLines never errors currently
	lines, _ := makeLines(tokens)
	debugSlice(lines, "lines")
	directives, err := makeDirectives(lines)
	if err != nil {
		return &Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	debugSlice(directives, "directives")

	l, err = l.fill(directives)
	if err != nil {
		return l, fmt.Errorf("in parse: %w", err)
	}
	return l, nil
}

// GetBalances returns the final balance of
// all accounts, separately for each currency
func (l *Ledger) GetBalances(date time.Time) (AccBal, error) {
	accBalances, err := getBalances(l.Postings, l.AccountTimeLine, date)
	return accBalances, err
}
