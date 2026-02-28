package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Token types
const (
	INT   = "INT"
	PLUS  = "PLUS"
	MINUS = "MINUS"
	EOF   = "EOF"
)

// Token represents a lexical token with a type and value.
type Token struct {
	Type  string
	Value interface{} // int for INT tokens, string for others
}

// Lexer reads the source code and produces tokens on demand.
type Lexer struct {
	source   string
	position int
	next     Token
}

// selectNext reads the next token from the source and updates the next attribute.
func (l *Lexer) selectNext() {
	// Skip whitespace
	for l.position < len(l.source) && l.source[l.position] == ' ' {
		l.position++
	}

	if l.position >= len(l.source) {
		l.next = Token{Type: EOF, Value: ""}
		return
	}

	ch := l.source[l.position]

	if ch == '+' {
		l.next = Token{Type: PLUS, Value: "+"}
		l.position++
		return
	}

	if ch == '-' {
		l.next = Token{Type: MINUS, Value: "-"}
		l.position++
		return
	}

	if unicode.IsDigit(rune(ch)) {
		var sb strings.Builder
		sb.WriteByte(ch)
		l.position++
		for l.position < len(l.source) && unicode.IsDigit(rune(l.source[l.position])) {
			sb.WriteByte(l.source[l.position])
			l.position++
		}
		num, err := strconv.Atoi(sb.String())
		if err != nil {
			panic("[Lexer] Failed to parse number: " + sb.String())
		}
		l.next = Token{Type: INT, Value: num}
		return
	}

	panic("[Lexer] Invalid symbol " + string(ch))
}

// lexer is the Parser's static reference to the Lexer instance.
var lexer *Lexer

// parseExpression consumes tokens from the Lexer and evaluates the expression.
// Grammar: expression = INT ( (PLUS | MINUS) INT )*
func parseExpression() int {
	if lexer.next.Type != INT {
		panic("[Parser] Unexpected token " + lexer.next.Type)
	}
	result := lexer.next.Value.(int)
	lexer.selectNext()

	for lexer.next.Type == PLUS || lexer.next.Type == MINUS {
		op := lexer.next.Type
		lexer.selectNext()
		if lexer.next.Type != INT {
			panic("[Parser] Unexpected token " + lexer.next.Type)
		}
		if op == PLUS {
			result += lexer.next.Value.(int)
		} else {
			result -= lexer.next.Value.(int)
		}
		lexer.selectNext()
	}

	return result
}

// run initializes the Lexer with the given source code and returns the evaluated result.
func run(code string) int {
	lexer = &Lexer{source: code, position: 0}
	lexer.selectNext()
	result := parseExpression()
	if lexer.next.Type != EOF {
		panic("[Parser] Unexpected token " + lexer.next.Type)
	}
	return result
}

func main() {
	if len(os.Args) < 2 {
		panic("Nenhum argumento fornecido. Por favor, forneça uma expressão matemática como argumento.")
	}

	input := os.Args[1]

	if len(input) == 0 {
		panic("Expressão vazia. Por favor, forneça uma expressão matemática válida.")
	}

	fmt.Println(run(input))
}
