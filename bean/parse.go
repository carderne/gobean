package bean

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"
)

const eol = "\n"

// getTokens loads all text from the file at `path`
// loading a single rune at a time.
// Tokens are manually split on newlines and spaces
// and an EOL is added at the end
// The result is a slice of tokens with EOLs that can be
// used to separate lines
func getTokens(path string) ([]Token, error) {
	// TODO: accept an io.Reader instead
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("in getTokens: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanRunes)

	tokens := make([]Token, 0, 1000) // TODO: estimate size needed?

	type stateType struct {
		current     string // current token accumulator
		prev        string // used to check if prev rune was EOL
		inQuotes    bool   // whether current runes are in quotes
		inComment   bool   // are current runes inside a comment token
		tokenQuoted bool   // whether last finished token was in quotes
		indented    bool   // does the current token follow an indent
	}
	s := stateType{}

	for scanner.Scan() {
		r := scanner.Text()
		isEOL := r == eol
		isSpace := unicode.IsSpace([]rune(r)[0])
		onNewline := s.prev == "" || s.prev == eol
		s.indented = s.indented || (onNewline && isSpace)
		if isEOL || isSpace {
			// EOL or space will close a token
			// unless we are in a commend/quotation
			if s.inQuotes || (s.inComment && !isEOL) {
				s.current += r
			} else {
				// dont add empty tokens
				if s.current != "" {
					tokens = append(tokens, Token{
						Indent:  s.indented,
						Quote:   s.tokenQuoted,
						Comment: s.inComment,
						Text:    s.current,
					})
					s = stateType{} // reset state
				}
				// insert EOL so that subsequent funcs can split lines
				if isEOL {
					tokens = append(tokens, Token{
						EOL: true,
					})
					s = stateType{} // reset state
				}
			}
		} else if r == "\"" {
			if s.inQuotes {
				// we are closing quotes, and the accumulated token will be quoted
				s.tokenQuoted = true
			}
			s.inQuotes = !s.inQuotes
		} else if r == ";" || (r == "*" && onNewline) {
			// * only counts as a comment if it's the first rune on a line
			log.Println("COM", r)
			s.inComment = true
		} else {
			s.current += r
		}
		s.prev = r
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("in getTokens: %w", err)
	}
	// manually added to make subsequent funcs lives easier
	tokens = append(tokens, Token{
		EOL: true,
	})
	debugTokens(tokens)
	return tokens, nil
}

// makesLines simply splits the slice of Token
// into a nested slice with Tokens groups into Lines
func makeLines(tokens []Token) ([]Line, error) {
	var lines []Line
	curLine := Line{}
	prevToken := Token{}

	for _, t := range tokens {
		if t.Comment {
			log.Printf("ignoring comment %s", t.Text)
		} else if t.EOL {
			// blank lines are semantically significant
			if prevToken.EOL {
				lines = append(lines, Line{Blank: true})
			}
			// otherwise ignore lines with no tokens
			if len(curLine.Tokens) > 0 {
				lines = append(lines, curLine)
				curLine = Line{}
			}
		} else {
			curLine.Tokens = append(curLine.Tokens, t)
		}
		prevToken = t
	}
	return lines, nil
}

// makeDirectives groups together Lines that are logically joined.
// The 'root' line is always unindented, and subsequent lines
// must be indented to form part of the directive.
// Metadata of the form key:value (lower-case) can be added to any directive.
// This is mostly used for adding Postings to Transactions
func makeDirectives(lines []Line) ([]Directive, error) {
	var directives []Directive
	var curDirective Directive

	appendAndBlank := func() {
		if len(curDirective.Lines) > 0 {
			directives = append(directives, curDirective)
		}
		curDirective = Directive{}
	}

	for _, line := range lines {
		// a blank line always ends a directive
		if line.Blank {
			log.Println("blank")
			appendAndBlank()
		} else if line.Tokens[0].Indent {
			if len(curDirective.Lines) == 0 {
				return nil, fmt.Errorf("indented expression outside directive: %s", line)
			}
			log.Println("indent")
			r := rune(line.Tokens[0].Text[0])
			if unicode.IsLower(r) {
				log.Println("ignore Tags:", line.Tokens[0].Text)
			} else {
				curDirective.Lines = append(curDirective.Lines, line)
			}
		} else if line.Tokens[0].Text == "option" {
			// TODO do something with options
			log.Println("ignore: option")
		} else {
			log.Println("normal", line.Tokens[0].Text)
			appendAndBlank()
			curDirective.Lines = append(curDirective.Lines, line)
		}
	}
	log.Println()
	return directives, nil
}

// newLedger creates the basic ledger with
// accountEvents (open/close), balance directives and transactions.
// These are not yet logically validated, only checked semantically
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
			d, err := newBalance(directive)
			if err != nil {
				return Ledger{}, fmt.Errorf("in newLedger: %w", err)
			}
			balances = append(balances, d)
		case "open", "close":
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

	debugSlice(transactions, "transactions")
	debugSlice(accountEvents, "accountEvents")
	debugSlice(balances, "balances")
	ledger := Ledger{
		accountEvents,
		balances,
		transactions,
	}
	return ledger, nil
}

