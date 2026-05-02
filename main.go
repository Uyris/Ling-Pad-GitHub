package main

import (
	"fmt"
	"io/ioutil"
	"math"
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
	POW       = "POW"
	AND       = "AND"
	OR        = "OR"
	NOT       = "NOT"
	EQ        = "EQ"
	GT        = "GT"
	LT        = "LT"
	OPEN_PAR  = "OPEN_PAR"
	CLOSE_PAR = "CLOSE_PAR"
	OPEN_BRA  = "OPEN_BRA"
	CLOSE_BRA = "CLOSE_BRA"
	ASSIGN    = "ASSIGN"
	END       = "END"
	IF        = "IF"
	WHILE     = "WHILE"
	FOR       = "FOR"
	ELSE      = "ELSE"
	PRINT     = "PRINT"
	READ      = "READ"
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
	} else if currentChar == '*' && l.position+1 < len(l.source) && l.source[l.position+1] == '*' {
		l.next = Token{tokenType: POW, value: "**"}
		l.position += 2
	} else if currentChar == '*' {
		l.next = Token{tokenType: MULT, value: "*"}
		l.position++
	} else if currentChar == '/' {
		l.next = Token{tokenType: DIV, value: "/"}
		l.position++
	} else if currentChar == '&' && l.position+1 < len(l.source) && l.source[l.position+1] == '&' {
		l.next = Token{tokenType: AND, value: "&&"}
		l.position += 2
	} else if currentChar == '|' && l.position+1 < len(l.source) && l.source[l.position+1] == '|' {
		l.next = Token{tokenType: OR, value: "||"}
		l.position += 2
	} else if currentChar == '!' {
		l.next = Token{tokenType: NOT, value: "!"}
		l.position++
	} else if currentChar == '=' && l.position+1 < len(l.source) && l.source[l.position+1] == '=' {
		l.next = Token{tokenType: EQ, value: "=="}
		l.position += 2
	} else if currentChar == '=' {
		l.next = Token{tokenType: ASSIGN, value: "="}
		l.position++
	} else if currentChar == '>' {
		l.next = Token{tokenType: GT, value: ">"}
		l.position++
	} else if currentChar == '<' {
		l.next = Token{tokenType: LT, value: "<"}
		l.position++
	} else if currentChar == '(' {
		l.next = Token{tokenType: OPEN_PAR, value: "("}
		l.position++
	} else if currentChar == ')' {
		l.next = Token{tokenType: CLOSE_PAR, value: ")"}
		l.position++
	} else if currentChar == '{' {
		l.next = Token{tokenType: OPEN_BRA, value: "{"}
		l.position++
	} else if currentChar == '}' {
		l.next = Token{tokenType: CLOSE_BRA, value: "}"}
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

		// Verificar palavras reservadas
		if identStr == "println" && l.position < len(l.source) && rune(l.source[l.position]) == '!' {
			l.position++ // Consumir o "!"
			l.next = Token{tokenType: PRINT, value: "println"}
		} else if identStr == "scanln" && l.position < len(l.source) && rune(l.source[l.position]) == '!' {
			l.position++ // Consumir o "!"
			l.next = Token{tokenType: READ, value: "scanln"}
		} else if identStr == "if" {
			l.next = Token{tokenType: IF, value: "if"}
		} else if identStr == "while" {
			l.next = Token{tokenType: WHILE, value: "while"}
		} else if identStr == "for" {
			l.next = Token{tokenType: FOR, value: "for"}
		} else if identStr == "else" {
			l.next = Token{tokenType: ELSE, value: "else"}
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
	} else if n.value == "!" {
		if childResult == 0 {
			return 1
		}
		return 0
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
	if n.value == "&&" {
		leftResult := n.children[0].Evaluate(st)
		if leftResult == 0 {
			return 0
		}
		rightResult := n.children[1].Evaluate(st)
		if rightResult != 0 {
			return 1
		}
		return 0
	}

	if n.value == "||" {
		leftResult := n.children[0].Evaluate(st)
		if leftResult != 0 {
			return 1
		}
		rightResult := n.children[1].Evaluate(st)
		if rightResult != 0 {
			return 1
		}
		return 0
	}

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
	case "**":
		return int(math.Pow(float64(leftResult), float64(rightResult)))
	case "==":
		if leftResult == rightResult {
			return 1
		}
		return 0
	case ">":
		if leftResult > rightResult {
			return 1
		}
		return 0
	case "<":
		if leftResult < rightResult {
			return 1
		}
		return 0
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

// IfNode representa uma estrutura condicional if/else
type IfNode struct {
	children []Node
}

func NewIfNode(condition Node, thenBranch Node, elseBranch Node) *IfNode {
	children := []Node{condition, thenBranch}
	if elseBranch != nil {
		children = append(children, elseBranch)
	}
	return &IfNode{children: children}
}

func (n *IfNode) Evaluate(st *SymbolTable) int {
	condition := n.children[0].Evaluate(st)
	if condition != 0 {
		n.children[1].Evaluate(st)
	} else if len(n.children) == 3 {
		n.children[2].Evaluate(st)
	}
	return 0
}

// WhileNode representa uma estrutura de repetição while
type WhileNode struct {
	children []Node
}

func NewWhileNode(condition Node, body Node) *WhileNode {
	return &WhileNode{children: []Node{condition, body}}
}

func (n *WhileNode) Evaluate(st *SymbolTable) int {
	for n.children[0].Evaluate(st) != 0 {
		n.children[1].Evaluate(st)
	}
	return 0
}

// ForNode representa uma estrutura de repetição for
type ForNode struct {
	children []Node // [init, condition, increment, body]
}

func NewForNode(init Node, condition Node, increment Node, body Node) *ForNode {
	return &ForNode{children: []Node{init, condition, increment, body}}
}

func (n *ForNode) Evaluate(st *SymbolTable) int {
	n.children[0].Evaluate(st) // init
	for n.children[1].Evaluate(st) != 0 { // condition
		n.children[3].Evaluate(st) // body
		n.children[2].Evaluate(st) // increment
	}
	return 0
}

// IfExprNode representa uma expressão condicional if/else (ternária)
type IfExprNode struct {
	children []Node // [condition, thenExpr, elseExpr]
}

func NewIfExprNode(condition Node, thenExpr Node, elseExpr Node) *IfExprNode {
	return &IfExprNode{children: []Node{condition, thenExpr, elseExpr}}
}

func (n *IfExprNode) Evaluate(st *SymbolTable) int {
	condition := n.children[0].Evaluate(st)
	if condition != 0 {
		return n.children[1].Evaluate(st)
	}
	return n.children[2].Evaluate(st)
}

// ReadNode representa a leitura de inteiro do terminal
type ReadNode struct {
	children []Node
}

func NewReadNode() *ReadNode {
	return &ReadNode{children: []Node{}}
}

func (n *ReadNode) Evaluate(st *SymbolTable) int {
	var value int
	_, err := fmt.Fscan(os.Stdin, &value)
	if err != nil {
		panic("[Semantic] Failed to read integer input")
	}
	return value
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
	lastValue := 0
	for _, child := range n.children {
		lastValue = child.Evaluate(st)
	}
	return lastValue
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

// ExprStmt representa uma expressão como statement
type ExprStmt struct {
	children []Node
}

func NewExprStmt(expr Node) *ExprStmt {
	return &ExprStmt{children: []Node{expr}}
}

func (n *ExprStmt) Evaluate(st *SymbolTable) int {
	return n.children[0].Evaluate(st)
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

func ParseBlock() Node {
	if parserLexer.next.tokenType != OPEN_BRA {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_BRA")
	}
	parserLexer.SelectNext()

	block := NewBlock()
	for parserLexer.next.tokenType != CLOSE_BRA {
		if parserLexer.next.tokenType == EOF {
			panic("[Parser] Unexpected EOF, expected CLOSE_BRA")
		}
		stmt := ParseStatement()
		block.AddChild(stmt)
	}

	parserLexer.SelectNext()
	return block
}

func ParseStatement() Node {
	// Bloco: { STATEMENT* }
	if parserLexer.next.tokenType == OPEN_BRA {
		return ParseBlock()
	}

	// WHILE: while ( BOOLEXPRESSION ) STATEMENT
	if parserLexer.next.tokenType == WHILE {
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		condition := ParseBoolExpression()

		if parserLexer.next.tokenType != CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		body := ParseStatement()
		return NewWhileNode(condition, body)
	}

	// FOR: for ( STATEMENT ; BOOLEXPRESSION ; STATEMENT ) STATEMENT
	if parserLexer.next.tokenType == FOR {
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		// Parse init (pode ser vazio ou uma atribuição)
		var init Node
		if parserLexer.next.tokenType == END {
			init = NewNoOp()
			parserLexer.SelectNext()
		} else {
			init = ParseForInit()
			if parserLexer.next.tokenType != END {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;)")
			}
			parserLexer.SelectNext()
		}

		// Parse condition
		var condition Node
		if parserLexer.next.tokenType == END {
			condition = NewIntVal(1) // Condição sempre verdadeira se vazia
			parserLexer.SelectNext()
		} else {
			condition = ParseBoolExpression()
			if parserLexer.next.tokenType != END {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;)")
			}
			parserLexer.SelectNext()
		}

		// Parse increment
		var increment Node
		if parserLexer.next.tokenType == CLOSE_PAR {
			increment = NewNoOp()
			parserLexer.SelectNext()
		} else {
			increment = ParseForIncrement()
			if parserLexer.next.tokenType != CLOSE_PAR {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
			}
			parserLexer.SelectNext()
		}

		body := ParseStatement()
		return NewForNode(init, condition, increment, body)
	}

	// Impressão: PRINT ( BOOLEXPRESSION ) ;
	if parserLexer.next.tokenType == PRINT {
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()

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

	// Expression statement (para if expressions que retornam valores)
	// Tenta parsear como expressão se for um token válido para começar expressão
	if parserLexer.next.tokenType == INT ||
		parserLexer.next.tokenType == IDEN ||
		parserLexer.next.tokenType == OPEN_PAR ||
		parserLexer.next.tokenType == READ ||
		parserLexer.next.tokenType == NOT ||
		parserLexer.next.tokenType == MINUS ||
		parserLexer.next.tokenType == PLUS ||
		parserLexer.next.tokenType == IF {

		expr := ParseBoolExpression()

		// Verifica se é uma atribuição: Identifier = expression
		if iden, ok := expr.(*Identifier); ok && parserLexer.next.tokenType == ASSIGN {
			parserLexer.SelectNext()
			rightExpr := ParseBoolExpression()

			if parserLexer.next.tokenType == END {
				parserLexer.SelectNext()
			}

			return NewAssignment(iden, rightExpr)
		}

		// Se houver `;`, consome; senão, é a última expressão do bloco
		if parserLexer.next.tokenType == END {
			parserLexer.SelectNext()
		}

		return NewExprStmt(expr)
	}

	panic("[Parser] Unexpected token in statement: " + parserLexer.next.tokenType)
}

func ParseBoolExpression() Node {
	result := ParseBoolTerm()

	for parserLexer.next.tokenType == OR {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		right := ParseBoolTerm()
		result = NewBinOp(op, result, right)
	}

	return result
}

func ParseBoolTerm() Node {
	result := ParseRelExpression()

	for parserLexer.next.tokenType == AND {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		right := ParseRelExpression()
		result = NewBinOp(op, result, right)
	}

	return result
}

func ParseRelExpression() Node {
	left := ParseExpression()

	if parserLexer.next.tokenType == EQ || parserLexer.next.tokenType == GT || parserLexer.next.tokenType == LT {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		right := ParseExpression()
		return NewBinOp(op, left, right)
	}

	return left
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
	result := ParseUnary()

	for parserLexer.next.tokenType == MULT || parserLexer.next.tokenType == DIV {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		right := ParseUnary()
		result = NewBinOp(op, result, right)
	}

	return result
}

func ParseUnary() Node {
	if parserLexer.next.tokenType == PLUS || parserLexer.next.tokenType == MINUS || parserLexer.next.tokenType == NOT {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		operand := ParseUnary()
		return NewUnOp(op, operand)
	}

	return ParsePower()
}

func ParsePower() Node {
	base := ParseFactor()

	if parserLexer.next.tokenType == POW {
		op := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		exponent := ParseUnary()
		return NewBinOp(op, base, exponent)
	}

	return base
}

func ParseFactor() Node {
	// If expression: if BOOLEXPRESSION { BOOLEXPRESSION } else { BOOLEXPRESSION }
	if parserLexer.next.tokenType == IF {
		parserLexer.SelectNext()

		condition := ParseBoolExpression()

		// Then branch (bloco com chaves)
		if parserLexer.next.tokenType != OPEN_BRA {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_BRA")
		}
		thenBranch := ParseBlock()

		// Else branch (obrigatório para if expression)
		if parserLexer.next.tokenType != ELSE {
			panic("[Parser] If expression must have else clause")
		}
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != OPEN_BRA {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_BRA")
		}
		elseBranch := ParseBlock()

		return NewIfExprNode(condition, thenBranch, elseBranch)
	}

	// Parênteses
	if parserLexer.next.tokenType == OPEN_PAR {
		parserLexer.SelectNext()
		result := ParseBoolExpression()
		if parserLexer.next.tokenType != CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()
		return result
	}

	// Leitura: scanln!()
	if parserLexer.next.tokenType == READ {
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		return NewReadNode()
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

func ParseForInit() Node {
	// Atribuição: IDENTIFIER = BOOLEXPRESSION
	if parserLexer.next.tokenType == IDEN {
		name := parserLexer.next.value.(string)
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected ASSIGN")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()
		return NewAssignment(NewIdentifier(name), expr)
	}

	panic("[Parser] Unexpected token in for init: " + parserLexer.next.tokenType)
}

func ParseForIncrement() Node {
	// Atribuição: IDENTIFIER = BOOLEXPRESSION
	if parserLexer.next.tokenType == IDEN {
		name := parserLexer.next.value.(string)
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected ASSIGN")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()
		return NewAssignment(NewIdentifier(name), expr)
	}

	panic("[Parser] Unexpected token in for increment: " + parserLexer.next.tokenType)
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
