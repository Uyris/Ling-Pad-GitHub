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
	FACT      = "FACT"
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
	} else if currentChar == '!' {
		l.next = Token{tokenType: FACT, value: "!"}
		l.position++
	} else {
		panic("[Lexer] Invalid Symbol: " + string(currentChar))
	}
}

// ==================== AST Nodes ====================

// Node é a interface base para todos os nós da AST
type Node interface {
	Evaluate() int
}

// IntVal representa um valor inteiro (nó terminal, sem filhos)
type IntVal struct {
	value    int
	children []Node
}

type Fact struct {
	value    int
	children []Node
}

func NewIntVal(value int) *IntVal {
	return &IntVal{value: value, children: []Node{}}
}

func (n *IntVal) Evaluate() int {
	return n.value
}

// UnOp representa uma operação unária (1 filho)
type UnOp struct {
	value    string
	children []Node
}
func NewFact(value int) *Fact {
	return &Fact{value: value, children: []Node{}}
}

func (n *Fact) Evaluate() int {
	if n.value < 0 {
		panic("[Semantic] Factorial is not defined for negative numbers: " + strconv.Itoa(n.value))
	}
	result := 1
	for i := 2; i <= n.value; i++ {
		result *= i
	}
	return result
}

func NewUnOp(operator string, operand Node) *UnOp {
	return &UnOp{value: operator, children: []Node{operand}}
}

func (n *UnOp) Evaluate() int {
	childResult := n.children[0].Evaluate()
	if n.value == "-" {
		return -childResult
	} else if n.value == "+" {
		return childResult
	}
	panic("[Semantic] Unknown unary operator: " + n.value)
}

// BinOp representa uma operação binária (2 filhos)
type BinOp struct {
	value    string
	children []Node
}

func NewBinOp(operator string, left Node, right Node) *BinOp {
	return &BinOp{value: operator, children: []Node{left, right}}
}

func (n *BinOp) Evaluate() int {
	leftResult := n.children[0].Evaluate()
	rightResult := n.children[1].Evaluate()

	switch n.value {
	case "+":
		return leftResult + rightResult
	case "-":
		return leftResult - rightResult
	case "^":
		return leftResult ^ rightResult
	case "*":
		return leftResult * rightResult
	case "/":
		if rightResult == 0 {
			panic("[Semantic] Division by zero")
		}
		return leftResult / rightResult
	}
	panic("[Semantic] Unknown binary operator: " + n.value)
}

// ==================== Parser ====================

// Parser realiza a análise sintática consumindo tokens do Lexer
type Parser struct{}

// lexer é o atributo estático (variável de pacote) do Parser
var parserLexer *Lexer

func ParseExpression() Node {
	result := ParseTerm()

	for parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS || parserLexer.next.tokenType == XOR {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		right := ParseTerm()
		result = NewBinOp(op, result, right)
	}

	return result
}

func ParseTerm() Node {
	result := ParseFactor()

	for parserLexer.next.tokenType == MULT || parserLexer.next.tokenType == DIV {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		right := ParseFactor()
		result = NewBinOp(op, result, right)
	}

	return result
}

func ParseFactor() Node {
	var result Node

	// Parênteses
	if parserLexer.next.tokenType == OPEN_PAR {
		parserLexer.SelectNext()
		result := ParseExpression()
		if parserLexer.next.tokenType != CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()
		return result	
		
	} else if parserLexer.next.tokenType == INT { // Número inteiro
		value := parserLexer.next.value.(int)
		parserLexer.SelectNext()
		result = NewIntVal(value)

	} else if parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS { // Operador unário
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		operand := ParseFactor()
		return NewUnOp(op, operand)

	} else { 
		panic("[Parser] Unexpected token in factor: " + parserLexer.next.tokenType)
	}

	if parserLexer.next.tokenType == FACT { // Fatorial
		parserLexer.SelectNext()
		result = NewFact(result.Evaluate())
	}

	return result;

}

// Run é o ponto de entrada do Parser. Retorna a raiz da AST.
func Run(code string) Node {
	parserLexer = NewLexer(code)
	parserLexer.SelectNext()

	result := ParseExpression()

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
	ast := Run(input)
	result := ast.Evaluate()
	fmt.Println(result)
}
