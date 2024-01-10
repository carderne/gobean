package bean_test

import (
	"fmt"
	"testing"

	"github.com/carderne/gobean/bean"
	"github.com/cockroachdb/apd/v3"
	"github.com/google/go-cmp/cmp"
)

func decimalCompare(x, y bean.CcyBal) bool {
	for k, v := range x {
		other := y[k]
		fmt.Println("---", k, v.Text('f'), other.Text('f'))
		if v.Text('f') != other.Text('f') {
			return false
		}
	}
	return true
}

func TestGetBalances(t *testing.T) {
	res, _ := bean.GetBalances("./testdata/basic.bean")
	expected := bean.AccBal{
		"Assets:Bank":   bean.NewCcyBal(map[string]string{"GBP": "860.00"}),
		"Income:Job":    bean.NewCcyBal(map[string]string{"GBP": "-1000.00"}),
		"Expenses:Food": bean.NewCcyBal(map[string]string{"GBP": "140.00"}),
	}
	apdCtx := apd.BaseContext

	opts := []cmp.Option{
		cmp.Comparer(func(x, y apd.Decimal) bool {
			equal := apd.New(0, 0)
			res := &apd.Decimal{}
			_, err := apdCtx.Cmp(res, &x, &y)
			if err != nil {
				return false
			}
			return *res == *equal

		}),
	}
	if diff := cmp.Diff(expected, res, opts...); diff != "" {
		t.Error(diff)
	}
}