// newAccountEvent creates an AccountEvent from a Directive
func newAccountEvent(directive Directive) (AccountEvent, error) {
	line := directive.Lines[0] // TODO include metadata lines
	log.Println("newAccountEvent", line.Tokens[0].Text)
	tokens := line.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return AccountEvent{}, fmt.Errorf("in newAccountEvent: %w", err)
	}
	open := tokens[1].Text == "open"
	account := tokens[2].Text
	var ccy Ccy
	if len(tokens) >= 4 {
		// TODO should return err if ccy provided on close
		ccy = Ccy(tokens[3].Text)
	}
	if len(tokens) >= 5 {
		// TODO handle additional currencies
		log.Println("ignoring extra open/close tokens")
	}
	accountEvent := AccountEvent{
		Date:    date,
		Open:    open,
		Account: Account{AccountName(account)},
		Ccy:     ccy,
	}
	return accountEvent, nil
}

// newBalance creates a Balance from a Directive
func newBalance(directive Directive) (Balance, error) {
	line := directive.Lines[0] // TODO include metadata lines
	log.Println("newBalance", line.Tokens[0].Text)
	tokens := line.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Balance{}, fmt.Errorf("in newBalance: %w", err)
	}
	account := tokens[2].Text
	numberStr := tokens[3].Text
	number, _ := strconv.ParseFloat(string(numberStr), 64)
	ccy := tokens[4].Text
	if len(tokens) > 5 {
		return Balance{}, fmt.Errorf("too many balance tokens: %s", directive)
	}

	balance := Balance{
		Date:    date,
		Account: Account{AccountName(account)},
		Amount: Amount{
			Number: number,
			Ccy:    Ccy(ccy),
		},
	}
	return balance, nil
}

// newPosting creates a Posting from a Line
// NB: Not from a Directive, as Postings are not directives!
func newPosting(line Line) (Posting, error) {
	log.Println("newPosting", line.Tokens[0].Text)
	tokens := line.Tokens
	accountStr := tokens[0].Text
	var amount *Amount
	if len(tokens) >= 3 {
		numberStr := tokens[1].Text
		number, _ := strconv.ParseFloat(string(numberStr), 64)
		var ccy = tokens[2].Text
		amount = &Amount{
			Number: number,
			Ccy:    Ccy(ccy),
		}
	}
	posting := Posting{
		Account: Account{AccountName(accountStr)},
		Amount:  amount,
	}
	return posting, nil
}

// newTransaction creates a Transaction (with Postings)
// from a Directive.
func newTransaction(directive Directive) (Transaction, error) {
	// first line is the root transaction line
	rootLine := directive.Lines[0]
	log.Println("newTransaction", rootLine.Tokens[0].Text)
	tokens := rootLine.Tokens
	date, err := getDate(tokens[0].Text)
	if err != nil {
		return Transaction{}, fmt.Errorf("in newTransaction: %w", err)
	}
	txType := tokens[1].Text

	// If there is only one text, it is the narration
	// if there are two, first is payee, second is narration.
	// Dont ask me, I didn't design beancount!
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

// parse does all the work of loading a file path
// and returning a Ledger
func parse(path string) (Ledger, error) {
	tokens, err := getTokens(path)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	lines, err := makeLines(tokens)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	debugSlice(lines, "lines")
	directives, err := makeDirectives(lines)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	debugSlice(directives, "directives")

	ledger, err := newLedger(directives)
	if err != nil {
		return Ledger{}, fmt.Errorf("in parse: %w", err)
	}
	return ledger, nil
}
