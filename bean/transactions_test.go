package bean

import (
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/google/go-cmp/cmp"
)

func Test_balanceTransaction(t *testing.T) {
	val1, _, _ := apd.NewFromString("100")
	val2, _, _ := apd.NewFromString("-100")
	transaction := Transaction{
		Date:      time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		Type:      "*",
		Narration: "test",
		Postings: []Posting{
			{
				Account: Account{"Assets:Bank"},
				Amount:  &Amount{*val1, "GBP"},
			},
			{
				Account: Account{"Income:Job"},
				Amount:  &Amount{*val2, "GBP"},
			},
		},
	}
	res, _ := balanceTransaction(transaction)

	if diff := cmp.Diff(transaction, res, cmp.AllowUnexported(apd.BigInt{})); diff != "" {
		t.Error(diff)
	}
}
