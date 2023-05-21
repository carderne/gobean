package bean

import (
	"fmt"
	"log"
)

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func balanceTransaction(transaction Transaction) (Transaction, error) {
	log.Println("Balancing", transaction.Date.Format(dateLayout), transaction.Narration)
	ccyBalances := make(map[string]float64, 3)
	postings := make([]Posting, 0, len(transaction.Postings))
	emptyPostingIndex := -1
	for i, p := range transaction.Postings {
		log.Printf("  Posting %v", p.String())
		if p.Amount == nil {
			if emptyPostingIndex != -1 {
				log.Panic("cannot have multiple empty postings")
			} else {
				emptyPostingIndex = i
			}
		} else {
			ccyBalances[p.Amount.Ccy] += p.Amount.Number
			postings = append(postings, p)
		}
	}
	if emptyPostingIndex != -1 {
		log.Println("found empty postings!")
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
