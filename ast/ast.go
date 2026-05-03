package ast

import (
	"fmt"
	"math"
	"os"

	"lingpad/semantic"
	"lingpad/token"
)

// Variáveis de pacote para geração de código
var CodeGenerator *CodeGen
var NextNodeID int = 0

// GetNextNodeID retorna e incrementa o ID único para nós
func GetNextNodeID() int {
	NextNodeID++
	return NextNodeID
}

// Node é a interface base para todos os nós da AST
type Node interface {
	Evaluate(st *semantic.SymbolTable) *semantic.Variable
	Generate(st *semantic.SymbolTable) // Novo método para geração de código
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

func (n *IntVal) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	return semantic.NewVariable(n.value, token.TYPE_I32, false)
}

func (n *IntVal) Generate(st *semantic.SymbolTable) {
	CodeGenerator.Append(fmt.Sprintf("  mov eax, %d ; IntVal %d", n.value, n.id))
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

func (n *BoolVal) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	return semantic.NewVariable(n.value, token.TYPE_BOOL, false)
}

func (n *BoolVal) Generate(st *semantic.SymbolTable) {
	val := 0
	if n.value {
		val = 1
	}
	CodeGenerator.Append(fmt.Sprintf("  mov eax, %d ; BoolVal %d", val, n.id))
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

func (n *StringVal) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	return semantic.NewVariable(n.value, token.TYPE_STR, false)
}

func (n *StringVal) Generate(st *semantic.SymbolTable) {
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

func (n *UnOp) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	childResult := n.children[0].Evaluate(st)
	if n.value == "-" {
		if childResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Unary '-' requires i32")
		}
		return semantic.NewVariable(-(childResult.Value.(int)), token.TYPE_I32, false)
	} else if n.value == "+" {
		if childResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Unary '+' requires i32")
		}
		return semantic.NewVariable(childResult.Value.(int), token.TYPE_I32, false)
	} else if n.value == "!" {
		if childResult.VarType != token.TYPE_BOOL {
			panic("[Semantic] Unary '!' requires bool")
		}
		return semantic.NewVariable(!childResult.Value.(bool), token.TYPE_BOOL, false)
	}
	panic("[Semantic] Unknown unary operator: " + n.value)
}

