package bean

import (
	"fmt"
	"log"
	"time"
)

func getDate(dateStr string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("in getDate: %w", err)
	}
	return date, nil
}

func debugSlice[T fmt.Stringer](els []T, msg string) {
	if debug {
		fmt.Println("--", msg)
		for _, el := range els {
			fmt.Print(el)
			fmt.Println()
		}
		fmt.Println("-- END", msg)
	}
}

// PrintAccBalances pretty prints account balances by currency
func PrintAccBalances(accBalances AccBal) {
	for acc, bals := range accBalances {
		fmt.Println(acc)
		for ccy, amt := range bals {
			fmt.Printf("  %s %s\n", amt.Number.Text('f'), ccy)
		}
	}
}

func debugTokens(tokens []Token) {
	if debug {
		fmt.Println()
		log.Println("debugTokens:")
		for _, t := range tokens {
			fmt.Print(t.Text, " | ")
		}
		log.Println("\ndebugTokens END")
		fmt.Println()
	}
}
