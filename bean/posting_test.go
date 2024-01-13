package bean

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_getBalances(t *testing.T) {
	postings := []Posting{}
	atl := AccountTimeLine{}
	date := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	got, _ := getBalances(postings, atl, date)
	want := AccBal{}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}

	// should fail with non-open accounts
	val1 := MustNewAmount("100", "GBP")
	val2 := MustNewAmount("-100", "GBP")
	tx := Transaction{
		Date:      time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		Type:      "*",
		Narration: "test",
	}
	postings = []Posting{
		{
			Account:     Account{"Assets:Bank"},
			Amount:      &val1,
			Transaction: &tx,
		},
		{
			Account:     Account{"Income:Job"},
			Amount:      &val2,
			Transaction: &tx,
		},
	}
	atl = AccountTimeLine{}
	_, err := getBalances(postings, atl, date)

	if err == nil {
		t.Errorf("getBalances should fail with non-open account")
	}
}
