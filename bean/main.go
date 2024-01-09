// Package bean contains all beancount logic
// Including parsing, validating, calculating etc
package bean

import (
	"io"
	"log"
	"os"
)

// Debug indicates whether DEBUG env var is set to 1
var Debug bool

func init() {
	var err error
	if err != nil {
		panic(err)
	}

	log.SetFlags(0)
	log.SetOutput(io.Discard)
	if os.Getenv("DEBUG") == "1" {
		log.SetOutput(os.Stderr)
		Debug = true
	}
}

// GetBalances returns the final balance of
// all accounts, separately for each currency
func GetBalances(path string) (AccBal, error) {
	ledger, err := parse(path)
	if err != nil {
		log.Fatal("parse failed: ", err)
	}
	ledger.Transactions, err = balanceTransactions(ledger.Transactions)
	if err != nil {
		log.Fatal("balanceTransactions failed: ", err)
	}
	debugSlice(ledger.Transactions, "ledger.Transactions")

	postings, err := extractPostings(ledger.Transactions)
	if err != nil {
		log.Fatal("extractPostings failed: ", err)
	}
	debugSlice(postings, "postings")

	accBalances, err := getBalances(postings)
	if err != nil {
		log.Fatal("Validate failed: ", err)
	}
	return accBalances, nil
}
