package bean

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_getBalances(t *testing.T) {
	postings := []Posting{}
	got, _ := getBalances(postings)
	want := AccBal{}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}