func (n *UnOp) Generate(st *semantic.SymbolTable) {
	// Gera código para o operando (resultado em EAX)
	n.children[0].Generate(st)

	// Aplica a operação unária
	switch n.value {
	case "-":
		CodeGenerator.Append("  neg eax ; UnOp negation")
	case "+":
		// Nada a fazer, mantém EAX como está
	case "!":
		CodeGenerator.Append("  cmp eax, 0")
		CodeGenerator.Append("  mov eax, 0")
		CodeGenerator.Append("  mov ecx, 1")
		CodeGenerator.Append("  cmove eax, ecx ; UnOp not")
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

func (n *BinOp) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	if n.value == "&&" {
		leftResult := n.children[0].Evaluate(st)
		if leftResult.VarType != token.TYPE_BOOL {
			panic("[Semantic] Operator && requires bool operands")
		}
		if !leftResult.Value.(bool) {
			return semantic.NewVariable(false, token.TYPE_BOOL, false)
		}
		rightResult := n.children[1].Evaluate(st)
		if rightResult.VarType != token.TYPE_BOOL {
			panic("[Semantic] Operator && requires bool operands")
		}
		return semantic.NewVariable(rightResult.Value.(bool), token.TYPE_BOOL, false)
	}

	if n.value == "||" {
		leftResult := n.children[0].Evaluate(st)
		if leftResult.VarType != token.TYPE_BOOL {
			panic("[Semantic] Operator || requires bool operands")
		}
		if leftResult.Value.(bool) {
			return semantic.NewVariable(true, token.TYPE_BOOL, false)
		}
		rightResult := n.children[1].Evaluate(st)
		if rightResult.VarType != token.TYPE_BOOL {
			panic("[Semantic] Operator || requires bool operands")
		}
		return semantic.NewVariable(rightResult.Value.(bool), token.TYPE_BOOL, false)
	}

	leftResult := n.children[0].Evaluate(st)
	rightResult := n.children[1].Evaluate(st)

	switch n.value {
	case "+":
		if leftResult.VarType == token.TYPE_I32 && rightResult.VarType == token.TYPE_I32 {
			return semantic.NewVariable(leftResult.Value.(int)+rightResult.Value.(int), token.TYPE_I32, false)
		}
		if leftResult.VarType == token.TYPE_STR || rightResult.VarType == token.TYPE_STR {
			return semantic.NewVariable(semantic.VariableToString(leftResult)+semantic.VariableToString(rightResult), token.TYPE_STR, false)
		}
		panic("[Semantic] Type mismatch in '+'")
	case "-":
		if leftResult.VarType != token.TYPE_I32 || rightResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Operator '-' requires i32 operands")
		}
		return semantic.NewVariable(leftResult.Value.(int)-rightResult.Value.(int), token.TYPE_I32, false)
	case "^":
		if leftResult.VarType != token.TYPE_I32 || rightResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Operator '^' requires i32 operands")
		}
		return semantic.NewVariable(leftResult.Value.(int)^rightResult.Value.(int), token.TYPE_I32, false)
	case "*":
		if leftResult.VarType != token.TYPE_I32 || rightResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Operator '*' requires i32 operands")
		}
		return semantic.NewVariable(leftResult.Value.(int)*rightResult.Value.(int), token.TYPE_I32, false)
	case "/":
		if leftResult.VarType != token.TYPE_I32 || rightResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Operator '/' requires i32 operands")
		}
		if rightResult.Value.(int) == 0 {
			panic("[Semantic] Division by zero")
		}
		return semantic.NewVariable(leftResult.Value.(int)/rightResult.Value.(int), token.TYPE_I32, false)
	case "**":
		if leftResult.VarType != token.TYPE_I32 || rightResult.VarType != token.TYPE_I32 {
			panic("[Semantic] Operator '**' requires i32 operands")
		}
		return semantic.NewVariable(int(math.Pow(float64(leftResult.Value.(int)), float64(rightResult.Value.(int)))), token.TYPE_I32, false)
	case "==":
		if leftResult.VarType != rightResult.VarType {
			panic("[Semantic] Type mismatch in '=='")
		}
		switch leftResult.VarType {
		case token.TYPE_I32:
			return semantic.NewVariable(leftResult.Value.(int) == rightResult.Value.(int), token.TYPE_BOOL, false)
		case token.TYPE_BOOL:
			return semantic.NewVariable(leftResult.Value.(bool) == rightResult.Value.(bool), token.TYPE_BOOL, false)
		case token.TYPE_STR:
			return semantic.NewVariable(leftResult.Value.(string) == rightResult.Value.(string), token.TYPE_BOOL, false)
		}
		panic("[Semantic] Unsupported type in '=='")
	case ">":
		if leftResult.VarType == token.TYPE_I32 && rightResult.VarType == token.TYPE_I32 {
			return semantic.NewVariable(leftResult.Value.(int) > rightResult.Value.(int), token.TYPE_BOOL, false)
		}
		if leftResult.VarType == token.TYPE_STR && rightResult.VarType == token.TYPE_STR {
			return semantic.NewVariable(leftResult.Value.(string) > rightResult.Value.(string), token.TYPE_BOOL, false)
		}
		panic("[Semantic] Operator '>' requires i32 or str operands of same type")
	case "<":
		if leftResult.VarType == token.TYPE_I32 && rightResult.VarType == token.TYPE_I32 {
			return semantic.NewVariable(leftResult.Value.(int) < rightResult.Value.(int), token.TYPE_BOOL, false)
		}
		if leftResult.VarType == token.TYPE_STR && rightResult.VarType == token.TYPE_STR {
			return semantic.NewVariable(leftResult.Value.(string) < rightResult.Value.(string), token.TYPE_BOOL, false)
		}
		panic("[Semantic] Operator '<' requires i32 or str operands of same type")
	}
	panic("[Semantic] Unknown binary operator: " + n.value)
}

