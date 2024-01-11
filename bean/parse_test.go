package bean

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGetTokens(t *testing.T) {
	// normal text should be parsed to tokens
	text := `
** Transactions
2023-02-01 * "Salary"
  Assets:Bank                          1000 GBP
  Income:Job
`
	rc := io.NopCloser(strings.NewReader(text))
	got, _ := getTokens(rc)
	want := []Token{
		{LineNum: 1, EOL: true},
		{LineNum: 2, Comment: true, Text: "** Transactions"},
		{LineNum: 2, EOL: true},
		{LineNum: 3, Text: "2023-02-01"},
		{LineNum: 3, Text: "*"},
		{LineNum: 3, Quote: true, Text: "Salary"},
		{LineNum: 3, EOL: true},
		{LineNum: 4, Indent: true, Text: "Assets:Bank"},
		{LineNum: 4, Text: "1000"},
		{LineNum: 4, Text: "GBP"},
		{LineNum: 4, EOL: true},
		{LineNum: 5, Indent: true, Text: "Income:Job"},
		{LineNum: 5, EOL: true},
		{LineNum: 6, EOL: true},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestMakeDirectives(t *testing.T) {
	// dangling indent should return an error
	text := `
  Income:Job
  `
	rc := io.NopCloser(strings.NewReader(text))
	tokens, _ := getTokens(rc)
	lines, _ := makeLines(tokens)
	_, err := makeDirectives(lines)
	if err == nil {
		t.Error("parse should fail with indented expression outside directive")
	}

	// tags should be ignored
	text = `
option "operating_currency" "GBP"
2023-02-01 * "Salary"
  tag:value
  Assets:Bank                          1000 GBP
  Income:Job
  `
	rc = io.NopCloser(strings.NewReader(text))
	tokens, _ = getTokens(rc)
	lines, _ = makeLines(tokens)
	got, _ := makeDirectives(lines)
	want := []Directive{{[]Line{
		{Tokens: []Token{
			{LineNum: 3, Text: "2023-02-01"},
			{LineNum: 3, Text: "*"},
			{LineNum: 3, Quote: true, Text: "Salary"},
		}},
		{Tokens: []Token{
			{LineNum: 5, Indent: true, Text: "Assets:Bank"},
			{LineNum: 5, Text: "1000"},
			{LineNum: 5, Text: "GBP"},
		}},
		{Tokens: []Token{
			{LineNum: 6, Indent: true, Text: "Income:Job"},
		}},
	}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

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
	got, _ := newLedger(directives)
	comparer := cmp.Comparer(func(x, y Amount) bool {
		return x.Eq(y)
	})
	if diff := cmp.Diff(want, got, comparer); diff != "" {
		t.Error(diff)
	}

	// empty directive should be ignored
	directives = []Directive{{[]Line{}}}
	want = Ledger{}
	got, _ = newLedger(directives)
	if diff := cmp.Diff(want, got); diff != "" {
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
	_, err := newLedger(directives)
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
	_, err = newLedger(directives)
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
	_, err = newLedger(directives)
	if err == nil {
		t.Error("invalid account event should raise")
	}
}

func TestNewAccountEvent(t *testing.T) {
	// extra account currencies should be ignored
	acc := "Assets:Bank"
	ccy := "GBP"
	directive := Directive{[]Line{
		{Tokens: []Token{
			{LineNum: 1, Text: "2023-01-01"},
			{LineNum: 1, Text: "open"},
			{LineNum: 1, Text: acc},
			{LineNum: 1, Text: ccy},
			{LineNum: 1, Text: ccy},
		}},
	}}
	want := AccountEvent{
		Date:    time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
		Open:    true,
		Account: Account{AccountName(acc)},
		Ccy:     Ccy(ccy),
	}
	got, _ := newAccountEvent(directive)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestNewBalance(t *testing.T) {
	// too many balance tokens should fail
	acc := "Assets:Bank"
	num := "100"
	ccy := "GBP"
	directive := Directive{[]Line{
		{Tokens: []Token{
			{LineNum: 1, Text: "2023-01-01"},
			{LineNum: 1, Text: "balance"},
			{LineNum: 1, Text: acc},
			{LineNum: 1, Text: num},
			{LineNum: 1, Text: ccy},
			{LineNum: 1, Text: ccy},
		}},
	}}
	_, err := newBalance(directive)
	if err == nil {
		t.Error("too many balance tokens should error")
	}
}

func TestParse(t *testing.T) {
	// dangling indent should error
	text := `
  Assets:Bank
`
	rc := io.NopCloser(strings.NewReader(text))
	_, err := parse(rc)
	if err == nil {
		t.Error("dangling indent should error")
	}
}
