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
	BOOL      = "BOOL"
	STR       = "STR"
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
	COLON     = "COLON"
	END       = "END"
	IF        = "IF"
	WHILE     = "WHILE"
	ELSE      = "ELSE"
	LET       = "LET"
	MUT       = "MUT"
	TYPE      = "TYPE"
	PRINT     = "PRINT"
	READ      = "READ"
	IDEN      = "IDEN"
	EOF       = "EOF"
)

const (
	TYPE_I32  = "i32"
	TYPE_BOOL = "bool"
	TYPE_STR  = "str"
	TYPE_VOID = "void"
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
	value   interface{}
	varType string
	mutable bool
}

// NewVariable cria uma nova variável
func NewVariable(value interface{}, varType string, mutable bool) *Variable {
	return &Variable{value: value, varType: varType, mutable: mutable}
}

func (v *Variable) Clone() *Variable {
	return &Variable{value: v.value, varType: v.varType, mutable: v.mutable}
}

func NewVoidVariable() *Variable {
	return NewVariable(nil, TYPE_VOID, false)
}

func defaultValueForType(varType string) interface{} {
	switch varType {
	case TYPE_I32:
		return 0
	case TYPE_BOOL:
		return false
	case TYPE_STR:
		return ""
	default:
		panic("[Semantic] Unknown type: " + varType)
	}
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
func (st *SymbolTable) Get(name string) *Variable {
	if variable, exists := st.table[name]; exists {
		return variable
	}
	panic("[Semantic] Variável não definida: " + name)
}

func (st *SymbolTable) Exists(name string) bool {
	_, exists := st.table[name]
	return exists
}

// CreateVariable declara uma variável na tabela
func (st *SymbolTable) CreateVariable(name string, varType string, mutable bool) {
	if st.Exists(name) {
		panic("[Semantic] Variável já declarada: " + name)
	}
	st.table[name] = NewVariable(defaultValueForType(varType), varType, mutable)
}

// Initialize define o valor inicial no momento da declaração
func (st *SymbolTable) Initialize(name string, value *Variable) {
	variable := st.Get(name)
	if variable.varType != value.varType {
		panic("[Semantic] Type mismatch na inicialização de " + name)
	}
	variable.value = value.value
}

// Set atualiza uma variável já declarada
func (st *SymbolTable) Set(name string, value *Variable) {
	if !st.Exists(name) {
		panic("[Semantic] Variável não declarada: " + name)
	}

	variable := st.table[name]
	if !variable.mutable {
		panic("[Semantic] Variável imutável: " + name)
	}

	if variable.varType != value.varType {
		panic("[Semantic] Type mismatch em atribuição para " + name)
	}

	variable.value = value.value
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
	} else if currentChar == ':' {
		l.next = Token{tokenType: COLON, value: ":"}
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
	} else if currentChar == '"' {
		l.position++
		strValue := ""
		for l.position < len(l.source) && rune(l.source[l.position]) != '"' {
			strValue += string(l.source[l.position])
			l.position++
		}
		if l.position >= len(l.source) {
			panic("[Lexer] Unterminated string literal")
		}
		l.position++
		l.next = Token{tokenType: STR, value: strValue}
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
		} else if identStr == "else" {
			l.next = Token{tokenType: ELSE, value: "else"}
		} else if identStr == "let" {
			l.next = Token{tokenType: LET, value: "let"}
		} else if identStr == "mut" {
			l.next = Token{tokenType: MUT, value: "mut"}
		} else if identStr == "true" {
			l.next = Token{tokenType: BOOL, value: true}
		} else if identStr == "false" {
			l.next = Token{tokenType: BOOL, value: false}
		} else if identStr == TYPE_I32 || identStr == TYPE_BOOL || identStr == TYPE_STR {
			l.next = Token{tokenType: TYPE, value: identStr}
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
	Evaluate(st *SymbolTable) *Variable
}

// IntVal representa um valor inteiro (nó terminal, sem filhos)
type IntVal struct {
	value    int
	children []Node
}

func NewIntVal(value int) *IntVal {
	return &IntVal{value: value, children: []Node{}}
}

func (n *IntVal) Evaluate(st *SymbolTable) *Variable {
	return NewVariable(n.value, TYPE_I32, false)
}

// BoolVal representa um valor booleano
type BoolVal struct {
	value    bool
	children []Node
}

func NewBoolVal(value bool) *BoolVal {
	return &BoolVal{value: value, children: []Node{}}
}

func (n *BoolVal) Evaluate(st *SymbolTable) *Variable {
	return NewVariable(n.value, TYPE_BOOL, false)
}

// StringVal representa um valor string
type StringVal struct {
	value    string
	children []Node
}

func NewStringVal(value string) *StringVal {
	return &StringVal{value: value, children: []Node{}}
}

func (n *StringVal) Evaluate(st *SymbolTable) *Variable {
	return NewVariable(n.value, TYPE_STR, false)
}

// UnOp representa uma operação unária (1 filho)
type UnOp struct {
	value    string
	children []Node
}

func NewUnOp(operator string, operand Node) *UnOp {
	return &UnOp{value: operator, children: []Node{operand}}
}

func (n *UnOp) Evaluate(st *SymbolTable) *Variable {
	childResult := n.children[0].Evaluate(st)
	if n.value == "-" {
		if childResult.varType != TYPE_I32 {
			panic("[Semantic] Unary '-' requires i32")
		}
		return NewVariable(-(childResult.value.(int)), TYPE_I32, false)
	} else if n.value == "+" {
		if childResult.varType != TYPE_I32 {
			panic("[Semantic] Unary '+' requires i32")
		}
		return NewVariable(childResult.value.(int), TYPE_I32, false)
	} else if n.value == "!" {
		if childResult.varType != TYPE_BOOL {
			panic("[Semantic] Unary '!' requires bool")
		}
		return NewVariable(!childResult.value.(bool), TYPE_BOOL, false)
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

func (n *BinOp) Evaluate(st *SymbolTable) *Variable {
	if n.value == "&&" {
		leftResult := n.children[0].Evaluate(st)
		if leftResult.varType != TYPE_BOOL {
			panic("[Semantic] Operator && requires bool operands")
		}
		if !leftResult.value.(bool) {
			return NewVariable(false, TYPE_BOOL, false)
		}
		rightResult := n.children[1].Evaluate(st)
		if rightResult.varType != TYPE_BOOL {
			panic("[Semantic] Operator && requires bool operands")
		}
		return NewVariable(rightResult.value.(bool), TYPE_BOOL, false)
	}

	if n.value == "||" {
		leftResult := n.children[0].Evaluate(st)
		if leftResult.varType != TYPE_BOOL {
			panic("[Semantic] Operator || requires bool operands")
		}
		if leftResult.value.(bool) {
			return NewVariable(true, TYPE_BOOL, false)
		}
		rightResult := n.children[1].Evaluate(st)
		if rightResult.varType != TYPE_BOOL {
			panic("[Semantic] Operator || requires bool operands")
		}
		return NewVariable(rightResult.value.(bool), TYPE_BOOL, false)
	}

	leftResult := n.children[0].Evaluate(st)
	rightResult := n.children[1].Evaluate(st)

	switch n.value {
	case "+":
		if leftResult.varType == TYPE_I32 && rightResult.varType == TYPE_I32 {
			return NewVariable(leftResult.value.(int)+rightResult.value.(int), TYPE_I32, false)
		}
		if leftResult.varType == TYPE_STR && rightResult.varType == TYPE_STR {
			return NewVariable(leftResult.value.(string)+rightResult.value.(string), TYPE_STR, false)
		}
		panic("[Semantic] Type mismatch in '+'")
	case "-":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '-' requires i32 operands")
		}
		return NewVariable(leftResult.value.(int)-rightResult.value.(int), TYPE_I32, false)
	case "^":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '^' requires i32 operands")
		}
		return NewVariable(leftResult.value.(int)^rightResult.value.(int), TYPE_I32, false)
	case "*":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '*' requires i32 operands")
		}
		return NewVariable(leftResult.value.(int)*rightResult.value.(int), TYPE_I32, false)
	case "/":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '/' requires i32 operands")
		}
		if rightResult.value.(int) == 0 {
			panic("[Semantic] Division by zero")
		}
		return NewVariable(leftResult.value.(int)/rightResult.value.(int), TYPE_I32, false)
	case "**":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '**' requires i32 operands")
		}
		return NewVariable(int(math.Pow(float64(leftResult.value.(int)), float64(rightResult.value.(int)))), TYPE_I32, false)
	case "==":
		if leftResult.varType != rightResult.varType {
			panic("[Semantic] Type mismatch in '=='")
		}
		switch leftResult.varType {
		case TYPE_I32:
			return NewVariable(leftResult.value.(int) == rightResult.value.(int), TYPE_BOOL, false)
		case TYPE_BOOL:
			return NewVariable(leftResult.value.(bool) == rightResult.value.(bool), TYPE_BOOL, false)
		case TYPE_STR:
			return NewVariable(leftResult.value.(string) == rightResult.value.(string), TYPE_BOOL, false)
		}
		panic("[Semantic] Unsupported type in '=='")
	case ">":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '>' requires i32 operands")
		}
		return NewVariable(leftResult.value.(int) > rightResult.value.(int), TYPE_BOOL, false)
	case "<":
		if leftResult.varType != TYPE_I32 || rightResult.varType != TYPE_I32 {
			panic("[Semantic] Operator '<' requires i32 operands")
		}
		return NewVariable(leftResult.value.(int) < rightResult.value.(int), TYPE_BOOL, false)
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

