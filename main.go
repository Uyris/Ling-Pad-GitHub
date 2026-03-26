package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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
	ASSIGN    = "ASSIGN"
	END       = "END"
	PRINT     = "PRINT"
	IDEN      = "IDEN"
	EOF       = "EOF"
)

// Token representa um token léxico com tipo e valor
type Token struct {
	tokenType string
	value     interface{}
}

// ==================== PrePro ====================

// PrePro realiza o pré-processamento do código-fonte
type PrePro struct{}

// Filter remove comentários inline do código
func (p *PrePro) Filter(code string) string {
	// Remove tudo entre "//" e "\n", mantendo o "\n"
	re := regexp.MustCompile(`//[^\n]*`)
	return re.ReplaceAllString(code, "")
}

// ==================== Variable ====================

// Variable representa uma variável com seu valor
type Variable struct {
	value int
}

// NewVariable cria uma nova variável
func NewVariable(value int) *Variable {
	return &Variable{value: value}
}

// ==================== SymbolTable ====================

// SymbolTable armazena variáveis e seus valores
type SymbolTable struct {
	table map[string]*Variable
}

// NewSymbolTable cria uma nova tabela de símbolos
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		table: make(map[string]*Variable),
	}
}

// Get retorna o valor de uma variável
func (st *SymbolTable) Get(name string) int {
	if variable, exists := st.table[name]; exists {
		return variable.value
	}
	panic("[Semantic] Variável não definida: " + name)
}

// Set adiciona ou atualiza uma variável
func (st *SymbolTable) Set(name string, value int) {
	st.table[name] = NewVariable(value)
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
	} else if currentChar == '=' {
		l.next = Token{tokenType: ASSIGN, value: "="}
		l.position++
	} else if currentChar == ';' {
		l.next = Token{tokenType: END, value: ";"}
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
	} else if unicode.IsLetter(currentChar) || currentChar == '_' {
		// Identifiers e palavras reservadas
		if currentChar == '_' {
			panic("[Lexer] Identificador não pode começar com '_'")
		}
		identStr := string(currentChar)
		l.position++
		for l.position < len(l.source) && (unicode.IsLetter(rune(l.source[l.position])) || unicode.IsDigit(rune(l.source[l.position])) || rune(l.source[l.position]) == '_') {
			identStr += string(l.source[l.position])
			l.position++
		}

		// Verificar se é "println" seguido de "!"
		if identStr == "println" && l.position < len(l.source) && rune(l.source[l.position]) == '!' {
			l.position++ // Consumir o "!"
			l.next = Token{tokenType: PRINT, value: "println"}
		} else if identStr == "println" {
			// "println" sem "!" é um identificador normal
			l.next = Token{tokenType: IDEN, value: identStr}
		} else {
			l.next = Token{tokenType: IDEN, value: identStr}
		}
	} else {
		panic("[Lexer] Invalid Symbol: " + string(currentChar))
	}
}

// ==================== AST Nodes ====================

// Node é a interface base para todos os nós da AST
type Node interface {
	Evaluate(st *SymbolTable) int
}

// IntVal representa um valor inteiro (nó terminal, sem filhos)
type IntVal struct {
	value    int
	children []Node
}

func NewIntVal(value int) *IntVal {
	return &IntVal{value: value, children: []Node{}}
}

func (n *IntVal) Evaluate(st *SymbolTable) int {
	return n.value
}

// UnOp representa uma operação unária (1 filho)
type UnOp struct {
	value    string
	children []Node
}

func NewUnOp(operator string, operand Node) *UnOp {
	return &UnOp{value: operator, children: []Node{operand}}
}

