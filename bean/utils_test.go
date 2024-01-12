package bean

import (
	"testing"
	"time"
)

func TestGetDate(t *testing.T) {
	got, _ := getDate("2022-01-01")
	want := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	if got != want {
		t.Errorf("incorrect result: want %s, got %s", want, got)
	}

	got, err := getDate("invalid")
	if err == nil {
		t.Errorf("incorrect result: expected error")
	}
}

func TestDebugSlice(t *testing.T) {
	EmptyLedger(true)
	lines := []Line{
		{Tokens: []Token{{Text: "hi"}}},
	}
	debugSlice(lines, "lines")
}

func Test_PrintAccBalances(t *testing.T) {
	EmptyLedger(true)
	bals := map[string]string{"GBP": "100"}
	amt := MustNewCcyAmount(bals)
	accbals := AccBal{"Assets:Bank": amt}
	PrintAccBalances(accbals)
}

func TestDebugTokens(t *testing.T) {
	EmptyLedger(true)
	tokens := []Token{{Text: "hi"}}
	debugTokens(tokens)
}
