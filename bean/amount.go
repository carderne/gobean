package bean

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// Amount is a number with a currency
type Amount struct {
	Number apd.Decimal
	Ccy    Ccy
}

// NewAmount creates an Amount
func NewAmount(num string, ccy string) (Amount, error) {
	d, _, err := apdCtx.NewFromString(num)
	if err != nil {
		return Amount{}, fmt.Errorf("in NewAmount: %w", err)
	}
	return Amount{*d, Ccy(ccy)}, nil
}

// MustNewAmount creates an Amount and panics! on underlying errors
func MustNewAmount(num string, ccy string) Amount {
	res, err := NewAmount(num, ccy)
	if err != nil {
		panic(err)
	}
	return res
}

// Eq returns true if the Amounts have same Number and Ccy
func (a Amount) Eq(other Amount) bool {
	if a.Ccy != other.Ccy {
		return false
	}

	equal := apd.New(0, 0)
	res := &apd.Decimal{}
	_, err := apdCtx.Cmp(res, &a.Number, &other.Number)
	if err != nil {
		return false
	}
	return *res == *equal
}

// Add adds other to the amount
func (a Amount) Add(other Amount) (Amount, error) {
	if a.Ccy != other.Ccy {
		return Amount{}, fmt.Errorf("cant add Amount with different currency: a: %s other: %s", a.Ccy, other.Ccy)
	}
	curVal := a.Number
	newVal := apd.Decimal{}
	apdCtx.Add(&newVal, &curVal, &other.Number)
	a.Number = newVal
	return a, nil
}

// MustAdd adds other to the Amount
func (a Amount) MustAdd(other Amount) Amount {
	res, err := a.Add(other)
	if err != nil {
		panic(err)
	}
	return res
}

// Neg returns the negated Amount
func (a Amount) Neg() Amount {
	neg := apd.Decimal{}
	apdCtx.Neg(&neg, &a.Number)
	a.Number = neg
	return a
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.Number.Text('f'), a.Ccy)
}

// CcyAmount is a map of Ccy -> number
type CcyAmount = map[Ccy]Amount

// NewCcyAmount converts a string map to a CcyAmount
func NewCcyAmount(bals map[string]string) (CcyAmount, error) {
	res := make(CcyAmount, len(bals))
	for ccy, num := range bals {
		val, err := NewAmount(num, ccy)
		if err != nil {
			return nil, fmt.Errorf("in NewCcyBal: %w", err)
		}
		res[Ccy(ccy)] = val
	}
	return res, nil
}

// MustNewCcyAmount converts a regular string map to a CcyAmount
func MustNewCcyAmount(bals map[string]string) CcyAmount {
	res := make(CcyAmount, len(bals))
	for ccy, num := range bals {
		res[Ccy(ccy)] = MustNewAmount(num, ccy)
	}
	return res
}

// AccBal is a map of Account -> CcyBal
type AccBal = map[AccountName]CcyAmount