func (n *UnOp) Evaluate(st *SymbolTable) int {
	childResult := n.children[0].Evaluate(st)
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

func (n *BinOp) Evaluate(st *SymbolTable) int {
	leftResult := n.children[0].Evaluate(st)
	rightResult := n.children[1].Evaluate(st)

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

// Identifier representa uma variável
type Identifier struct {
	name     string
	children []Node
}

func NewIdentifier(name string) *Identifier {
	return &Identifier{name: name, children: []Node{}}
}

func (n *Identifier) Evaluate(st *SymbolTable) int {
	return st.Get(n.name)
}

// Print representa a impressão de um valor
type Print struct {
	children []Node
}

func NewPrint(expr Node) *Print {
	return &Print{children: []Node{expr}}
}

func (n *Print) Evaluate(st *SymbolTable) int {
	result := n.children[0].Evaluate(st)
	fmt.Println(result)
	return 0
}

// Assignment representa uma atribuição de variável
type Assignment struct {
	children []Node
}

func NewAssignment(identifier Node, expr Node) *Assignment {
	return &Assignment{children: []Node{identifier, expr}}
}

func (n *Assignment) Evaluate(st *SymbolTable) int {
	identNode := n.children[0].(*Identifier)
	value := n.children[1].Evaluate(st)
	st.Set(identNode.name, value)
	return 0
}

// Block representa um bloco de instruções
type Block struct {
	children []Node
}

func NewBlock() *Block {
	return &Block{children: []Node{}}
}

func (b *Block) AddChild(node Node) {
	b.children = append(b.children, node)
}

func (n *Block) Evaluate(st *SymbolTable) int {
	for _, child := range n.children {
		child.Evaluate(st)
	}
	return 0
}

// NoOp representa uma operação vazia (dummy)
type NoOp struct {
	children []Node
}

func NewNoOp() *NoOp {
	return &NoOp{children: []Node{}}
}

func (n *NoOp) Evaluate(st *SymbolTable) int {
	return 0
}

// ==================== Parser ====================

// Parser realiza a análise sintática consumindo tokens do Lexer
type Parser struct{}

// lexer é o atributo estático (variável de pacote) do Parser
var parserLexer *Lexer

func ParseProgram() Node {
	block := NewBlock()

	for parserLexer.next.tokenType != EOF {
		stmt := ParseStatement()
		block.AddChild(stmt)
	}

	return block
}

func ParseStatement() Node {
	// Atribuição: IDENTIFIER = EXPRESSION ;
	if parserLexer.next.tokenType == IDEN {
		name := parserLexer.next.value.(string)
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected ASSIGN")
		}
		parserLexer.SelectNext()

		expr := ParseExpression()

		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;)")
		}
		parserLexer.SelectNext()

		return NewAssignment(NewIdentifier(name), expr)
	}

	// Impressão: PRINT ( EXPRESSION ) ;
	if parserLexer.next.tokenType == PRINT {
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		expr := ParseExpression()

		if parserLexer.next.tokenType != CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;)")
		}
		parserLexer.SelectNext()

		return NewPrint(expr)
	}

	// Linha vazia: ;
	if parserLexer.next.tokenType == END {
		parserLexer.SelectNext()
		return NewNoOp()
	}

	panic("[Parser] Unexpected token in statement: " + parserLexer.next.tokenType)
}

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

	// Operadores unários
	if parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		operand := ParseFactor()
		return NewUnOp(op, operand)
	}

	// Identificador
	if parserLexer.next.tokenType == IDEN {
		name := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		return NewIdentifier(name)
	}

	// Número
	if parserLexer.next.tokenType == INT {
		value := parserLexer.next.value.(int)
		parserLexer.SelectNext()
		return NewIntVal(value)
	}

	panic("[Parser] Unexpected token in factor: " + parserLexer.next.tokenType)
}

// Run é o ponto de entrada do Parser. Retorna a raiz da AST.
func Run(code string) Node {
	parserLexer = NewLexer(code)
	parserLexer.SelectNext()

	result := ParseProgram()

	if parserLexer.next.tokenType != EOF {
		panic("[Parser] Unexpected token after program: " + parserLexer.next.tokenType)
	}

	return result
}

// ==================== Main ====================

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "[Main] Nenhum arquivo fornecido. Uso: main <arquivo>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Ler arquivo
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Main] Erro ao ler arquivo: "+filename)
		os.Exit(1)
	}

	code := string(content) + "\n"

	// Pré-processamento: remover comentários
	prePro := &PrePro{}
	code = prePro.Filter(code)

	// Análise sintática
	ast := Run(code)

	// Criar tabela de símbolos e executar
	st := NewSymbolTable()
	ast.Evaluate(st)
}
