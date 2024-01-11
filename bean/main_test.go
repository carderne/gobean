package bean_test

import (
	"testing"

	"github.com/carderne/gobean/bean"
	"github.com/google/go-cmp/cmp"
)

func TestGetBalances(t *testing.T) {
	res, _ := bean.GetBalances("./testdata/basic.bean")
	expected := bean.AccBal{
		"Assets:Bank":   bean.MustNewCcyAmount(map[string]string{"GBP": "860.00"}),
		"Income:Job":    bean.MustNewCcyAmount(map[string]string{"GBP": "-1000.00"}),
		"Expenses:Food": bean.MustNewCcyAmount(map[string]string{"GBP": "140.00"}),
	}

	comparer := cmp.Comparer(func(x, y bean.Amount) bool {
		return x.Eq(y)
	})
	if diff := cmp.Diff(expected, res, comparer); diff != "" {
		t.Error(diff)
	}
}
