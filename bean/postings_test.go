package bean

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func Test_getBalances(t *testing.T) {
	postings := []Posting{}
	res, _ := getBalances(postings)
	expected := AccBal{}

	if diff := cmp.Diff(expected, res); diff != "" {
		t.Error(diff)
	}
}
