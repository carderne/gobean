package bean

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewLedger(t *testing.T) {
	// balance statement should work
	acc := "Assets:Bank"
	num := "100"
	ccy := "GBP"
	directives := []Directive{{[]Line{
		{Tokens: []Token{
			{LineNum: 1, Text: "2023-01-01"},
			{LineNum: 1, Text: "balance"},
			{LineNum: 1, Text: acc},
			{LineNum: 1, Text: num},
			{LineNum: 1, Text: ccy},
		}},
	}}}
	want := Ledger{
		Balances: []Balance{{
			Date:    time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			Account: Account{AccountName(acc)},
			Amount:  MustNewAmount(num, ccy),
		}},
	}
	got, _ := NewLedger(false).fill(directives)
	comparer := cmp.Comparer(func(x, y Amount) bool {
		return x.Eq(y)
	})
	if diff := cmp.Diff(&want, got, comparer); diff != "" {
		t.Error(diff)
	}

	// empty directive should be ignored
	directives = []Directive{{[]Line{}}}
	want = Ledger{}
	got, _ = NewLedger(false).fill(directives)
	if diff := cmp.Diff(&want, got); diff != "" {
		t.Error(diff)
	}

	// invalid balance should raise
	directives = []Directive{{[]Line{
		{Tokens: []Token{
			{LineNum: 1, Text: "xxx-01-01"},
			{LineNum: 1, Text: "balance"},
			{LineNum: 1, Text: acc},
			{LineNum: 1, Text: num},
			{LineNum: 1, Text: ccy},
		}},
	}}}
	want = Ledger{}
	_, err := NewLedger(false).fill(directives)
	if err == nil {
		t.Error("invalid balance should raise")
	}

	// invalid account event should raise
	directives = []Directive{{[]Line{
		{Tokens: []Token{
			{LineNum: 1, Text: "xxx-01-01"},
			{LineNum: 1, Text: "open"},
			{LineNum: 1, Text: acc},
			{LineNum: 1, Text: num},
			{LineNum: 1, Text: ccy},
		}},
	}}}
	want = Ledger{}
	_, err = NewLedger(false).fill(directives)
	if err == nil {
		t.Error("invalid account event should raise")
	}

	// invalid transaction should raise
	directives = []Directive{{[]Line{
		{Tokens: []Token{
			{LineNum: 1, Text: "xxx-01-01"},
			{LineNum: 1, Text: "*"},
			{LineNum: 1, Text: "hello"},
		}},
	}}}
	want = Ledger{}
	_, err = NewLedger(false).fill(directives)
	if err == nil {
		t.Error("invalid account event should raise")
	}
}

func TestParse(t *testing.T) {
	// dangling indent should error
	text := `
  Assets:Bank
`
	rc := io.NopCloser(strings.NewReader(text))
	_, err := NewLedger(false).parse(rc)
	if err == nil {
		t.Error("dangling indent should error")
	}
}
