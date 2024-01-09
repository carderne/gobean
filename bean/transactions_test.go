package bean

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_balanceTransaction(t *testing.T) {
	transaction := Transaction{
		Date:      time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		Type:      "*",
		Narration: "test",
		Postings: []Posting{
			{
				Account: Account{"Assets:Bank"},
				Amount:  &Amount{100, "GBP"},
			},
			{
				Account: Account{"Income:Job"},
				Amount:  &Amount{-100, "GBP"},
			},
		},
	}
	res, _ := balanceTransaction(transaction)

	if diff := cmp.Diff(transaction, res); diff != "" {
		t.Error(diff)
	}
}
