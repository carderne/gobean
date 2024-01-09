package bean

import ()

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
		val := p.Amount.Number
		ccy := p.Amount.Ccy

		if bals[acc] == nil {
			bals[acc] = make(CcyBal, 3)
		}

		bals[acc][ccy] += val
	}
	return bals, nil
}