func (n *Identifier) Evaluate(st *SymbolTable) *Variable {
	return st.Get(n.name).Clone()
}

// Print representa a impressão de um valor
type Print struct {
	children []Node
}

func NewPrint(expr Node) *Print {
	return &Print{children: []Node{expr}}
}

func (n *Print) Evaluate(st *SymbolTable) *Variable {
	result := n.children[0].Evaluate(st)
	if result.varType == TYPE_VOID {
		fmt.Println("<void>")
	} else {
		fmt.Println(result.value)
	}
	return NewVoidVariable()
}

// Assignment representa uma atribuição de variável
type Assignment struct {
	children []Node
}

func NewAssignment(identifier Node, expr Node) *Assignment {
	return &Assignment{children: []Node{identifier, expr}}
}

func (n *Assignment) Evaluate(st *SymbolTable) *Variable {
	identNode := n.children[0].(*Identifier)
	value := n.children[1].Evaluate(st)

	if !st.Exists(identNode.name) {
		// Compatibilidade com programas antigos: primeira atribuição declara implicitamente.
		st.CreateVariable(identNode.name, value.varType, true)
		st.Initialize(identNode.name, value)
		return NewVoidVariable()
	}

	st.Set(identNode.name, value)
	return NewVoidVariable()
}

// VarDec representa declaração de variável com tipo
type VarDec struct {
	value    string
	mutable  bool
	children []Node
}

