package bean

import (
	"testing"
	"time"

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

func Test_Line_LineNum(t *testing.T) {
	line := Line{true, []Token{}}
	got := line.LineNum()
	want := -1
	if got != want {
		t.Error("blank line should return -1 line num")
	}
}

func Test_Directive_LineNum(t *testing.T) {
	line := Line{false, []Token{{LineNum: 2}}}
	directive := Directive{[]Line{line}}
	got := directive.LineNum()
	want := 2
	if got != want {
		t.Error("directive should return line num")
	}
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

func Test_Balance_String(t *testing.T) {
	b := Balance{
		Date:    time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		Account: Account{AccountName("Assets:Bank")},
		Amount:  MustNewAmount("100", "GBP"),
	}
	got := b.String()
	want := "2022-01-01 balance Assets:Bank 100 GBP\n"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
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
