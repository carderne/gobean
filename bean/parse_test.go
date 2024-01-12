package bean

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetTokens(t *testing.T) {
	// normal text should be parsed to tokens
	text := `
** Transactions
2023-02-01 * "Salary"
  Assets:Bank                          1000 GBP
  Income:Job
`
	rc := io.NopCloser(strings.NewReader(text))
	got, _ := getTokens(rc)
	want := []Token{
		{LineNum: 1, EOL: true},
		{LineNum: 2, Comment: true, Text: "** Transactions"},
		{LineNum: 2, EOL: true},
		{LineNum: 3, Text: "2023-02-01"},
		{LineNum: 3, Text: "*"},
		{LineNum: 3, Quote: true, Text: "Salary"},
		{LineNum: 3, EOL: true},
		{LineNum: 4, Indent: true, Text: "Assets:Bank"},
		{LineNum: 4, Text: "1000"},
		{LineNum: 4, Text: "GBP"},
		{LineNum: 4, EOL: true},
		{LineNum: 5, Indent: true, Text: "Income:Job"},
		{LineNum: 5, EOL: true},
		{LineNum: 6, EOL: true},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
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

func TestMakeDirectives(t *testing.T) {
	// dangling indent should return an error
	text := `
  Income:Job
  `
	rc := io.NopCloser(strings.NewReader(text))
	tokens, _ := getTokens(rc)
	lines, _ := makeLines(tokens)
	_, err := makeDirectives(lines)
	if err == nil {
		t.Error("parse should fail with indented expression outside directive")
	}

	// tags should be ignored
	text = `
option "operating_currency" "GBP"
2023-02-01 * "Salary"
  tag:value
  Assets:Bank                          1000 GBP
  Income:Job
  `
	rc = io.NopCloser(strings.NewReader(text))
	tokens, _ = getTokens(rc)
	lines, _ = makeLines(tokens)
	got, _ := makeDirectives(lines)
	want := []Directive{{[]Line{
		{Tokens: []Token{
			{LineNum: 3, Text: "2023-02-01"},
			{LineNum: 3, Text: "*"},
			{LineNum: 3, Quote: true, Text: "Salary"},
		}},
		{Tokens: []Token{
			{LineNum: 5, Indent: true, Text: "Assets:Bank"},
			{LineNum: 5, Text: "1000"},
			{LineNum: 5, Text: "GBP"},
		}},
		{Tokens: []Token{
			{LineNum: 6, Indent: true, Text: "Income:Job"},
		}},
	}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}
