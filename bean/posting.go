package bean

import (
	"fmt"
	"log"
)

// Posting is an individual leg of a transaction
type Posting struct {
	Account     Account
	Amount      *Amount      // to allow nil
	Transaction *Transaction // nil until exctractPostings is run
}

// newPosting creates a Posting from a Line
// NB: Not from a Directive, as Postings are not directives!
func newPosting(line Line) (Posting, error) {
	log.Println("newPosting", line.Tokens[0].Text)
	tokens := line.Tokens
	accountStr := tokens[0].Text
	var amount *Amount
	if len(tokens) >= 3 {
		numberStr := tokens[1].Text
		ccy := tokens[2].Text
		amountVal := MustNewAmount(numberStr, ccy)
		amount = &amountVal
	}
	posting := Posting{
		Account: Account{AccountName(accountStr)},
		Amount:  amount,
	}
	return posting, nil
}

func (p Posting) String() string {
	var amountStr string
	if p.Amount != nil {
		amountStr = fmt.Sprintf("%v", p.Amount)
	}
	return fmt.Sprintf("%v: %v", p.Account.Name, amountStr)
}

// extractPostings flattens the Postings inside the slice of Transactions
// into a single slice of Postings
func extractPostings(transactions []Transaction) ([]Posting, error) {
	postings := make([]Posting, 0, 2*len(transactions))
	for _, t := range transactions {
		for _, p := range t.Postings {
			p.Transaction = &t
			postings = append(postings, p)
		}
	}
	return postings, nil
}

// getBalances returns a map containing the balance for each account-ccy pair
func getBalances(postings []Posting) (AccBal, error) {
	bals := make(AccBal, 20)
	for _, p := range postings {
		acc := p.Account.Name
		num := p.Amount.Number
		ccy := p.Amount.Ccy

		if bals[acc] == nil {
			bals[acc] = make(CcyAmount, 3)
		}

		other := Amount{num, ccy}
		cur, ok := bals[acc][ccy]
		if ok {
			bals[acc][ccy] = cur.MustAdd(other)
		} else {
			bals[acc][ccy] = other
		}
	}
	return bals, nil
}
