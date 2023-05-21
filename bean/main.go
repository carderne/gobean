package bean

import (
	"io"
	"log"
	"os"
	"regexp"
)

var dateRegex *regexp.Regexp

// Debug indicates whether DEBUG env var is set to 1
var Debug bool

const dateLayout = "2006-01-02"
const datePattern = `^[0-9]{4}-[0-9]{2}-[0-9]{2}$`

func init() {
	var err error
	dateRegex, err = regexp.Compile(datePattern)
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

// GetBalances returns the first balance statement from the file
func GetBalances(path string) (AccBal, error) {
	ledger, err := parse(path)
	if err != nil {
		log.Fatal("parse failed: ", err)
	}
	ledger.Transactions, err = balanceTransactions(ledger.Transactions)
	if err != nil {
		log.Fatal("balanceTransactions failed: ", err)
	}
	debugSlice(ledger.Transactions)

	postings, err := extractPostings(ledger.Transactions)
	if err != nil {
		log.Fatal("extractPostings failed: ", err)
	}
	debugSlice(postings)

	accBalances, err := getBalances(postings)
	if err != nil {
		log.Fatal("Validate failed: ", err)
	}
	printAccBalances(accBalances)
	return accBalances, nil
}
