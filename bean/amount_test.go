package bean

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func assertPanic(t *testing.T, f func(), msg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(msg)
		}
	}()
	f()
}

func Test_NewAmount(t *testing.T) {
	// invalid in NewAmount should error
	_, err := NewAmount("foo", "GBP")
	if err == nil {
		t.Error("invalid number in NewAmount should error")
	}

	// invalid in MustNewAmount should panic
	f := func() {
		MustNewAmount("foo", "GBP")
	}
	assertPanic(t, f, "MustNewAmount should panic with invalid")
}

func Test_Amount_Eq(t *testing.T) {
	// unequal ccy should be false
	a := MustNewAmount("100", "GBP")
	b := MustNewAmount("100", "USD")
	got := a.Eq(b)
	want := false
	if got != want {
		t.Error("unequal ccy should be false")
	}
}

func Test_Amount_MustAdd(t *testing.T) {
	// invalid Add should error
	a := MustNewAmount("100", "GBP")
	b := MustNewAmount("100", "USD")
	_, err := a.Add(b)
	if err == nil {
		t.Error("invalid Add should error")
	}

	// invalid MustAdd should panic
	f := func() {
		a.MustAdd(b)
	}
	assertPanic(t, f, "MustAdd should panic")
}

func Test_NewCcyAmount(t *testing.T) {
	ccy := "GBP"
	num := "100"
	amt, _ := NewAmount(num, ccy)
	bals := map[string]string{ccy: num}
	got, _ := NewCcyAmount(bals)
	want := CcyAmount{Ccy(ccy): amt}
	comparer := cmp.Comparer(func(x, y Amount) bool {
		return x.Eq(y)
	})
	if diff := cmp.Diff(want, got, comparer); diff != "" {
		t.Error(diff)
	}

	// invalid should error
	ccy = "GBP"
	bals = map[string]string{ccy: "foo"}
	_, err := NewCcyAmount(bals)
	if err == nil {
		t.Error("invalid NewCcyAmount should error")
	}
}