func (n *BinOp) Generate(st *semantic.SymbolTable) {
	// Gera código para o operando esquerdo
	n.children[0].Generate(st)
	// Empilha o resultado
	CodeGenerator.Append("  push eax ; BinOp left operand")

	// Gera código para o operando direito
	n.children[1].Generate(st)
	// Desempilha e realiza a operação
	CodeGenerator.Append("  pop ecx ; BinOp right operand")

	switch n.value {
	case "+":
		CodeGenerator.Append("  add eax, ecx")
	case "-":
		CodeGenerator.Append("  mov edx, ecx")
		CodeGenerator.Append("  mov ecx, eax")
		CodeGenerator.Append("  mov eax, edx")
		CodeGenerator.Append("  sub eax, ecx")
	case "*":
		CodeGenerator.Append("  imul ecx")
	case "/":
		CodeGenerator.Append("  mov edx, eax")
		CodeGenerator.Append("  mov eax, ecx")
		CodeGenerator.Append("  mov ecx, edx")
		CodeGenerator.Append("  cdq")
		CodeGenerator.Append("  idiv ecx")
	case "^":
		CodeGenerator.Append("  xor eax, ecx")
	case "==":
		CodeGenerator.Append("  cmp eax, ecx")
		CodeGenerator.Append("  mov eax, 0")
		CodeGenerator.Append("  mov ecx, 1")
		CodeGenerator.Append("  cmove eax, ecx")
	case ">":
		CodeGenerator.Append("  mov edx, eax")
		CodeGenerator.Append("  mov eax, ecx")
		CodeGenerator.Append("  cmp eax, edx")
		CodeGenerator.Append("  mov eax, 0")
		CodeGenerator.Append("  mov ecx, 1")
		CodeGenerator.Append("  cmovg eax, ecx")
	case "<":
		CodeGenerator.Append("  mov edx, eax")
		CodeGenerator.Append("  mov eax, ecx")
		CodeGenerator.Append("  cmp eax, edx")
		CodeGenerator.Append("  mov eax, 0")
		CodeGenerator.Append("  mov ecx, 1")
		CodeGenerator.Append("  cmovl eax, ecx")
	case "&&":
		CodeGenerator.Append("  and eax, ecx")
	case "||":
		CodeGenerator.Append("  or eax, ecx")
	case "**":
		// Exponenciação requer função auxiliar
		CodeGenerator.Append("  mov ebx, ecx ; base em EBX")
		CodeGenerator.Append("  mov ecx, eax ; expoente em ECX")
		CodeGenerator.Append("  mov eax, 1 ; resultado = 1")
		CodeGenerator.Append("  cmp ecx, 0")
		CodeGenerator.Append("  je pow_end")
		CodeGenerator.Append("pow_loop:")
		CodeGenerator.Append("  imul eax, ebx")
		CodeGenerator.Append("  dec ecx")
		CodeGenerator.Append("  jne pow_loop")
		CodeGenerator.Append("pow_end:")
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

func (n *Identifier) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	return st.Get(n.name).Clone()
}

func (n *Identifier) Generate(st *semantic.SymbolTable) {
	variable := st.Get(n.name)
	if variable.Shift > 0 {
		// Parâmetro (acesso relativo ao EBP positivo)
		CodeGenerator.Append(fmt.Sprintf("  mov eax, [ebp+%d] ; Identifier %s (param)", variable.Shift, n.name))
	} else {
		// Variável local (acesso relativo ao EBP negativo)
		CodeGenerator.Append(fmt.Sprintf("  mov eax, [ebp-%d] ; Identifier %s", -variable.Shift, n.name))
	}
}

// Print representa a impressão de um valor
type Print struct {
	children []Node
	id       int
}

func NewPrint(expr Node) *Print {
	return &Print{children: []Node{expr}, id: GetNextNodeID()}
}

func (n *Print) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	result := n.children[0].Evaluate(st)
	if result.VarType == token.TYPE_VOID {
		fmt.Println("<void>")
	} else if result.VarType == token.TYPE_BOOL {
		// Converte booleano para 0/1
		if result.Value.(bool) {
			fmt.Println(1)
		} else {
			fmt.Println(0)
		}
	} else {
		fmt.Println(result.Value)
	}
	return semantic.NewVoidVariable()
}

func (n *Print) Generate(st *semantic.SymbolTable) {
	// Gera código para calcular a expressão
	n.children[0].Generate(st)
	// EAX contém o valor a imprimir
	CodeGenerator.Append("  push eax ; Push valor a imprimir")
	CodeGenerator.Append("  push format_out ; Push formato")
	CodeGenerator.Append("  call printf")
	CodeGenerator.Append("  add esp, 8 ; Remove argumentos")
}

// Assignment representa uma atribuição de variável
type Assignment struct {
	children []Node
	id       int
}

func NewAssignment(identifier Node, expr Node) *Assignment {
	return &Assignment{children: []Node{identifier, expr}, id: GetNextNodeID()}
}

