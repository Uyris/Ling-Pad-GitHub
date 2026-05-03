package semantic

import (
	"strconv"

	"lingpad/token"
)

// Variable representa uma variável com seu valor e deslocamento na pilha
type Variable struct {
	Value   interface{}
	VarType string
	Mutable bool
	Shift   int // Deslocamento relativo ao EBP na pilha (em bytes)
}

// NewVariable cria uma nova variável
func NewVariable(value interface{}, varType string, mutable bool) *Variable {
	return &Variable{Value: value, VarType: varType, Mutable: mutable, Shift: 0}
}

func (v *Variable) Clone() *Variable {
	return &Variable{Value: v.Value, VarType: v.VarType, Mutable: v.Mutable, Shift: v.Shift}
}

func NewVoidVariable() *Variable {
	return NewVariable(nil, token.TYPE_VOID, false)
}

func DefaultValueForType(varType string) interface{} {
	switch varType {
	case token.TYPE_I32:
		return 0
	case token.TYPE_BOOL:
		return false
	case token.TYPE_STR:
		return ""
	default:
		panic("[Semantic] Unknown type: " + varType)
	}
}

func VariableToString(v *Variable) string {
	switch v.VarType {
	case token.TYPE_STR:
		return v.Value.(string)
	case token.TYPE_I32:
		return strconv.Itoa(v.Value.(int))
	case token.TYPE_BOOL:
		if v.Value.(bool) {
			return "true"
		}
		return "false"
	default:
		panic("[Semantic] Cannot convert type to string: " + v.VarType)
	}
}

// FunctionSignature armazena informações sobre uma função
type FunctionSignature struct {
	Name       string
	Params     []string    // nomes dos parâmetros
	ParamTypes []string    // tipos dos parâmetros
	Node       interface{} // *ast.FunctionDec - guardado como interface{} para evitar ciclo de importação
}

// SymbolTable armazena variáveis e seus valores
type SymbolTable struct {
	Table         map[string]*Variable
	Functions     map[string]*FunctionSignature // funções definidas
	NextShift     int                           // Rastreia o próximo deslocamento na pilha
	VariableCount int                           // Número de variáveis declaradas
	ReturnValue   *Variable                     // Valor retornado pela função
	HasReturned   bool                          // Flag para indicar que um return foi encontrado
}

// NewSymbolTable cria uma nova tabela de símbolos
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Table:     make(map[string]*Variable),
		Functions: make(map[string]*FunctionSignature),
		NextShift: -4, // Primeira variável em [EBP-4] (negativo para escopo global)
	}
}

// Get retorna o valor de uma variável
func (st *SymbolTable) Get(name string) *Variable {
	if variable, exists := st.Table[name]; exists {
		return variable
	}
	panic("[Semantic] Variável não definida: " + name)
}

func (st *SymbolTable) Exists(name string) bool {
	_, exists := st.Table[name]
	return exists
}

// CreateVariable declara uma variável na tabela
func (st *SymbolTable) CreateVariable(name string, varType string, mutable bool) {
	if st.Exists(name) {
		panic("[Semantic] Variável já declarada: " + name)
	}
	variable := NewVariable(DefaultValueForType(varType), varType, mutable)
	variable.Shift = st.NextShift // Atribui o shift atual
	st.Table[name] = variable
	st.NextShift -= 4 // Próxima variável será 4 bytes adiante (mais negativa)
	st.VariableCount++
}

// Initialize define o valor inicial no momento da declaração
func (st *SymbolTable) Initialize(name string, value *Variable) {
	variable := st.Get(name)
	if variable.VarType != value.VarType {
		panic("[Semantic] Type mismatch na inicialização de " + name)
	}
	variable.Value = value.Value
}

// Set atualiza uma variável já declarada
func (st *SymbolTable) Set(name string, value *Variable) {
	if !st.Exists(name) {
		panic("[Semantic] Variável não declarada: " + name)
	}

	variable := st.Table[name]
	if !variable.Mutable {
		panic("[Semantic] Variável imutável: " + name)
	}

	if variable.VarType != value.VarType {
		panic("[Semantic] Type mismatch em atribuição para " + name)
	}

	variable.Value = value.Value
}
