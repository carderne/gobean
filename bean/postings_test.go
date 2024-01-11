package bean

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func Test_getBalances(t *testing.T) {
	postings := []Posting{}
	got, _ := getBalances(postings)
	want := AccBal{}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}