func (n *Assignment) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	identNode := n.children[0].(*Identifier)
	value := n.children[1].Evaluate(st)

	if !st.Exists(identNode.name) {
		// Compatibilidade com programas antigos: primeira atribuição declara implicitamente.
		st.CreateVariable(identNode.name, value.VarType, true)
		st.Initialize(identNode.name, value)
		return semantic.NewVoidVariable()
	}

	st.Set(identNode.name, value)
	return semantic.NewVoidVariable()
}

func (n *Assignment) Generate(st *semantic.SymbolTable) {
	identNode := n.children[0].(*Identifier)
	// Gera código para a expressão
	n.children[1].Generate(st)
	// EAX contém o resultado
	variable := st.Get(identNode.name)
	if variable.Shift < 0 {
		CodeGenerator.Append(fmt.Sprintf("  mov [ebp%d], eax ; Assignment %s", variable.Shift, identNode.name))
	} else {
		CodeGenerator.Append(fmt.Sprintf("  mov [ebp+%d], eax ; Assignment %s", variable.Shift, identNode.name))
	}
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

func (n *VarDec) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	identNode := n.children[0].(*Identifier)
	st.CreateVariable(identNode.name, n.value, n.mutable)

	if len(n.children) == 2 {
		initialValue := n.children[1].Evaluate(st)
		st.Initialize(identNode.name, initialValue)
	}

	return semantic.NewVoidVariable()
}

func (n *VarDec) Generate(st *semantic.SymbolTable) {
	identNode := n.children[0].(*Identifier)

	// Primeiro, cria a variável na tabela de símbolos
	if !st.Exists(identNode.name) {
		st.CreateVariable(identNode.name, n.value, n.mutable)
	}

	variable := st.Get(identNode.name)

	// Aloca espaço na pilha
	if variable.Shift < 0 {
		CodeGenerator.Append(fmt.Sprintf("  sub esp, 4 ; var %s int [EBP%d]", identNode.name, variable.Shift))
	} else {
		CodeGenerator.Append(fmt.Sprintf("  sub esp, 4 ; var %s int [EBP+%d]", identNode.name, variable.Shift))
	}

	if len(n.children) == 2 {
		// Gera código para a inicialização
		n.children[1].Generate(st)
		// EAX contém o valor
		if variable.Shift < 0 {
			CodeGenerator.Append(fmt.Sprintf("  mov [ebp%d], eax ; Initialize %s", variable.Shift, identNode.name))
		} else {
			CodeGenerator.Append(fmt.Sprintf("  mov [ebp+%d], eax ; Initialize %s", variable.Shift, identNode.name))
		}
	}
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

func (n *IfNode) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	condition := n.children[0].Evaluate(st)
	if condition.VarType != token.TYPE_BOOL {
		panic("[Semantic] if condition must be bool")
	}
	if condition.Value.(bool) {
		n.children[1].Evaluate(st)
	} else if len(n.children) == 3 {
		n.children[2].Evaluate(st)
	}
	return semantic.NewVoidVariable()
}

func (n *IfNode) Generate(st *semantic.SymbolTable) {
	else_label := fmt.Sprintf("else_%d", n.id)
	exit_label := fmt.Sprintf("exit_%d", n.id)

	// Gera código para a condição
	n.children[0].Generate(st)
	// EAX contém o resultado da condição
	CodeGenerator.Append("  cmp eax, 0 ; if condition")
	CodeGenerator.Append(fmt.Sprintf("  je %s", else_label))

	// Bloco then
	n.children[1].Generate(st)
	CodeGenerator.Append(fmt.Sprintf("  jmp %s", exit_label))

	// Label else
	CodeGenerator.Append(fmt.Sprintf("%s:", else_label))
	if len(n.children) == 3 {
		n.children[2].Generate(st)
	}

	// Label exit
	CodeGenerator.Append(fmt.Sprintf("%s:", exit_label))
}

// WhileNode representa uma estrutura de repetição while
type WhileNode struct {
	children []Node
	id       int
}

func NewWhileNode(condition Node, body Node) *WhileNode {
	return &WhileNode{children: []Node{condition, body}, id: GetNextNodeID()}
}

func (n *WhileNode) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	for {
		condition := n.children[0].Evaluate(st)
		if condition.VarType != token.TYPE_BOOL {
			panic("[Semantic] while condition must be bool")
		}
		if !condition.Value.(bool) {
			break
		}
		n.children[1].Evaluate(st)
	}
	return semantic.NewVoidVariable()
}

