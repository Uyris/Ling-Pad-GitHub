package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// ==================== Token ====================

// Tipos de token
const (
	INT       = "INT"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	XOR       = "XOR"
	MULT      = "MULT"
	DIV       = "DIV"
	OPEN_PAR  = "OPEN_PAR"
	CLOSE_PAR = "CLOSE_PAR"
	EOF       = "EOF"
)

// Token representa um token léxico com tipo e valor
type Token struct {
	tokenType string
	value     interface{}
}

// ==================== Lexer ====================

// Lexer realiza a análise léxica do código-fonte
type Lexer struct {
	source   string
	position int
	next     Token
}

// NewLexer cria um novo Lexer a partir do código-fonte
func NewLexer(source string) *Lexer {
	return &Lexer{
		source:   source,
		position: 0,
	}
}

// SelectNext lê o próximo token e atualiza o atributo next
func (l *Lexer) SelectNext() {
	// Ignora espaços em branco
	for l.position < len(l.source) && unicode.IsSpace(rune(l.source[l.position])) {
		l.position++
	}

	// Verifica se chegou ao final da string
	if l.position >= len(l.source) {
		l.next = Token{tokenType: EOF, value: ""}
		return
	}

	currentChar := rune(l.source[l.position])

	if currentChar == '+' {
		l.next = Token{tokenType: PLUS, value: "+"}
		l.position++
	} else if currentChar == '-' {
		l.next = Token{tokenType: MINUS, value: "-"}
		l.position++
	} else if currentChar == '^' {
		l.next = Token{tokenType: XOR, value: "^"}
		l.position++
	} else if currentChar == '*' {
		l.next = Token{tokenType: MULT, value: "*"}
		l.position++
	} else if currentChar == '/' {
		l.next = Token{tokenType: DIV, value: "/"}
		l.position++
	} else if currentChar == '(' {
		l.next = Token{tokenType: OPEN_PAR, value: "("}
		l.position++
	} else if currentChar == ')' {
		l.next = Token{tokenType: CLOSE_PAR, value: ")"}
		l.position++
	} else if unicode.IsDigit(currentChar) {
		numStr := string(currentChar)
		l.position++
		for l.position < len(l.source) && unicode.IsDigit(rune(l.source[l.position])) {
			numStr += string(l.source[l.position])
			l.position++
		}
		num, err := strconv.Atoi(numStr)
		if err != nil {
			panic("[Lexer] Erro ao converter número: " + numStr)
		}
		l.next = Token{tokenType: INT, value: num}
	} else {
		panic("[Lexer] Invalid Symbol: " + string(currentChar))
	}
}

// ==================== Parser ====================

// Parser realiza a análise sintática consumindo tokens do Lexer
type Parser struct{}

// lexer é o atributo estático (variável de pacote) do Parser
var parserLexer *Lexer


func ParseExpression() int {
	result := ParseTerm()

	// Enquanto o próximo token for PLUS, MINUS ou XOR
	for parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS || parserLexer.next.tokenType == XOR {
		op := parserLexer.next.tokenType
		parserLexer.SelectNext()

		rightValue := ParseTerm()

		if op == PLUS {
			result += rightValue
		} else if op == MINUS {
			result -= rightValue
		} else {
			result ^= rightValue
		}
	}

	return result
}

func ParseTerm() int {
	result := ParseFactor()

	// Enquanto o próximo token for MULT ou DIV
	for parserLexer.next.tokenType == MULT || parserLexer.next.tokenType == DIV {
		op := parserLexer.next.tokenType
		parserLexer.SelectNext()

		rightValue := ParseFactor()

		if op == MULT {
			result *= rightValue
		} else {
			if rightValue == 0 {
				panic("[Parser] Division by zero")
			}
			result /= rightValue
		}
	}

	return result
}

func ParseFactor() int {
	// Parênteses
	if parserLexer.next.tokenType == OPEN_PAR {
		parserLexer.SelectNext()
		result := ParseExpression()
		if parserLexer.next.tokenType != CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()
		return result
	}

	// Operadores unários (permite encadeamento como --5 ou ++3)
	if parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS {
		op := parserLexer.next.tokenType
		parserLexer.SelectNext()
		result := ParseFactor()
		if op == MINUS {
			result = -result
		}
		return result
	}

	// Número
	if parserLexer.next.tokenType == INT {
		result := parserLexer.next.value.(int)
		parserLexer.SelectNext()
		return result
	}

	panic("[Parser] Unexpected token in factor: " + parserLexer.next.tokenType)
}

// Run é o ponto de entrada do Parser. Recebe o código-fonte, inicializa o Lexer e retorna o resultado.
func Run(code string) int {
	parserLexer = NewLexer(code)
	parserLexer.SelectNext()

	result := ParseExpression()

	// Verifica se consumiu toda a entrada
	if parserLexer.next.tokenType != EOF {
		panic("[Parser] Unexpected token after expression: " + parserLexer.next.tokenType)
	}

	return result
}

// ==================== Main ====================

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "[Main] Nenhum argumento fornecido. Uso: main <expressão>")
		os.Exit(1)
	}

	input := os.Args[1]
	result := Run(input)
	fmt.Println(result)
}
