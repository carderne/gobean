package bean_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/carderne/gobean/bean"
	"github.com/google/go-cmp/cmp"
)

func Test_NewLedger(t *testing.T) {
	// debug setting should be applied
	_ = bean.NewLedger(true)
}

func Test_GetBalances(t *testing.T) {
	// file should be correctly balanced
	file, err := os.Open("./testdata/basic.bean")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	date := time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)
	l, _ := bean.NewLedger(false).Load(file)
	got, _ := l.GetBalances(date)
	want := bean.AccBal{
		"Assets:Bank":   bean.MustNewCcyAmount(map[string]string{"GBP": "860.00"}),
		"Income:Job":    bean.MustNewCcyAmount(map[string]string{"GBP": "-1000.00"}),
		"Expenses:Food": bean.MustNewCcyAmount(map[string]string{"GBP": "140.00"}),
	}

	comparer := cmp.Comparer(func(x, y bean.Amount) bool {
		return x.Eq(y)
	})
	if diff := cmp.Diff(want, got, comparer); diff != "" {
		t.Error(diff)
	}

	// multiple blank postings should error
	text := `
** Transactions
2023-02-01 * "Salary"
  Assets:Bank
  Income:Job
`
	rc := io.NopCloser(strings.NewReader(text))
	date = time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)
	_, err = bean.NewLedger(false).Load(rc)
	fmt.Println("AAAAAAA", l, err)
	if err == nil {
		t.Error("must fail with multiple blank postings")
	}

	// invalid file should error
	text = `
2023 02 01 * "Salary" Assets:Bank
`
	rc = io.NopCloser(strings.NewReader(text))
	date = time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)
	_, err = bean.NewLedger(false).Load(rc)
	if err == nil {
		t.Error("must fail with invalid file")
	}
}