func (n *WhileNode) Generate(st *semantic.SymbolTable) {
	loop_label := fmt.Sprintf("loop_%d", n.id)
	exit_label := fmt.Sprintf("exit_%d", n.id)

	// Label do loop
	CodeGenerator.Append(fmt.Sprintf("%s:", loop_label))

	// Gera código para a condição
	n.children[0].Generate(st)
	// EAX contém o resultado da condição
	CodeGenerator.Append("  cmp eax, 0 ; while condition")
	CodeGenerator.Append(fmt.Sprintf("  je %s", exit_label))

	// Corpo do loop
	n.children[1].Generate(st)

	// Pula de volta para o teste
	CodeGenerator.Append(fmt.Sprintf("  jmp %s", loop_label))

	// Label de saída
	CodeGenerator.Append(fmt.Sprintf("%s:", exit_label))
}

// ReadNode representa a leitura de inteiro do terminal
type ReadNode struct {
	children []Node
	id       int
}

func NewReadNode() *ReadNode {
	return &ReadNode{children: []Node{}, id: GetNextNodeID()}
}

func (n *ReadNode) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	var value int
	_, err := fmt.Fscan(os.Stdin, &value)
	if err != nil {
		panic("[Semantic] Failed to read integer input")
	}
	return semantic.NewVariable(value, token.TYPE_I32, false)
}

func (n *ReadNode) Generate(st *semantic.SymbolTable) {
	CodeGenerator.Append("  push scan_int ; endereço de memória de suporte")
	CodeGenerator.Append("  push format_in ; formato de entrada (int)")
	CodeGenerator.Append("  call scanf")
	CodeGenerator.Append("  add esp, 8 ; Remove os argumentos da pilha")
	CodeGenerator.Append("  mov eax, dword [scan_int] ; retorna o valor lido em EAX")
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

func (n *Block) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	for _, child := range n.children {
		child.Evaluate(st)
		if st.HasReturned {
			return st.ReturnValue
		}
	}
	return semantic.NewVoidVariable()
}

func (n *Block) Generate(st *semantic.SymbolTable) {
	for _, child := range n.children {
		child.Generate(st)
	}
}

// ReturnNode representa um comando return
type ReturnNode struct {
	children []Node
	id       int
}

func NewReturnNode(expr Node) *ReturnNode {
	children := []Node{}
	if expr != nil {
		children = append(children, expr)
	}
	return &ReturnNode{children: children, id: GetNextNodeID()}
}

func (n *ReturnNode) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	if len(n.children) == 0 {
		st.ReturnValue = semantic.NewVoidVariable()
	} else {
		st.ReturnValue = n.children[0].Evaluate(st)
	}
	st.HasReturned = true
	return st.ReturnValue
}

func (n *ReturnNode) Generate(st *semantic.SymbolTable) {
	if len(n.children) > 0 {
		n.children[0].Generate(st)
		// EAX contém o valor de retorno
	}
	// Epilogue da função
	CodeGenerator.Append("  mov esp, ebp ; epilogue")
	CodeGenerator.Append("  pop ebp")
	CodeGenerator.Append("  ret")
}

// FunctionDec representa a declaração de uma função
type FunctionDec struct {
	name       string
	params     []string // nomes dos parâmetros
	paramTypes []string // tipos dos parâmetros
	children   []Node   // [0] = body (Block)
	id         int
}

func NewFunctionDec(name string, params []string, paramTypes []string, body Node) *FunctionDec {
	return &FunctionDec{
		name:       name,
		params:     params,
		paramTypes: paramTypes,
		children:   []Node{body},
		id:         GetNextNodeID(),
	}
}

func (n *FunctionDec) GetBody() Node {
	if len(n.children) > 0 {
		return n.children[0]
	}
	return nil
}

func (n *FunctionDec) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	// Registra a função na tabela de símbolos
	sig := &semantic.FunctionSignature{
		Name:       n.name,
		Params:     n.params,
		ParamTypes: n.paramTypes,
		Node:       n,
	}
	st.Functions[n.name] = sig
	return semantic.NewVoidVariable()
}

