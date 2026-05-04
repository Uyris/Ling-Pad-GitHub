package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
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
	COMMA     = "COMMA"
	ARROW     = "ARROW"
	END       = "END"
	FUNC      = "FUNC"
	RETURN    = "RETURN"
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

// ==================== Code ====================

// Code armazena e gerencia o código assembly gerado
type Code struct{}

var codeInstructions []string

// Append adiciona uma instrução assembly ao arquivo de saída
func (c *Code) Append(instruction string) {
	codeInstructions = append(codeInstructions, instruction)
}

// Dump escreve o código assembly em um arquivo
func (c *Code) Dump(filename string) {
	header := `section .data
  format_out: db "%d", 10, 0 ; format do printf
  format_in: db "%d", 0 ; format do scanf
  scan_int: dd 0; 32-bits integer

section .text
  extern printf ; usar printf
  extern scanf ; usar scanf
  extern exit ; exit function
  global _start ; início do programa

_start:
  push ebp ; guarda o EBP
  mov ebp, esp ; zera a pilha

  ; aqui começa o codigo gerado:`

	footer := `
  ; aqui termina o código gerado

  mov esp, ebp ; reestabelece a pilha
  pop ebp

  ; chamada de exit(0)
  push 0
  call exit`

	file, err := os.Create(filename)
	if err != nil {
		panic("[Code] Erro ao criar arquivo: " + filename)
	}
	defer file.Close()

	file.WriteString(header + "\n")
	for _, instr := range codeInstructions {
		file.WriteString(instr + "\n")
	}
	file.WriteString(footer + "\n")
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

// Variable representa uma variável com seu valor e deslocamento na pilha
type Variable struct {
	value   interface{}
	varType string
	mutable bool
	isFunc  bool
	shift   int // Deslocamento relativo ao EBP na pilha (em bytes)
}

// NewVariable cria uma nova variável
func NewVariable(value interface{}, varType string, mutable bool) *Variable {
	return &Variable{value: value, varType: varType, mutable: mutable, isFunc: false, shift: 0}
}

func (v *Variable) Clone() *Variable {
	return &Variable{value: v.value, varType: v.varType, mutable: v.mutable, isFunc: v.isFunc, shift: v.shift}
}

func NewVoidVariable() *Variable {
	return NewVariable(nil, TYPE_VOID, false)
}

func NewFunctionVariable(funcNode *FuncDec) *Variable {
	return &Variable{value: funcNode, varType: funcNode.value, mutable: false, isFunc: true, shift: 0}
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

func variableToString(v *Variable) string {
	switch v.varType {
	case TYPE_STR:
		return v.value.(string)
	case TYPE_I32:
		return strconv.Itoa(v.value.(int))
	case TYPE_BOOL:
		if v.value.(bool) {
			return "true"
		}
		return "false"
	default:
		panic("[Semantic] Cannot convert type to string: " + v.varType)
	}
}

// ==================== SymbolTable ====================

// SymbolTable armazena variáveis e seus valores
type SymbolTable struct {
	table         map[string]*Variable
	parent        *SymbolTable
	nextShift     int // Rastreia o próximo deslocamento na pilha
	variableCount int // Número de variáveis declaradas
}

// NewSymbolTable cria uma nova tabela de símbolos
func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		table:     make(map[string]*Variable),
		parent:    parent,
		nextShift: 4, // Primeira variável em [EBP-4]
	}
}

// Get retorna o valor de uma variável
func (st *SymbolTable) Get(name string) *Variable {
	if variable, exists := st.table[name]; exists {
		return variable
	}
	if st.parent != nil {
		return st.parent.Get(name)
	}
	panic("[Semantic] Variável não definida: " + name)
}

func (st *SymbolTable) Exists(name string) bool {
	if _, exists := st.table[name]; exists {
		return true
	}
	if st.parent != nil {
		return st.parent.Exists(name)
	}
	return false
}

