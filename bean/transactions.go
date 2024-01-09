package bean

import (
	"fmt"
	"log"
)

// balanceTransaction checks that a Transaction balances for all ccys
// The Posting _without_ an Amount (max one) will be used to auto-balance
// any currencies that dont already balance.
func balanceTransaction(transaction Transaction) (Transaction, error) {
	log.Println("Balancing", transaction.Date.Format(dateLayout), transaction.Narration)
	ccyBalances := make(CcyBal, 3)
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
			ccyBalances[p.Amount.Ccy] += p.Amount.Number
			postings = append(postings, p)
		}
	}
	if emptyPostingIndex != -1 {
		log.Println("found empty postings!")
		// note that we dont actually use the empty posting, just its account
		// because we will need more than 1 if there are multiple unbalanced ccys
		// and this makes the logic slightly easier
		account := transaction.Postings[emptyPostingIndex].Account
		for ccy, num := range ccyBalances {
			p := Posting{
				Account: account,
				Amount: &Amount{
					Number: -num,
					Ccy:    ccy,
				},
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
