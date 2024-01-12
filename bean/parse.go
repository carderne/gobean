package bean

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"unicode"
)

const eol = "\n"

type dirType string

const (
	dirOpen      dirType = "open"
	dirClose     dirType = "close"
	dirBalance   dirType = "balance"
	dirTxn       dirType = "txn"
	dirStar      dirType = "*"
	dirBang      dirType = "!"
	dirPrice     dirType = "price"
	dirPad       dirType = "pad"
	dirNote      dirType = "note"
	dirCommodity dirType = "commodity"
	dirQuery     dirType = "query"
	dirCustom    dirType = "custom"
	dirOption    dirType = "option"
)

// Token is raw token from input file with a bunch of flags
// quotes are removed, newlines inside quotes are maintained
type Token struct {
	LineNum int
	Indent  bool
	Quote   bool
	Comment bool
	EOL     bool
	Text    string
}

// A Line from the beancount file
type Line struct {
	Blank  bool
	Tokens []Token
}

// LineNum returns the source file line number of this Line
func (l Line) LineNum() int {
	if l.Blank {
		return -1
	}
	return l.Tokens[0].LineNum
}

func (l Line) String() string {
	str := fmt.Sprintf("line:%d", l.LineNum())
	for _, t := range l.Tokens {
		str += fmt.Sprint(" | ", t.Text)
	}
	return str
}

// Directive is one or more lines that go together
type Directive struct {
	Lines []Line
}

// LineNum returns the soruce file number of the
// _first line_ of this line
func (d Directive) LineNum() int {
	return d.Lines[0].LineNum()
}

func (d Directive) String() string {
	str := ""
	for _, l := range d.Lines {
		str += fmt.Sprint(l, "\n")
	}
	str += "\n"
	return str
}

// getTokens loads all text from the file at `path`
// loading a single rune at a time.
// Tokens are manually split on newlines and spaces
// and an EOL is added at the end
// The result is a slice of tokens with EOLs that can be
// used to separate lines
func getTokens(rc io.ReadCloser) ([]Token, error) {
	scanner := bufio.NewScanner(rc)
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
	lineNum := 1
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
					t := Token{
						Indent:  s.indented,
						Quote:   s.tokenQuoted,
						Comment: s.inComment,
						LineNum: lineNum,
						Text:    s.current,
					}
					tokens = append(tokens, t)
					s = stateType{} // reset state
				}
				// insert EOL so that subsequent funcs can split lines
				if isEOL {
					tokens = append(tokens, Token{
						EOL:     true,
						LineNum: lineNum,
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
		} else {
			if r == ";" || (r == "*" && onNewline) {
				// * only counts as a comment if it's the first rune on a line
				s.inComment = true
			}
			s.current += r
		}
		s.prev = r
		if isEOL {
			lineNum++
		}
	}
	if err := scanner.Err(); err != nil {
		// havent seen this yet, lets be loud about it!
		panic(fmt.Errorf("in getTokens: %w", err))
	}
	// manually added to make subsequent funcs lives easier
	tokens = append(tokens, Token{
		LineNum: lineNum,
		EOL:     true,
	})
	debugTokens(tokens)
	rc.Close()
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
			// as they end directives
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
		} else if line.Tokens[0].Text == string(dirOption) {
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
