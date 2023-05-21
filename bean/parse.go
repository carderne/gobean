package bean

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"
)

func getTokens(path string) ([]Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("in getTokens: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanRunes)

	var prev string
	var current string
	tokens := make([]Token, 0, 1000)

	inQuotes := false
	inComment := false
	tokenQuoted := false
	onNewline := true
	indented := false

	for scanner.Scan() {
		r := scanner.Text()
		isNewline := r == "\n"
		isSpace := unicode.IsSpace([]rune(r)[0])
		indented = indented || (prev == "\n" && isSpace)
		if isNewline || isSpace {
			if inQuotes || (inComment && !isNewline) {
				current += r
			} else {
				if current != "" {
					tokens = append(tokens, Token{
						Indent:  indented,
						Quote:   tokenQuoted,
						Comment: inComment,
						Text:    current,
					})
					inComment = false
					tokenQuoted = false
					onNewline = false
					indented = false
					current = ""
				}
				if isNewline {
					tokens = append(tokens, Token{
						EOL: true,
					})
					inComment = false
					onNewline = true
					indented = false
					current = ""
				}
			}
		} else if r == "\"" {
			if inQuotes {
				tokenQuoted = true
			}
			inQuotes = !inQuotes
		} else if r == ";" || (r == "*" && onNewline) {
			inComment = true
		} else {
			current += r
		}
		prev = r
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("in getTokens: %w", err)
	}
	tokens = append(tokens, Token{
		EOL: true,
	})
	debugTokens(tokens)
	return tokens, nil
}

func makeLines(tokens []Token) ([]Line, error) {
	var lines []Line
	curLine := Line{}
	prevToken := Token{}

	for i, t := range tokens {
		if t.Comment {
			log.Println("TOK", i, "COM")
			continue
		} else if t.EOL {
			log.Println("TOK", i, "EOL")
			if prevToken.EOL {
				lines = append(lines, Line{Blank: true})
			}
			if len(curLine.Tokens) > 0 {
				lines = append(lines, curLine)
				curLine = Line{}
			}
		} else {
			log.Println("TOK", i, t.Text)
			curLine.Tokens = append(curLine.Tokens, t)
		}
	}
	return lines, nil
}

func makeDirectives(lines []Line) ([]Directive, error) {
	var directives []Directive
	var curDirective Directive

	for _, line := range lines {
		if line.Blank {
			//
		} else if line.Tokens[0].Indent {
			log.Println("INDENT")
			r := rune(line.Tokens[0].Text[0])
			if unicode.IsLower(r) {
				log.Println("ignore Tags:", line.Tokens[0].Text)
				continue
			}
			curDirective.Lines = append(curDirective.Lines, line)
		} else if line.Tokens[0].Text == "option" {
			log.Println("ignore: OPTION")
		} else {
			log.Println("NORMAL", line.Tokens[0].Text)
			if len(curDirective.Lines) > 0 {
				directives = append(directives, curDirective)
			}
			curDirective = Directive{}
			curDirective.Lines = append(curDirective.Lines, line)
		}
	}
	log.Println()
	return directives, nil
}

func newLedger(directives []Directive) (Ledger, error) {
	var accountEvents []AccountEvent
	var balances []Balance
	var transactions []Transaction

	for _, directive := range directives {
		if len(directive.Lines) == 0 {
			continue
		}
		switch typeStr := directive.Lines[0].Tokens[1].Text; typeStr {
		case "balance":
			// ignoring children for now
			d, err := newBalance(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			balances = append(balances, d)
		case "open", "close":
			// ignoring children for now
			d, err := newAccountEvent(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			accountEvents = append(accountEvents, d)
		case "pad", "price", "note", "commodity", "query", "custom":
		default:
			d, err := newTransaction(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			transactions = append(transactions, d)
		}
	}

	debugSlice(transactions)
	debugSlice(accountEvents)
	debugSlice(balances)
	ledger := Ledger{
		accountEvents,
		balances,
		transactions,
	}
	return ledger, nil
}

func newAccountEvent(directive Directive) (AccountEvent, error) {
	line := directive.Lines[0] // ignore subsequent lines for now
	log.Println("newAccountEvent", line.Tokens[0].Text)
	tokens := line.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return AccountEvent{}, fmt.Errorf("in newAccountEvent: %w", err)
	}
	open := tokens[1].Text == "open"
	account := tokens[2].Text
	var ccy string
	if len(tokens) >= 4 {
		ccy = tokens[3].Text
	}
	if len(tokens) >= 5 {
		log.Println("ignoring extra open/close tokens")
	}
	accountEvent := AccountEvent{
		Date:    date,
		Open:    open,
		Account: Account{account},
		Ccy:     ccy,
	}
	return accountEvent, nil
}

func newBalance(directive Directive) (Balance, error) {
	line := directive.Lines[0] // ignore subsequent lines for now
	log.Println("newBalance", line.Tokens[0].Text)
	tokens := line.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Balance{}, fmt.Errorf("in newBalance: %w", err)
	}
	account := tokens[2].Text
	numberStr := tokens[3].Text
	number, _ := strconv.ParseFloat(numberStr, 64)
	ccy := tokens[4].Text
	if len(tokens) > 5 {
		panic("Expected EOL")
	}

	balance := Balance{
		Date:    date,
		Account: Account{account},
		Amount: Amount{
			Number: number,
			Ccy:    ccy,
		},
	}
	return balance, nil
}

func newPosting(line Line) (Posting, error) {
	log.Println("newPosting", line.Tokens[0].Text)
	tokens := line.Tokens
	accountStr := tokens[0].Text
	var amount *Amount
	if len(tokens) >= 3 {
		numberStr := tokens[1].Text
		number, _ := strconv.ParseFloat(numberStr, 64)
		var ccy = tokens[2].Text
		amount = &Amount{
			Number: number,
			Ccy:    ccy,
		}
	}
	posting := Posting{
		Account: Account{accountStr},
		Amount:  amount,
	}
	return posting, nil
}

func newTransaction(directive Directive) (Transaction, error) {
	// first rootLine is the root transaction
	rootLine := directive.Lines[0]
	log.Println("newTransaction", rootLine.Tokens[0].Text)
	tokens := rootLine.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Transaction{}, fmt.Errorf("in newTransaction: %w", err)
	}
	txType := tokens[1].Text

	narration := tokens[2].Text
	var payee string
	if len(tokens) >= 4 {
		if tokens[3].Quote {
			payee = narration
			narration = tokens[3].Text
		}
	}

	var postings []Posting
	for _, line := range directive.Lines[1:] {
		p, err := newPosting(line)
		if err != nil {
			return Transaction{}, fmt.Errorf("in newTransaction: %w", err)
		}
		postings = append(postings, p)
	}

	transaction := Transaction{
		Date:      date,
		Type:      txType,
		Payee:     payee,
		Narration: narration,
		Postings:  postings,
	}
	return transaction, nil
}

func parse(path string) (Ledger, error) {
	tokens, err := getTokens(path)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	lines, err := makeLines(tokens)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	debugSlice(lines)
	directives, err := makeDirectives(lines)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}

	ledger, err := newLedger(directives)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	return ledger, nil
}
