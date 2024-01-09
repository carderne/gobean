package bean_test

import (
	"github.com/carderne/gobean/bean"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestGetBalances(t *testing.T) {
	res, _ := bean.GetBalances("./testdata/basic.bean")
	expected := bean.AccBal{
		"Assets:Bank":   bean.CcyBal{"GBP": 860.00},
		"Income:Job":    bean.CcyBal{"GBP": -1000.00},
		"Expenses:Food": bean.CcyBal{"GBP": 140.00},
	}

	if diff := cmp.Diff(expected, res); diff != "" {
		t.Error(diff)
	}
}
