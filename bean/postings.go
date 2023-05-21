package bean

import ()

// CcyBal is a map of Ccy -> number
type CcyBal = map[string]float64

// AccBal is a map of Account -> CcyBal
type AccBal = map[string]CcyBal

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

func getBalances(postings []Posting) (AccBal, error) {
	bals := make(AccBal, 20)
	for _, p := range postings {
		acc := p.Account.Full
		val := p.Amount.Number
		ccy := p.Amount.Ccy

		if bals[acc] == nil {
			bals[acc] = make(CcyBal, 3)
		}

		bals[acc][ccy] += val
	}
	return bals, nil
}
