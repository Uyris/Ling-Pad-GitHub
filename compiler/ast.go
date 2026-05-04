package compiler

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

// ==================== Code ====================

// Code armazena e gerencia o codigo assembly gerado
type Code struct{}

var codeInstructions []string

// Append adiciona uma instrucao assembly ao arquivo de saida
func (c *Code) Append(instruction string) {
	codeInstructions = append(codeInstructions, instruction)
}

// Dump escreve o codigo assembly em um arquivo
func (c *Code) Dump(filename string) {
	header := `section .data
  format_out: db "%d", 10, 0 ; format do printf
  format_in: db "%d", 0 ; format do scanf
  scan_int: dd 0; 32-bits integer

section .text
  extern printf ; usar printf
  extern scanf ; usar scanf
  extern exit ; exit function
  global _start ; inicio do programa

_start:
  push ebp ; guarda o EBP
  mov ebp, esp ; zera a pilha

  ; aqui comeca o codigo gerado:`

	footer := `
  ; aqui termina o codigo gerado

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

// ==================== Variable ====================

// Variable representa uma variavel com seu valor e deslocamento na pilha
type Variable struct {
	value    interface{}
	varType  string
	mutable  bool
	isFunc   bool
	isStruct bool
	shift    int // Deslocamento relativo ao EBP na pilha (em bytes)
}

// NewVariable cria uma nova variavel
func NewVariable(value interface{}, varType string, mutable bool) *Variable {
	return &Variable{value: value, varType: varType, mutable: mutable, isFunc: false, isStruct: false, shift: 0}
}

func (v *Variable) Clone() *Variable {
	return &Variable{value: v.value, varType: v.varType, mutable: v.mutable, isFunc: v.isFunc, isStruct: v.isStruct, shift: v.shift}
}

func NewVoidVariable() *Variable {
	return NewVariable(nil, TYPE_VOID, false)
}

func NewFunctionVariable(funcNode *FuncDec) *Variable {
	return &Variable{value: funcNode, varType: funcNode.value, mutable: false, isFunc: true, isStruct: false, shift: 0}
}

func NewStructVariable(structNode *StructDec) *Variable {
	return &Variable{value: structNode, varType: structNode.name, mutable: false, isFunc: false, isStruct: true, shift: 0}
}

func isBuiltinType(varType string) bool {
	return varType == TYPE_I32 || varType == TYPE_BOOL || varType == TYPE_STR
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

// SymbolTable armazena variaveis e seus valores
type SymbolTable struct {
	table         map[string]*Variable
	parent        *SymbolTable
	nextShift     int // Rastreia o proximo deslocamento na pilha
	variableCount int // Numero de variaveis declaradas
}

// NewSymbolTable cria uma nova tabela de simbolos
func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		table:     make(map[string]*Variable),
		parent:    parent,
		nextShift: 4, // Primeira variavel em [EBP-4]
	}
}

// Get retorna o valor de uma variavel
func (st *SymbolTable) Get(name string) *Variable {
	if variable, exists := st.table[name]; exists {
		return variable
	}
	if st.parent != nil {
		return st.parent.Get(name)
	}
	panic("[Semantic] Variavel nao definida: " + name)
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

// CreateVariable declara uma variavel na tabela
func (st *SymbolTable) CreateVariable(name string, varType string, mutable bool) {
	if st.ExistsCurrent(name) {
		panic("[Semantic] Variavel ja declarada: " + name)
	}
	variable := NewVariable(nil, varType, mutable)
	if isBuiltinType(varType) {
		variable.value = defaultValueForType(varType)
	}
	variable.shift = st.nextShift // Atribui o shift atual
	st.table[name] = variable
	st.nextShift += 4 // Proxima variavel sera 4 bytes adiante
	st.variableCount++
}

func (st *SymbolTable) CreateVariableWithValue(name string, value *Variable) {
	if st.ExistsCurrent(name) {
		panic("[Semantic] Variavel ja declarada: " + name)
	}
	value.shift = st.nextShift
	st.table[name] = value
	st.nextShift += 4
	st.variableCount++
}

func (st *SymbolTable) CreateFunction(name string, funcNode *FuncDec) {
	if st.ExistsCurrent(name) {
		panic("[Semantic] Funcao ja declarada: " + name)
	}
	st.table[name] = NewFunctionVariable(funcNode)
}

func (st *SymbolTable) CreateStruct(name string, structNode *StructDec) {
	if st.ExistsCurrent(name) {
		panic("[Semantic] Struct ja declarada: " + name)
	}
	st.table[name] = NewStructVariable(structNode)
}

func (st *SymbolTable) GetStruct(name string) *StructDec {
	root := st.Root()
	if variable, exists := root.table[name]; exists && variable.isStruct {
		structNode, ok := variable.value.(*StructDec)
		if !ok {
			panic("[Semantic] Referencia invalida para struct: " + name)
		}
		return structNode
	}
	panic("[Semantic] Struct nao definida: " + name)
}

func (st *SymbolTable) IsStructType(name string) bool {
	root := st.Root()
	if variable, exists := root.table[name]; exists {
		return variable.isStruct
	}
	return false
}

// Initialize define o valor inicial no momento da declaracao
func (st *SymbolTable) Initialize(name string, value *Variable) {
	variable := st.Get(name)
	if variable.varType != value.varType {
		panic("[Semantic] Type mismatch na inicializacao de " + name)
	}
	variable.value = value.value
}

// Set atualiza uma variavel ja declarada
func (st *SymbolTable) Set(name string, value *Variable) {
	variable, exists := st.table[name]
	if !exists {
		if st.parent != nil {
			st.parent.Set(name, value)
			return
		}
		panic("[Semantic] Variavel nao declarada: " + name)
	}
	if !variable.mutable {
		panic("[Semantic] Variavel imutavel: " + name)
	}

	if variable.varType != value.varType {
		panic("[Semantic] Type mismatch em atribuicao para " + name)
	}

	variable.value = value.value
}

// ==================== Code Generator ====================

// Variaveis de pacote para geracao de codigo
var codeGenerator = &Code{}
var nextNodeID int = 0

// GetNextNodeID retorna e incrementa o ID unico para nos
func GetNextNodeID() int {
	nextNodeID++
	return nextNodeID
}

// ==================== AST Nodes ====================

// Node e a interface base para todos os nos da AST
type Node interface {
	Evaluate(st *SymbolTable) *Variable
	Generate(st *SymbolTable) // Novo metodo para geracao de codigo
}

// IntVal representa um valor inteiro (no terminal, sem filhos)
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
	// Strings nao sao geradas em assembly
	// No dummy para compatibilidade
}

// UnOp representa uma operacao unaria (1 filho)
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
	// Gera codigo para o operando (resultado em EAX)
	n.children[0].Generate(st)

	// Aplica a operacao unaria
	switch n.value {
	case "-":
		codeGenerator.Append("  neg eax ; UnOp negation")
	case "+":
		// Nada a fazer, mantem EAX como esta
	case "!":
		codeGenerator.Append("  cmp eax, 0")
		codeGenerator.Append("  mov eax, 0")
		codeGenerator.Append("  mov ecx, 1")
		codeGenerator.Append("  cmove eax, ecx ; UnOp not")
	}
}

// BinOp representa uma operacao binaria (2 filhos)
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
		panic("[Semantic] Type mismatch in '+'.")
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
	// Gera codigo para o operando esquerdo
	n.children[0].Generate(st)
	// Empilha o resultado
	codeGenerator.Append("  push eax ; BinOp left operand")

	// Gera codigo para o operando direito
	n.children[1].Generate(st)
	// Desempilha e realiza a operacao
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
		// Exponenciacao requer funcao auxiliar
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

// Identifier representa uma variavel
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
	if variable.isFunc || variable.isStruct {
		panic("[Semantic] Identificador nao e variavel: " + n.name)
	}
	return variable.Clone()
}

func (n *Identifier) Generate(st *SymbolTable) {
	variable := st.Get(n.name)
	codeGenerator.Append(fmt.Sprintf("  mov eax, [ebp-%d] ; Identifier %s", variable.shift, n.name))
}

// FieldAccess representa acesso a atributo de struct
type FieldAccess struct {
	base  Node
	field string
	id    int
}

func NewFieldAccess(base Node, field string) *FieldAccess {
	return &FieldAccess{base: base, field: field, id: GetNextNodeID()}
}

func (n *FieldAccess) Evaluate(st *SymbolTable) *Variable {
	fieldVar := resolveFieldTarget(n.base, n.field, st)
	return fieldVar.Clone()
}

func (n *FieldAccess) Generate(st *SymbolTable) {
	// Acesso a struct nao e gerado em assembly neste estagio
}

func resolveFieldTarget(base Node, field string, st *SymbolTable) *Variable {
	baseVar := resolveVariableReference(base, st)
	if baseVar.isStruct {
		panic("[Semantic] Identificador nao e instancia de struct")
	}
	if !st.IsStructType(baseVar.varType) {
		panic("[Semantic] Acesso de atributo em tipo nao struct")
	}
	fields, ok := baseVar.value.(map[string]*Variable)
	if !ok {
		panic("[Semantic] Struct mal formada")
	}
	fieldVar, exists := fields[field]
	if !exists {
		panic("[Semantic] Atributo inexistente: " + field)
	}
	return fieldVar
}

func resolveVariableReference(node Node, st *SymbolTable) *Variable {
	switch target := node.(type) {
	case *Identifier:
		return st.Get(target.name)
	case *FieldAccess:
		return resolveFieldTarget(target.base, target.field, st)
	default:
		panic("[Semantic] Atribuicao invalida")
	}
}

// Print representa a impressao de um valor
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
		if result.value.(bool) {
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}
	} else {
		fmt.Println(result.value)
	}
	return NewVoidVariable()
}

func (n *Print) Generate(st *SymbolTable) {
	// Gera codigo para calcular a expressao
	n.children[0].Generate(st)
	// EAX contem o valor a imprimir
	codeGenerator.Append("  push eax ; Push valor a imprimir")
	codeGenerator.Append("  push format_out ; Push formato")
	codeGenerator.Append("  call printf")
	codeGenerator.Append("  add esp, 8 ; Remove argumentos")
}

// Assignment representa uma atribuicao de variavel
type Assignment struct {
	children []Node
	id       int
}

func NewAssignment(identifier Node, expr Node) *Assignment {
	return &Assignment{children: []Node{identifier, expr}, id: GetNextNodeID()}
}

func (n *Assignment) Evaluate(st *SymbolTable) *Variable {
	value := n.children[1].Evaluate(st)

	switch target := n.children[0].(type) {
	case *Identifier:
		if st.Exists(target.name) {
			st.Set(target.name, value)
			return NewVoidVariable()
		}
		if st.parent == nil {
			// Compatibilidade com programas antigos: primeira atribuicao declara implicitamente.
			st.CreateVariable(target.name, value.varType, true)
			st.Initialize(target.name, value)
			return NewVoidVariable()
		}
		panic("[Semantic] Variavel nao declarada: " + target.name)
	case *FieldAccess:
		fieldVar := resolveFieldTarget(target.base, target.field, st)
		if !fieldVar.mutable {
			panic("[Semantic] Atributo imutavel: " + target.field)
		}
		if fieldVar.varType != value.varType {
			panic("[Semantic] Type mismatch em atribuicao para atributo: " + target.field)
		}
		fieldVar.value = value.value
		return NewVoidVariable()
	default:
		panic("[Semantic] Atribuicao invalida")
	}
	return NewVoidVariable()
}

func (n *Assignment) Generate(st *SymbolTable) {
	switch target := n.children[0].(type) {
	case *Identifier:
		// Gera codigo para a expressao
		n.children[1].Generate(st)
		// EAX contem o resultado
		variable := st.Get(target.name)
		codeGenerator.Append(fmt.Sprintf("  mov [ebp-%d], eax ; Assignment %s", variable.shift, target.name))
	case *FieldAccess:
		// Structs nao sao gerados em assembly neste estagio
		return
	default:
		return
	}
}

// VarDec representa declaracao de variavel com tipo
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
	if isBuiltinType(n.value) {
		st.CreateVariable(identNode.name, n.value, n.mutable)

		if len(n.children) == 2 {
			initialValue := n.children[1].Evaluate(st)
			st.Initialize(identNode.name, initialValue)
		}
		return NewVoidVariable()
	}
	if st.IsStructType(n.value) {
		if len(n.children) == 2 {
			panic("[Semantic] Struct nao permite inicializacao na declaracao: " + identNode.name)
		}
		structNode := st.GetStruct(n.value)
		instance := newStructInstance(structNode, st)
		instance.mutable = n.mutable
		st.CreateVariableWithValue(identNode.name, instance)
		return NewVoidVariable()
	}
	panic("[Semantic] Tipo desconhecido: " + n.value)

	return NewVoidVariable()
}

func (n *VarDec) Generate(st *SymbolTable) {
	identNode := n.children[0].(*Identifier)
	if !isBuiltinType(n.value) {
		if !st.ExistsCurrent(identNode.name) {
			st.CreateVariableWithValue(identNode.name, NewVariable(nil, n.value, n.mutable))
		}
		return
	}

	// Primeiro, cria a variavel na tabela de simbolos
	if !st.ExistsCurrent(identNode.name) {
		st.CreateVariable(identNode.name, n.value, n.mutable)
	}

	variable := st.Get(identNode.name)
	codeGenerator.Append(fmt.Sprintf("  sub esp, 4 ; var %s int [EBP-%d]", identNode.name, variable.shift))

	if len(n.children) == 2 {
		// Gera codigo para a inicializacao
		n.children[1].Generate(st)
		// EAX contem o valor
		codeGenerator.Append(fmt.Sprintf("  mov [ebp-%d], eax ; Initialize %s", variable.shift, identNode.name))
	}
}

// StructDec representa a declaracao de struct
type StructDec struct {
	name   string
	fields []*VarDec
	id     int
}

func NewStructDec(name string, fields []*VarDec) *StructDec {
	return &StructDec{name: name, fields: fields, id: GetNextNodeID()}
}

func (n *StructDec) Evaluate(st *SymbolTable) *Variable {
	st.Root().CreateStruct(n.name, n)
	return NewVoidVariable()
}

func (n *StructDec) Generate(st *SymbolTable) {
	// Structs nao sao geradas em assembly neste estagio
}

func newStructInstance(structNode *StructDec, st *SymbolTable) *Variable {
	fields := map[string]*Variable{}
	for _, field := range structNode.fields {
		fieldIdent := field.children[0].(*Identifier)
		fieldType := field.value
		var fieldValue *Variable
		if isBuiltinType(fieldType) {
			fieldValue = NewVariable(defaultValueForType(fieldType), fieldType, field.mutable)
		} else {
			nestedStruct := st.GetStruct(fieldType)
			fieldValue = newStructInstance(nestedStruct, st)
			fieldValue.mutable = field.mutable
		}
		fields[fieldIdent.name] = fieldValue
	}
	return &Variable{value: fields, varType: structNode.name, mutable: true, isFunc: false, isStruct: false, shift: 0}
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
	elseLabel := fmt.Sprintf("else_%d", n.id)
	exitLabel := fmt.Sprintf("exit_%d", n.id)

	// Gera codigo para a condicao
	n.children[0].Generate(st)
	// EAX contem o resultado da condicao
	codeGenerator.Append("  cmp eax, 0 ; if condition")
	codeGenerator.Append(fmt.Sprintf("  je %s", elseLabel))

	// Bloco then
	n.children[1].Generate(st)
	codeGenerator.Append(fmt.Sprintf("  jmp %s", exitLabel))

	// Label else
	codeGenerator.Append(fmt.Sprintf("%s:", elseLabel))
	if len(n.children) == 3 {
		n.children[2].Generate(st)
	}

	// Label exit
	codeGenerator.Append(fmt.Sprintf("%s:", exitLabel))
}

// WhileNode representa uma estrutura de repeticao while
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
	loopLabel := fmt.Sprintf("loop_%d", n.id)
	exitLabel := fmt.Sprintf("exit_%d", n.id)

	// Label do loop
	codeGenerator.Append(fmt.Sprintf("%s:", loopLabel))

	// Gera codigo para a condicao
	n.children[0].Generate(st)
	// EAX contem o resultado da condicao
	codeGenerator.Append("  cmp eax, 0 ; while condition")
	codeGenerator.Append(fmt.Sprintf("  je %s", exitLabel))

	// Corpo do loop
	n.children[1].Generate(st)

	// Pula de volta para o teste
	codeGenerator.Append(fmt.Sprintf("  jmp %s", loopLabel))

	// Label de saida
	codeGenerator.Append(fmt.Sprintf("%s:", exitLabel))
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
	codeGenerator.Append("  push scan_int ; endereco de memoria de suporte")
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

// Block representa um bloco de instrucoes
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

// NoOp representa uma operacao vazia (dummy)
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