func (n *FunctionDec) Generate(st *semantic.SymbolTable) {
	// Gera o rótulo da função
	CodeGenerator.Append("")
	CodeGenerator.Append(fmt.Sprintf("%s:", n.name))

	// Prologue da função
	CodeGenerator.Append("  push ebp ; prologue")
	CodeGenerator.Append("  mov ebp, esp")

	// Cria nova tabela de símbolos para escopo local
	localSt := semantic.NewSymbolTable()
	localSt.Functions = st.Functions // Herda as funções definidas

	// Adiciona os parâmetros à tabela de símbolos local
	// Os parâmetros estão em [EBP+8], [EBP+12], etc. (após EBP e endereço de retorno)
	paramShift := 8
	for i, paramName := range n.params {
		localSt.CreateVariable(paramName, n.paramTypes[i], false)
		localSt.Table[paramName].Shift = paramShift // Positivo para parâmetros
		paramShift += 4
	}

	// Gera código para o corpo da função
	n.children[0].Generate(localSt)

	// Se não houver return explícito, adiciona epilogue
	CodeGenerator.Append("  mov eax, 0 ; default return")
	CodeGenerator.Append("  mov esp, ebp ; epilogue (default)")
	CodeGenerator.Append("  pop ebp")
	CodeGenerator.Append("  ret")
}

// FunctionCall representa uma chamada de função
type FunctionCall struct {
	name     string
	children []Node // argumentos
	id       int
}

func NewFunctionCall(name string, args []Node) *FunctionCall {
	return &FunctionCall{
		name:     name,
		children: args,
		id:       GetNextNodeID(),
	}
}

func (n *FunctionCall) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	// Procura a função na tabela de símbolos
	sig, exists := st.Functions[n.name]
	if !exists {
		panic("[Semantic] Função não definida: " + n.name)
	}

	// Valida número de argumentos
	if len(n.children) != len(sig.Params) {
		panic("[Semantic] Número incorreto de argumentos para " + n.name)
	}

	// Avalia os argumentos
	args := make([]*semantic.Variable, len(n.children))
	for i, arg := range n.children {
		args[i] = arg.Evaluate(st)
	}

	// Cria escopo local para a função
	localSt := semantic.NewSymbolTable()
	localSt.Functions = st.Functions // Herda as funções definidas

	// Vincula os parâmetros aos argumentos
	for i, paramName := range sig.Params {
		localSt.CreateVariable(paramName, sig.ParamTypes[i], false)
		localSt.Initialize(paramName, args[i])
	}

	// Avalia o corpo da função
	funcDec := sig.Node.(*FunctionDec)
	result := funcDec.children[0].Evaluate(localSt)

	// Se houve return, retorna o valor retornado. Senão, retorna o resultado do bloco
	if localSt.HasReturned {
		return localSt.ReturnValue
	}
	return result
}

func (n *FunctionCall) Generate(st *semantic.SymbolTable) {
	// Procura a função
	sig, exists := st.Functions[n.name]
	if !exists {
		panic("[Semantic] Função não definida: " + n.name)
	}

	// Avalia os argumentos e empilha (ordem reversa para convenção cdecl)
	for i := len(n.children) - 1; i >= 0; i-- {
		n.children[i].Generate(st)
		CodeGenerator.Append("  push eax ; argumento " + sig.Params[i])
	}

	// Chama a função
	CodeGenerator.Append(fmt.Sprintf("  call %s", n.name))

	// Remove argumentos da pilha
	CodeGenerator.Append(fmt.Sprintf("  add esp, %d ; remove argumentos", len(n.children)*4))

	// Resultado está em EAX
}

// NoOp representa uma operação vazia (dummy)
type NoOp struct {
	children []Node
	id       int
}

func NewNoOp() *NoOp {
	return &NoOp{children: []Node{}, id: GetNextNodeID()}
}

func (n *NoOp) Evaluate(st *semantic.SymbolTable) *semantic.Variable {
	return semantic.NewVoidVariable()
}

func (n *NoOp) Generate(st *semantic.SymbolTable) {
	// Nada a gerar
}

// CodeGen representa o gerador de código Assembly
type CodeGen struct {
	Instructions []string
}

// NewCodeGen cria um novo gerador de código
func NewCodeGen() *CodeGen {
	return &CodeGen{Instructions: []string{}}
}

// Append adiciona uma instrução
func (c *CodeGen) Append(instruction string) {
	c.Instructions = append(c.Instructions, instruction)
}

// Clear limpa as instruções
func (c *CodeGen) Clear() {
	c.Instructions = []string{}
}

// GetInstructions retorna todas as instruções
func (c *CodeGen) GetInstructions() []string {
	return c.Instructions
}
