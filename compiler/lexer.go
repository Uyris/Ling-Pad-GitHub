package compiler

import (
	"strconv"
	"unicode"
)

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
	DOT       = "DOT"
	END       = "END"
	FUNC      = "FUNC"
	RETURN    = "RETURN"
	STRUCT    = "STRUCT"
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

// Token representa um token lexico com tipo e valor
type Token struct {
	tokenType string
	value     interface{}
}

// Lexer realiza a analise lexica do codigo-fonte
type Lexer struct {
	source   string
	position int
	next     Token
}

// NewLexer cria um novo Lexer a partir do codigo-fonte
func NewLexer(source string) *Lexer {
	return &Lexer{
		source:   source,
		position: 0,
	}
}

// SelectNext le o proximo token e atualiza o atributo next
func (l *Lexer) SelectNext() {
	// Ignora espacos em branco
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
	} else if currentChar == '.' {
		l.next = Token{tokenType: DOT, value: "."}
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
			panic("[Lexer] Erro ao converter numero: " + numStr)
		}
		l.next = Token{tokenType: INT, value: num}
	} else if unicode.IsLetter(currentChar) || currentChar == '_' {
		// Identifiers e palavras reservadas
		if currentChar == '_' {
			panic("[Lexer] Identificador nao pode comecar com '_'")
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
		} else if identStr == "struct" {
			l.next = Token{tokenType: STRUCT, value: "struct"}
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
			// "println" sem "!" e um identificador normal
			l.next = Token{tokenType: IDEN, value: identStr}
		} else {
			l.next = Token{tokenType: IDEN, value: identStr}
		}
	} else {
		panic("[Lexer] Invalid Symbol: " + string(currentChar))
	}
}
