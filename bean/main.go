// Package bean contains all beancount logic
// Including parsing, validating, calculating etc
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

// Bean is empty for now but sure to grow!
type Bean struct{}

// NewBean returns an empty bean struct
func NewBean(dbg bool) Bean {
	debug = dbg
	if debug {
		log.SetOutput(os.Stderr)
	}
	return Bean{}
}

// GetBalances returns the final balance of
// all accounts, separately for each currency
func (Bean) GetBalances(rc io.ReadCloser) (AccBal, error) {
	ledger, err := parse(rc)
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