func NewVarDec(varType string, mutable bool, identifier Node, expr Node) *VarDec {
	children := []Node{identifier}
	if expr != nil {
		children = append(children, expr)
	}
	return &VarDec{value: varType, mutable: mutable, children: children}
}

func (n *VarDec) Evaluate(st *SymbolTable) *Variable {
	identNode := n.children[0].(*Identifier)
	st.CreateVariable(identNode.name, n.value, n.mutable)

	if len(n.children) == 2 {
		initialValue := n.children[1].Evaluate(st)
		st.Initialize(identNode.name, initialValue)
	}

	return NewVoidVariable()
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

func (n *IfNode) Evaluate(st *SymbolTable) *Variable {
	condition := n.children[0].Evaluate(st)
	if condition.varType != TYPE_BOOL {
		panic("[Semantic] if condition must be bool")
	}
	if condition.value.(bool) {
		n.children[1].Evaluate(st)
	} else if len(n.children) == 3 {
		n.children[2].Evaluate(st)
	}
	return NewVoidVariable()
}

// WhileNode representa uma estrutura de repetição while
type WhileNode struct {
	children []Node
}

func NewWhileNode(condition Node, body Node) *WhileNode {
	return &WhileNode{children: []Node{condition, body}}
}

func (n *WhileNode) Evaluate(st *SymbolTable) *Variable {
	for {
		condition := n.children[0].Evaluate(st)
		if condition.varType != TYPE_BOOL {
			panic("[Semantic] while condition must be bool")
		}
		if !condition.value.(bool) {
			break
		}
		n.children[1].Evaluate(st)
	}
	return NewVoidVariable()
}

// ReadNode representa a leitura de inteiro do terminal
type ReadNode struct {
	children []Node
}

func NewReadNode() *ReadNode {
	return &ReadNode{children: []Node{}}
}

func (n *ReadNode) Evaluate(st *SymbolTable) *Variable {
	var value int
	_, err := fmt.Fscan(os.Stdin, &value)
	if err != nil {
		panic("[Semantic] Failed to read integer input")
	}
	return NewVariable(value, TYPE_I32, false)
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

func (n *Block) Evaluate(st *SymbolTable) *Variable {
	for _, child := range n.children {
		child.Evaluate(st)
	}
	return NewVoidVariable()
}

// NoOp representa uma operação vazia (dummy)
type NoOp struct {
	children []Node
}

func NewNoOp() *NoOp {
	return &NoOp{children: []Node{}}
}

func (n *NoOp) Evaluate(st *SymbolTable) *Variable {
	return NewVoidVariable()
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

	// Declaração: let [mut] IDENTIFIER : TYPE [= BOOLEXPRESSION] ;
	if parserLexer.next.tokenType == LET {
		parserLexer.SelectNext()

		isMutable := false
		if parserLexer.next.tokenType == MUT {
			isMutable = true
			parserLexer.SelectNext()
		}

		if parserLexer.next.tokenType != IDEN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected IDEN")
		}
		name := parserLexer.next.value.(string)
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != COLON {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected COLON")
		}
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != TYPE {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected TYPE")
		}
		declaredType := parserLexer.next.value.(string)
		parserLexer.SelectNext()

		var expr Node = nil
		if parserLexer.next.tokenType == ASSIGN {
			parserLexer.SelectNext()
			expr = ParseBoolExpression()
		}

		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
		}
		parserLexer.SelectNext()

		return NewVarDec(declaredType, isMutable, NewIdentifier(name), expr)
	}

	// IF: if ( BOOLEXPRESSION ) STATEMENT [else STATEMENT]
	if parserLexer.next.tokenType == IF {
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

		thenBranch := ParseStatement()

		if parserLexer.next.tokenType == ELSE {
			parserLexer.SelectNext()
			elseBranch := ParseStatement()
			return NewIfNode(condition, thenBranch, elseBranch)
		}

		return NewIfNode(condition, thenBranch, nil)
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

	// Atribuição: IDENTIFIER = BOOLEXPRESSION ;
	if parserLexer.next.tokenType == IDEN {
		name := parserLexer.next.value.(string)
		parserLexer.SelectNext()

		if parserLexer.next.tokenType != ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected ASSIGN")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()

		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;)")
		}
		parserLexer.SelectNext()

		return NewAssignment(NewIdentifier(name), expr)
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

	// Booleano
	if parserLexer.next.tokenType == BOOL {
		value := parserLexer.next.value.(bool)
		parserLexer.SelectNext()
		return NewBoolVal(value)
	}

	// String
	if parserLexer.next.tokenType == STR {
		value := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		return NewStringVal(value)
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
