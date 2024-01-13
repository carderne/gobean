package bean

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_AccountEvent_String(t *testing.T) {
	ae := AccountEvent{
		Date:    time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		Open:    true,
		Account: Account{AccountName("Assets:Bank")},
		Ccy:     Ccy("GBP"),
	}
	got := ae.String()
	want := "2022-01-01 open Assets:Bank GBP\n"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
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