func (st *SymbolTable) ExistsCurrent(name string) bool {
	_, exists := st.table[name]
	return exists
}

func (st *SymbolTable) Root() *SymbolTable {
	current := st
	for current.parent != nil {
		current = current.parent
	}
	return current
}

// CreateVariable declara uma variável na tabela
func (st *SymbolTable) CreateVariable(name string, varType string, mutable bool) {
	if st.ExistsCurrent(name) {
		panic("[Semantic] Variável já declarada: " + name)
	}
	variable := NewVariable(defaultValueForType(varType), varType, mutable)
	variable.shift = st.nextShift // Atribui o shift atual
	st.table[name] = variable
	st.nextShift += 4 // Próxima variável será 4 bytes adiante
	st.variableCount++
}

func (st *SymbolTable) CreateFunction(name string, funcNode *FuncDec) {
	if st.ExistsCurrent(name) {
		panic("[Semantic] Função já declarada: " + name)
	}
	st.table[name] = NewFunctionVariable(funcNode)
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
	variable, exists := st.table[name]
	if !exists {
		if st.parent != nil {
			st.parent.Set(name, value)
			return
		}
		panic("[Semantic] Variável não declarada: " + name)
	}
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
	} else if currentChar == '-' && l.position+1 < len(l.source) && l.source[l.position+1] == '>' {
		l.next = Token{tokenType: ARROW, value: "->"}
		l.position += 2
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
	} else if currentChar == ',' {
		l.next = Token{tokenType: COMMA, value: ","}
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
		} else if identStr == "fn" {
			l.next = Token{tokenType: FUNC, value: "fn"}
		} else if identStr == "while" {
			l.next = Token{tokenType: WHILE, value: "while"}
		} else if identStr == "return" {
			l.next = Token{tokenType: RETURN, value: "return"}
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

// ==================== Code Generator ====================

// Variáveis de pacote para geração de código
var codeGenerator = &Code{}
var nextNodeID int = 0

// GetNextNodeID retorna e incrementa o ID único para nós
func GetNextNodeID() int {
	nextNodeID++
	return nextNodeID
}

// ==================== AST Nodes ====================

// Node é a interface base para todos os nós da AST
type Node interface {
	Evaluate(st *SymbolTable) *Variable
	Generate(st *SymbolTable) // Novo método para geração de código
}

// IntVal representa um valor inteiro (nó terminal, sem filhos)
type IntVal struct {
	value    int
	children []Node
	id       int
}

func NewIntVal(value int) *IntVal {
	return &IntVal{value: value, children: []Node{}, id: GetNextNodeID()}
}

func (n *IntVal) Evaluate(st *SymbolTable) *Variable {
	return NewVariable(n.value, TYPE_I32, false)
}

func (n *IntVal) Generate(st *SymbolTable) {
	codeGenerator.Append(fmt.Sprintf("  mov eax, %d ; IntVal %d", n.value, n.id))
}

// BoolVal representa um valor booleano
type BoolVal struct {
	value    bool
	children []Node
	id       int
}

func NewBoolVal(value bool) *BoolVal {
	return &BoolVal{value: value, children: []Node{}, id: GetNextNodeID()}
}

func (n *BoolVal) Evaluate(st *SymbolTable) *Variable {
	return NewVariable(n.value, TYPE_BOOL, false)
}

func (n *BoolVal) Generate(st *SymbolTable) {
	val := 0
	if n.value {
		val = 1
	}
	codeGenerator.Append(fmt.Sprintf("  mov eax, %d ; BoolVal %d", val, n.id))
}

// StringVal representa um valor string
type StringVal struct {
	value    string
	children []Node
	id       int
}

func NewStringVal(value string) *StringVal {
	return &StringVal{value: value, children: []Node{}, id: GetNextNodeID()}
}

func (n *StringVal) Evaluate(st *SymbolTable) *Variable {
	return NewVariable(n.value, TYPE_STR, false)
}

func (n *StringVal) Generate(st *SymbolTable) {
	// Strings não são geradas em assembly
	// Nó dummy para compatibilidade
}

// UnOp representa uma operação unária (1 filho)
type UnOp struct {
	value    string
	children []Node
	id       int
}

func NewUnOp(operator string, operand Node) *UnOp {
	return &UnOp{value: operator, children: []Node{operand}, id: GetNextNodeID()}
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

func (n *UnOp) Generate(st *SymbolTable) {
	// Gera código para o operando (resultado em EAX)
	n.children[0].Generate(st)

	// Aplica a operação unária
	switch n.value {
	case "-":
		codeGenerator.Append("  neg eax ; UnOp negation")
	case "+":
		// Nada a fazer, mantém EAX como está
	case "!":
		codeGenerator.Append("  cmp eax, 0")
		codeGenerator.Append("  mov eax, 0")
		codeGenerator.Append("  mov ecx, 1")
		codeGenerator.Append("  cmove eax, ecx ; UnOp not")
	}
}

// BinOp representa uma operação binária (2 filhos)
type BinOp struct {
	value    string
	children []Node
	id       int
}

func NewBinOp(operator string, left Node, right Node) *BinOp {
	return &BinOp{value: operator, children: []Node{left, right}, id: GetNextNodeID()}
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
		if leftResult.varType == TYPE_STR || rightResult.varType == TYPE_STR {
			return NewVariable(variableToString(leftResult)+variableToString(rightResult), TYPE_STR, false)
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
		if leftResult.varType == TYPE_I32 && rightResult.varType == TYPE_I32 {
			return NewVariable(leftResult.value.(int) > rightResult.value.(int), TYPE_BOOL, false)
		}
		if leftResult.varType == TYPE_STR && rightResult.varType == TYPE_STR {
			return NewVariable(leftResult.value.(string) > rightResult.value.(string), TYPE_BOOL, false)
		}
		panic("[Semantic] Operator '>' requires i32 or str operands of same type")
	case "<":
		if leftResult.varType == TYPE_I32 && rightResult.varType == TYPE_I32 {
			return NewVariable(leftResult.value.(int) < rightResult.value.(int), TYPE_BOOL, false)
		}
		if leftResult.varType == TYPE_STR && rightResult.varType == TYPE_STR {
			return NewVariable(leftResult.value.(string) < rightResult.value.(string), TYPE_BOOL, false)
		}
		panic("[Semantic] Operator '<' requires i32 or str operands of same type")
	}
	panic("[Semantic] Unknown binary operator: " + n.value)
}

func (n *BinOp) Generate(st *SymbolTable) {
	// Gera código para o operando esquerdo
	n.children[0].Generate(st)
	// Empilha o resultado
	codeGenerator.Append("  push eax ; BinOp left operand")

	// Gera código para o operando direito
	n.children[1].Generate(st)
	// Desempilha e realiza a operação
	codeGenerator.Append("  pop ecx ; BinOp right operand")

	switch n.value {
	case "+":
		codeGenerator.Append("  add eax, ecx")
	case "-":
		codeGenerator.Append("  mov edx, ecx")
		codeGenerator.Append("  mov ecx, eax")
		codeGenerator.Append("  mov eax, edx")
		codeGenerator.Append("  sub eax, ecx")
	case "*":
		codeGenerator.Append("  imul ecx")
	case "/":
		codeGenerator.Append("  mov edx, eax")
		codeGenerator.Append("  mov eax, ecx")
		codeGenerator.Append("  mov ecx, edx")
		codeGenerator.Append("  cdq")
		codeGenerator.Append("  idiv ecx")
	case "^":
		codeGenerator.Append("  xor eax, ecx")
	case "==":
		codeGenerator.Append("  cmp eax, ecx")
		codeGenerator.Append("  mov eax, 0")
		codeGenerator.Append("  mov ecx, 1")
		codeGenerator.Append("  cmove eax, ecx")
	case ">":
		codeGenerator.Append("  mov edx, eax")
		codeGenerator.Append("  mov eax, ecx")
		codeGenerator.Append("  cmp eax, edx")
		codeGenerator.Append("  mov eax, 0")
		codeGenerator.Append("  mov ecx, 1")
		codeGenerator.Append("  cmovg eax, ecx")
	case "<":
		codeGenerator.Append("  mov edx, eax")
		codeGenerator.Append("  mov eax, ecx")
		codeGenerator.Append("  cmp eax, edx")
		codeGenerator.Append("  mov eax, 0")
		codeGenerator.Append("  mov ecx, 1")
		codeGenerator.Append("  cmovl eax, ecx")
	case "&&":
		codeGenerator.Append("  and eax, ecx")
	case "||":
		codeGenerator.Append("  or eax, ecx")
	case "**":
		// Exponenciação requer função auxiliar
		codeGenerator.Append("  mov ebx, ecx ; base em EBX")
		codeGenerator.Append("  mov ecx, eax ; expoente em ECX")
		codeGenerator.Append("  mov eax, 1 ; resultado = 1")
		codeGenerator.Append("  cmp ecx, 0")
		codeGenerator.Append("  je pow_end")
		codeGenerator.Append("pow_loop:")
		codeGenerator.Append("  imul eax, ebx")
		codeGenerator.Append("  dec ecx")
		codeGenerator.Append("  jne pow_loop")
		codeGenerator.Append("pow_end:")
	}
}

// Identifier representa uma variável
type Identifier struct {
	name     string
	children []Node
	id       int
}

func NewIdentifier(name string) *Identifier {
	return &Identifier{name: name, children: []Node{}, id: GetNextNodeID()}
}

func (n *Identifier) Evaluate(st *SymbolTable) *Variable {
	variable := st.Get(n.name)
	if variable.isFunc {
		panic("[Semantic] Identificador nao e variavel: " + n.name)
	}
	return variable.Clone()
}

func (n *Identifier) Generate(st *SymbolTable) {
	variable := st.Get(n.name)
	codeGenerator.Append(fmt.Sprintf("  mov eax, [ebp-%d] ; Identifier %s", variable.shift, n.name))
}

// Print representa a impressão de um valor
type Print struct {
	children []Node
	id       int
}

func NewPrint(expr Node) *Print {
	return &Print{children: []Node{expr}, id: GetNextNodeID()}
}

func (n *Print) Evaluate(st *SymbolTable) *Variable {
	result := n.children[0].Evaluate(st)
	if result.varType == TYPE_VOID {
		fmt.Println("<void>")
	} else if result.varType == TYPE_BOOL {
		// Converte booleano para 0/1
		if result.value.(bool) {
			fmt.Println(1)
		} else {
			fmt.Println(0)
		}
	} else {
		fmt.Println(result.value)
	}
	return NewVoidVariable()
}

func (n *Print) Generate(st *SymbolTable) {
	// Gera código para calcular a expressão
	n.children[0].Generate(st)
	// EAX contém o valor a imprimir
	codeGenerator.Append("  push eax ; Push valor a imprimir")
	codeGenerator.Append("  push format_out ; Push formato")
	codeGenerator.Append("  call printf")
	codeGenerator.Append("  add esp, 8 ; Remove argumentos")
}

// Assignment representa uma atribuição de variável
type Assignment struct {
	children []Node
	id       int
}

func NewAssignment(identifier Node, expr Node) *Assignment {
	return &Assignment{children: []Node{identifier, expr}, id: GetNextNodeID()}
}

func (n *Assignment) Evaluate(st *SymbolTable) *Variable {
	identNode := n.children[0].(*Identifier)
	value := n.children[1].Evaluate(st)

	if st.Exists(identNode.name) {
		st.Set(identNode.name, value)
		return NewVoidVariable()
	}
	if st.parent == nil {
		// Compatibilidade com programas antigos: primeira atribuição declara implicitamente.
		st.CreateVariable(identNode.name, value.varType, true)
		st.Initialize(identNode.name, value)
		return NewVoidVariable()
	}
	panic("[Semantic] Variável não declarada: " + identNode.name)
	return NewVoidVariable()
}

func (n *Assignment) Generate(st *SymbolTable) {
	identNode := n.children[0].(*Identifier)
	// Gera código para a expressão
	n.children[1].Generate(st)
	// EAX contém o resultado
	variable := st.Get(identNode.name)
	codeGenerator.Append(fmt.Sprintf("  mov [ebp-%d], eax ; Assignment %s", variable.shift, identNode.name))
}

// VarDec representa declaração de variável com tipo
type VarDec struct {
	value    string
	mutable  bool
	children []Node
	id       int
}

func NewVarDec(varType string, mutable bool, identifier Node, expr Node) *VarDec {
	children := []Node{identifier}
	if expr != nil {
		children = append(children, expr)
	}
	return &VarDec{value: varType, mutable: mutable, children: children, id: GetNextNodeID()}
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

func (n *VarDec) Generate(st *SymbolTable) {
	identNode := n.children[0].(*Identifier)

	// Primeiro, cria a variável na tabela de símbolos
	if !st.ExistsCurrent(identNode.name) {
		st.CreateVariable(identNode.name, n.value, n.mutable)
	}

	variable := st.Get(identNode.name)
	codeGenerator.Append(fmt.Sprintf("  sub esp, 4 ; var %s int [EBP-%d]", identNode.name, variable.shift))

	if len(n.children) == 2 {
		// Gera código para a inicialização
		n.children[1].Generate(st)
		// EAX contém o valor
		codeGenerator.Append(fmt.Sprintf("  mov [ebp-%d], eax ; Initialize %s", variable.shift, identNode.name))
	}
}

// FuncDec representa a declaracao de funcao
type FuncDec struct {
	value    string
	children []Node
	id       int
}

func NewFuncDec(returnType string, identifier Node, params []Node, body Node) *FuncDec {
	children := []Node{identifier}
	children = append(children, params...)
	children = append(children, body)
	return &FuncDec{value: returnType, children: children, id: GetNextNodeID()}
}

func (n *FuncDec) Evaluate(st *SymbolTable) *Variable {
	identNode := n.children[0].(*Identifier)
	st.Root().CreateFunction(identNode.name, n)
	return NewVoidVariable()
}

func (n *FuncDec) Generate(st *SymbolTable) {
	// Funcoes nao sao geradas em assembly neste estagio
}

// FuncCall representa a chamada de uma funcao
type FuncCall struct {
	name     string
	children []Node
	id       int
}

func NewFuncCall(name string, args []Node) *FuncCall {
	return &FuncCall{name: name, children: args, id: GetNextNodeID()}
}

func (n *FuncCall) Evaluate(st *SymbolTable) *Variable {
	variable := st.Get(n.name)
	if !variable.isFunc {
		panic("[Semantic] Identificador nao e funcao: " + n.name)
	}

	funcNode, ok := variable.value.(*FuncDec)
	if !ok {
		panic("[Semantic] Referencia invalida para funcao: " + n.name)
	}

	paramCount := len(funcNode.children) - 2
	if len(n.children) != paramCount {
		panic("[Semantic] Numero incorreto de argumentos em chamada de " + n.name)
	}

	callScope := NewSymbolTable(st.Root())
	for i := 0; i < paramCount; i++ {
		paramNode := funcNode.children[i+1].(*VarDec)
		paramIdent := paramNode.children[0].(*Identifier)

		argValue := n.children[i].Evaluate(st)
		if argValue.varType != paramNode.value {
			panic("[Semantic] Tipo de argumento invalido em chamada de " + n.name)
		}

		callScope.CreateVariable(paramIdent.name, paramNode.value, paramNode.mutable)
		callScope.Initialize(paramIdent.name, argValue)
	}

	body := funcNode.children[len(funcNode.children)-1]
	result := body.Evaluate(callScope)

	if result.varType == TYPE_VOID {
		if funcNode.value != TYPE_VOID {
			panic("[Semantic] Funcao sem retorno: " + n.name)
		}
		return NewVoidVariable()
	}

	if funcNode.value == TYPE_VOID {
		panic("[Semantic] Retorno inesperado em funcao void: " + n.name)
	}
	if result.varType != funcNode.value {
		panic("[Semantic] Tipo de retorno invalido em funcao: " + n.name)
	}

	return result
}

func (n *FuncCall) Generate(st *SymbolTable) {
	// Chamadas de funcao nao sao geradas em assembly neste estagio
}

// Return representa o retorno de funcao
type Return struct {
	children []Node
	id       int
}

func NewReturn(expr Node) *Return {
	return &Return{children: []Node{expr}, id: GetNextNodeID()}
}

func (n *Return) Evaluate(st *SymbolTable) *Variable {
	return n.children[0].Evaluate(st)
}

func (n *Return) Generate(st *SymbolTable) {
	// Retorno nao e gerado em assembly neste estagio
}

// IfNode representa uma estrutura condicional if/else
type IfNode struct {
	children []Node
	id       int
}

func NewIfNode(condition Node, thenBranch Node, elseBranch Node) *IfNode {
	children := []Node{condition, thenBranch}
	if elseBranch != nil {
		children = append(children, elseBranch)
	}
	return &IfNode{children: children, id: GetNextNodeID()}
}

func (n *IfNode) Evaluate(st *SymbolTable) *Variable {
	condition := n.children[0].Evaluate(st)
	if condition.varType != TYPE_BOOL {
		panic("[Semantic] if condition must be bool")
	}
	if condition.value.(bool) {
		result := evaluateWithScope(n.children[1], st)
		if result.varType != TYPE_VOID {
			return result
		}
	} else if len(n.children) == 3 {
		result := evaluateWithScope(n.children[2], st)
		if result.varType != TYPE_VOID {
			return result
		}
	}
	return NewVoidVariable()
}

func (n *IfNode) Generate(st *SymbolTable) {
	else_label := fmt.Sprintf("else_%d", n.id)
	exit_label := fmt.Sprintf("exit_%d", n.id)

	// Gera código para a condição
	n.children[0].Generate(st)
	// EAX contém o resultado da condição
	codeGenerator.Append("  cmp eax, 0 ; if condition")
	codeGenerator.Append(fmt.Sprintf("  je %s", else_label))

	// Bloco then
	n.children[1].Generate(st)
	codeGenerator.Append(fmt.Sprintf("  jmp %s", exit_label))

	// Label else
	codeGenerator.Append(fmt.Sprintf("%s:", else_label))
	if len(n.children) == 3 {
		n.children[2].Generate(st)
	}

	// Label exit
	codeGenerator.Append(fmt.Sprintf("%s:", exit_label))
}

// WhileNode representa uma estrutura de repetição while
type WhileNode struct {
	children []Node
	id       int
}

func NewWhileNode(condition Node, body Node) *WhileNode {
	return &WhileNode{children: []Node{condition, body}, id: GetNextNodeID()}
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
		result := evaluateWithScope(n.children[1], st)
		if result.varType != TYPE_VOID {
			return result
		}
	}
	return NewVoidVariable()
}

func (n *WhileNode) Generate(st *SymbolTable) {
	loop_label := fmt.Sprintf("loop_%d", n.id)
	exit_label := fmt.Sprintf("exit_%d", n.id)

	// Label do loop
	codeGenerator.Append(fmt.Sprintf("%s:", loop_label))

	// Gera código para a condição
	n.children[0].Generate(st)
	// EAX contém o resultado da condição
	codeGenerator.Append("  cmp eax, 0 ; while condition")
	codeGenerator.Append(fmt.Sprintf("  je %s", exit_label))

	// Corpo do loop
	n.children[1].Generate(st)

	// Pula de volta para o teste
	codeGenerator.Append(fmt.Sprintf("  jmp %s", loop_label))

	// Label de saída
	codeGenerator.Append(fmt.Sprintf("%s:", exit_label))
}

// ReadNode representa a leitura de inteiro do terminal
type ReadNode struct {
	children []Node
	id       int
}

func NewReadNode() *ReadNode {
	return &ReadNode{children: []Node{}, id: GetNextNodeID()}
}

func (n *ReadNode) Evaluate(st *SymbolTable) *Variable {
	var value int
	_, err := fmt.Fscan(os.Stdin, &value)
	if err != nil {
		panic("[Semantic] Failed to read integer input")
	}
	return NewVariable(value, TYPE_I32, false)
}

func (n *ReadNode) Generate(st *SymbolTable) {
	codeGenerator.Append("  push scan_int ; endereço de memória de suporte")
	codeGenerator.Append("  push format_in ; formato de entrada (int)")
	codeGenerator.Append("  call scanf")
	codeGenerator.Append("  add esp, 8 ; Remove os argumentos da pilha")
	codeGenerator.Append("  mov eax, dword [scan_int] ; retorna o valor lido em EAX")
}

func evaluateWithScope(child Node, st *SymbolTable) *Variable {
	if _, ok := child.(*Block); ok {
		return child.Evaluate(NewSymbolTable(st))
	}
	return child.Evaluate(st)
}

// Block representa um bloco de instruções
type Block struct {
	children []Node
	id       int
}

func NewBlock() *Block {
	return &Block{children: []Node{}, id: GetNextNodeID()}
}

func (b *Block) AddChild(node Node) {
	b.children = append(b.children, node)
}

func (n *Block) Evaluate(st *SymbolTable) *Variable {
	for _, child := range n.children {
		result := evaluateWithScope(child, st)
		if result.varType != TYPE_VOID {
			return result
		}
	}
	return NewVoidVariable()
}

func (n *Block) Generate(st *SymbolTable) {
	for _, child := range n.children {
		child.Generate(st)
	}
}

// NoOp representa uma operação vazia (dummy)
type NoOp struct {
	children []Node
	id       int
}

func NewNoOp() *NoOp {
	return &NoOp{children: []Node{}, id: GetNextNodeID()}
}

func (n *NoOp) Evaluate(st *SymbolTable) *Variable {
	return NewVoidVariable()
}

func (n *NoOp) Generate(st *SymbolTable) {
	// Nada a gerar
}

// ==================== Parser ====================

// Parser realiza a análise sintática consumindo tokens do Lexer
type Parser struct{}

// lexer é o atributo estático (variável de pacote) do Parser
var parserLexer *Lexer

func ParseProgram() Node {
	block := NewBlock()

	for parserLexer.next.tokenType != EOF {
		if parserLexer.next.tokenType == FUNC {
			block.AddChild(ParseFuncDeclaration())
			continue
		}
		if parserLexer.next.tokenType == LET {
			block.AddChild(ParseVarDeclaration())
			continue
		}
		panic("[Parser] Unexpected token in program: " + parserLexer.next.tokenType)
	}

	return block
}

func ParseVarDeclaration() Node {
	// Declaração: let [mut] IDENTIFIER : TYPE [= BOOLEXPRESSION] ;
	if parserLexer.next.tokenType != LET {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected LET")
	}
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

func ParseFuncDeclaration() Node {
	// Declaracao: fn IDENTIFIER ( [IDENTIFIER : TYPE {, IDENTIFIER : TYPE}] ) [-> TYPE | -> ()] BLOCK
	if parserLexer.next.tokenType != FUNC {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected FUNC")
	}
	parserLexer.SelectNext()

	if parserLexer.next.tokenType != IDEN {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected IDEN")
	}
	funcName := parserLexer.next.value.(string)
	parserLexer.SelectNext()

	if parserLexer.next.tokenType != OPEN_PAR {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
	}
	parserLexer.SelectNext()

	params := []Node{}
	if parserLexer.next.tokenType != CLOSE_PAR {
		for {
			if parserLexer.next.tokenType != IDEN {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected IDEN")
			}
			paramName := parserLexer.next.value.(string)
			parserLexer.SelectNext()

			if parserLexer.next.tokenType != COLON {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected COLON")
			}
			parserLexer.SelectNext()

			if parserLexer.next.tokenType != TYPE {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected TYPE")
			}
			paramType := parserLexer.next.value.(string)
			parserLexer.SelectNext()

			params = append(params, NewVarDec(paramType, false, NewIdentifier(paramName), nil))

			if parserLexer.next.tokenType != COMMA {
				break
			}
			parserLexer.SelectNext()
		}
	}

	if parserLexer.next.tokenType != CLOSE_PAR {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
	}
	parserLexer.SelectNext()

	returnType := TYPE_VOID
	if parserLexer.next.tokenType == ARROW {
		parserLexer.SelectNext()
		if parserLexer.next.tokenType == TYPE {
			returnType = parserLexer.next.value.(string)
			parserLexer.SelectNext()
		} else if parserLexer.next.tokenType == OPEN_PAR {
			parserLexer.SelectNext()
			if parserLexer.next.tokenType != CLOSE_PAR {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
			}
			parserLexer.SelectNext()
			returnType = TYPE_VOID
		} else {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected TYPE or ()")
		}
	}

	body := ParseBlock()
	return NewFuncDec(returnType, NewIdentifier(funcName), params, body)
}

func ParseFuncCallWithName(name string) Node {
	if parserLexer.next.tokenType != OPEN_PAR {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_PAR")
	}
	parserLexer.SelectNext()

	args := []Node{}
	if parserLexer.next.tokenType != CLOSE_PAR {
		args = append(args, ParseBoolExpression())
		for parserLexer.next.tokenType == COMMA {
			parserLexer.SelectNext()
			args = append(args, ParseBoolExpression())
		}
	}

	if parserLexer.next.tokenType != CLOSE_PAR {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected CLOSE_PAR")
	}
	parserLexer.SelectNext()

	return NewFuncCall(name, args)
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
		return ParseVarDeclaration()
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

		if parserLexer.next.tokenType == OPEN_PAR {
			call := ParseFuncCallWithName(name)
			if parserLexer.next.tokenType != END {
				panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
			}
			parserLexer.SelectNext()
			return call
		}
		if parserLexer.next.tokenType != ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected ASSIGN or OPEN_PAR")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()

		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
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

	// Return: return BOOLEXPRESSION ;
	if parserLexer.next.tokenType == RETURN {
		parserLexer.SelectNext()
		expr := ParseBoolExpression()
		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
		}
		parserLexer.SelectNext()
		return NewReturn(expr)
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
		if parserLexer.next.tokenType == OPEN_PAR {
			return ParseFuncCallWithName(name)
		}
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
	if programBlock, ok := result.(*Block); ok {
		programBlock.AddChild(NewFuncCall("main", []Node{}))
	}

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

	// EXECUÇÃO: Criar tabela de símbolos e executar
	st := NewSymbolTable(nil)
	ast.Evaluate(st)

	// GERAÇÃO DE CÓDIGO: Reinicializa para gerar assembly
	codeInstructions = []string{}
	nextNodeID = 0

	// Cria nova tabela de símbolos para geração
	stGen := NewSymbolTable(nil)

	// Gera o código assembly
	ast.Generate(stGen)

	// Gera o nome do arquivo de saída (.asm)
	outputFilename := filename[:len(filename)-len(filepath.Ext(filename))] + ".asm"

	// Escreve o arquivo assembly
	codeGenerator.Dump(outputFilename)
	fmt.Println("[Main] Arquivo gerado: " + outputFilename)
}
