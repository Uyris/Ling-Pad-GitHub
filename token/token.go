package token

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
	ARROW     = "ARROW"
	COLON     = "COLON"
	COMMA     = "COMMA"
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
	FUNC      = "FUNC"
	RETURN    = "RETURN"
	EOF       = "EOF"
)

// Token representa um token léxico com tipo e valor
type Token struct {
	TokenType string
	Value     interface{}
}

// Tipos de tipos de dados suportados
const (
	TYPE_I32  = "i32"
	TYPE_BOOL = "bool"
	TYPE_STR  = "str"
	TYPE_VOID = "void"
)
