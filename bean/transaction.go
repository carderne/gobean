package bean

import (
	"fmt"
	"log"
	"time"
)

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

// newTransaction creates a Transaction (with Postings)
// from a Directive.
func newTransaction(directive Directive) (Transaction, error) {
	// first line is the root transaction line
	rootLine := directive.Lines[0]
	log.Println("newTransaction", rootLine.Tokens[0].Text)
	tokens := rootLine.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Transaction{}, fmt.Errorf("in newTransaction: %w", err)
	}
	txType := tokens[1].Text

	// If there is only one text, it is the narration
	// if there are two, first is payee, second is narration.
	// Dont ask me, I didn't design beancount!
	narration := tokens[2].Text
	var payee string
	if len(tokens) >= 4 {
		if tokens[3].Quote {
			payee = narration
			narration = tokens[3].Text
		}
	}

	var postings []Posting
	for _, line := range directive.Lines[1:] {
		// newPosting doesnt error currently
		p, _ := newPosting(line)
		postings = append(postings, p)
	}

	transaction := Transaction{
		Date:      date,
		Type:      txType,
		Payee:     payee,
		Narration: narration,
		Postings:  postings,
	}
	return transaction, nil
}

// balanceTransaction checks that a Transaction balances for all ccys
// The Posting _without_ an Amount (max one) will be used to auto-balance
// any currencies that dont already balance.
func balanceTransaction(transaction Transaction) (Transaction, error) {
	log.Println("Balancing", transaction.Date.Format(time.DateOnly), transaction.Narration)
	ccyBalances := make(CcyAmount, 3)
	postings := make([]Posting, 0, len(transaction.Postings))
	emptyPostingIndex := -1
	for i, p := range transaction.Postings {
		log.Printf("  Posting %v", p.String())
		if p.Amount == nil {
			if emptyPostingIndex != -1 {
				return Transaction{}, fmt.Errorf("cannot have multiple empty postings: %s", transaction)
			}
			emptyPostingIndex = i
		} else {
			curVal, ok := ccyBalances[p.Amount.Ccy]
			if ok {
				ccyBalances[p.Amount.Ccy] = curVal.MustAdd(*p.Amount)
			} else {
				ccyBalances[p.Amount.Ccy] = *p.Amount
			}
			postings = append(postings, p)
		}
	}
	if emptyPostingIndex != -1 {
		log.Println("found empty postings!")
		// note that we dont actually use the empty posting, just its account
		// because we will need more than 1 if there are multiple unbalanced ccys
		// and this makes the logic slightly easier
		account := transaction.Postings[emptyPostingIndex].Account
		for _, num := range ccyBalances {
			neg := num.Neg()
			p := Posting{
				Account: account,
				Amount:  &neg,
			}
			log.Printf("  new posting %v", p.String())
			postings = append(postings, p)
		}
	}
	transaction.Postings = postings
	return transaction, nil
}

// balanceTransactions balances all Transactions and returns the new
// balanced Transactions (original not modified)
func balanceTransactions(transactions []Transaction) ([]Transaction, error) {
	for i, tx := range transactions {
		transaction, err := balanceTransaction(tx)
		if err != nil {
			return nil, fmt.Errorf("in balanceTransactions: %w", err)
		}
		transactions[i] = transaction
	}
	return transactions, nil
}
