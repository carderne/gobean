package bean

import (
	"fmt"
	"io"
	"log"
	"os"

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
	AccountEvents []AccountEvent
	Balances      []Balance
	Transactions  []Transaction
	Prices        []Price
	Pads          []Pad
}

// EmptyLedger returns an empty Ledger
func EmptyLedger(dbg bool) *Ledger {
	debug = dbg
	if debug {
		log.SetOutput(os.Stderr)
	}
	return &Ledger{}
}

// newLedger creates the basic ledger with
// accountEvents (open/close), balance directives and transactions.
// These are not yet logically validated, only checked semantically
func (l *Ledger) fill(directives []Directive) (Ledger, error) {
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
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			balances = append(balances, d)
		case dirOpen, dirClose:
			d, err := newAccountEvent(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			accountEvents = append(accountEvents, d)
		case dirTxn, dirStar, dirBang:
			d, err := newTransaction(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			transactions = append(transactions, d)
		case dirPrice:
			d, err := newPrice(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			prices = append(prices, d)
		case dirPad:
			d, err := newPad(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			pads = append(pads, d)
		case dirNote, dirCommodity, dirQuery, dirCustom:
		default:
			return Ledger{}, fmt.Errorf("in NewLedger: found unrecognised directive: %s", typeStr)
		}
	}
	debugSlice(transactions, "transactions")
	debugSlice(accountEvents, "accountEvents")
	debugSlice(balances, "balances")
	debugSlice(prices, "prices")
	debugSlice(pads, "pads")
	ledger := Ledger{
		accountEvents,
		balances,
		transactions,
		prices,
		pads,
	}
	l = &ledger
	return ledger, nil
}

// parse does all the work of loading a file path
// and returning a Ledger
func (l *Ledger) parse(rc io.ReadCloser) (Ledger, error) {
	// getTokens never errors currently
	tokens, _ := getTokens(rc)
	// makeLines never errors currently
	lines, _ := makeLines(tokens)
	debugSlice(lines, "lines")
	directives, err := makeDirectives(lines)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	debugSlice(directives, "directives")

	ledger, err := l.fill(directives)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	return ledger, nil
}

// GetBalances returns the final balance of
// all accounts, separately for each currency
func (l *Ledger) GetBalances(rc io.ReadCloser) (AccBal, error) {
	ledger, err := l.parse(rc)
	if err != nil {
		return nil, fmt.Errorf("in GetBalances: %w", err)
	}
	ledger.Transactions, err = balanceTransactions(ledger.Transactions)
	if err != nil {
		return nil, fmt.Errorf("in GetBalances: %w", err)
	}
	debugSlice(ledger.Transactions, "ledger.Transactions")

	// extractPostings never errors currently
	postings, _ := extractPostings(ledger.Transactions)
	debugSlice(postings, "postings")

	// getBalances never errors currently
	accBalances, _ := getBalances(postings)
	return accBalances, nil
}
