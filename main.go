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
	INT   = "INT"
	PLUS  = "PLUS"
	MINUS = "MINUS"
	EOF   = "EOF"
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

// ParseExpression consome tokens e analisa a expressão conforme o diagrama sintático
func ParseExpression() int {
	// O primeiro token deve ser um número
	if parserLexer.next.tokenType != INT {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected INT")
	}

	result := parserLexer.next.value.(int)
	parserLexer.SelectNext()

	// Enquanto o próximo token for PLUS ou MINUS
	for parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS {
		op := parserLexer.next.tokenType
		parserLexer.SelectNext()

		// O próximo token após o operador deve ser um número
		if parserLexer.next.tokenType != INT {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected INT after operator")
		}

		num := parserLexer.next.value.(int)

		if op == PLUS {
			result += num
		} else {
			result -= num
		}

		parserLexer.SelectNext()
	}

	return result
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
