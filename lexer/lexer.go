package lexer

import (
	"strconv"
	"unicode"

	"lingpad/token"
)

// Lexer realiza a análise léxica do código-fonte
type Lexer struct {
	Source   string
	Position int
	Next     token.Token
}

// NewLexer cria um novo Lexer a partir do código-fonte
func NewLexer(source string) *Lexer {
	return &Lexer{
		Source:   source,
		Position: 0,
	}
}

// SelectNext lê o próximo token e atualiza o atributo Next
func (l *Lexer) SelectNext() {
	// Ignora espaços em branco
	for l.Position < len(l.Source) && unicode.IsSpace(rune(l.Source[l.Position])) {
		l.Position++
	}

	// Verifica se chegou ao final da string
	if l.Position >= len(l.Source) {
		l.Next = token.Token{TokenType: token.EOF, Value: ""}
		return
	}

	currentChar := rune(l.Source[l.Position])

	if currentChar == '+' {
		l.Next = token.Token{TokenType: token.PLUS, Value: "+"}
		l.Position++
	} else if currentChar == '-' && l.Position+1 < len(l.Source) && l.Source[l.Position+1] == '>' {
		l.Next = token.Token{TokenType: token.ARROW, Value: "->"}
		l.Position += 2
	} else if currentChar == '-' {
		l.Next = token.Token{TokenType: token.MINUS, Value: "-"}
		l.Position++
	} else if currentChar == '^' {
		l.Next = token.Token{TokenType: token.XOR, Value: "^"}
		l.Position++
	} else if currentChar == '*' && l.Position+1 < len(l.Source) && l.Source[l.Position+1] == '*' {
		l.Next = token.Token{TokenType: token.POW, Value: "**"}
		l.Position += 2
	} else if currentChar == '*' {
		l.Next = token.Token{TokenType: token.MULT, Value: "*"}
		l.Position++
	} else if currentChar == '/' {
		l.Next = token.Token{TokenType: token.DIV, Value: "/"}
		l.Position++
	} else if currentChar == '&' && l.Position+1 < len(l.Source) && l.Source[l.Position+1] == '&' {
		l.Next = token.Token{TokenType: token.AND, Value: "&&"}
		l.Position += 2
	} else if currentChar == '|' && l.Position+1 < len(l.Source) && l.Source[l.Position+1] == '|' {
		l.Next = token.Token{TokenType: token.OR, Value: "||"}
		l.Position += 2
	} else if currentChar == '!' {
		l.Next = token.Token{TokenType: token.NOT, Value: "!"}
		l.Position++
	} else if currentChar == '=' && l.Position+1 < len(l.Source) && l.Source[l.Position+1] == '=' {
		l.Next = token.Token{TokenType: token.EQ, Value: "=="}
		l.Position += 2
	} else if currentChar == '=' {
		l.Next = token.Token{TokenType: token.ASSIGN, Value: "="}
		l.Position++
	} else if currentChar == ':' {
		l.Next = token.Token{TokenType: token.COLON, Value: ":"}
		l.Position++
	} else if currentChar == ',' {
		l.Next = token.Token{TokenType: token.COMMA, Value: ","}
		l.Position++
	} else if currentChar == '>' {
		l.Next = token.Token{TokenType: token.GT, Value: ">"}
		l.Position++
	} else if currentChar == '<' {
		l.Next = token.Token{TokenType: token.LT, Value: "<"}
		l.Position++
	} else if currentChar == '(' {
		l.Next = token.Token{TokenType: token.OPEN_PAR, Value: "("}
		l.Position++
	} else if currentChar == ')' {
		l.Next = token.Token{TokenType: token.CLOSE_PAR, Value: ")"}
		l.Position++
	} else if currentChar == '{' {
		l.Next = token.Token{TokenType: token.OPEN_BRA, Value: "{"}
		l.Position++
	} else if currentChar == '}' {
		l.Next = token.Token{TokenType: token.CLOSE_BRA, Value: "}"}
		l.Position++
	} else if currentChar == ';' {
		l.Next = token.Token{TokenType: token.END, Value: ";"}
		l.Position++
	} else if currentChar == '"' {
		l.Position++
		strValue := ""
		for l.Position < len(l.Source) && rune(l.Source[l.Position]) != '"' {
			strValue += string(l.Source[l.Position])
			l.Position++
		}
		if l.Position >= len(l.Source) {
			panic("[Lexer] Unterminated string literal")
		}
		l.Position++
		l.Next = token.Token{TokenType: token.STR, Value: strValue}
	} else if unicode.IsDigit(currentChar) {
		numStr := string(currentChar)
		l.Position++
		for l.Position < len(l.Source) && unicode.IsDigit(rune(l.Source[l.Position])) {
			numStr += string(l.Source[l.Position])
			l.Position++
		}
		num, err := strconv.Atoi(numStr)
		if err != nil {
			panic("[Lexer] Erro ao converter número: " + numStr)
		}
		l.Next = token.Token{TokenType: token.INT, Value: num}
	} else if unicode.IsLetter(currentChar) || currentChar == '_' {
		// Identifiers e palavras reservadas
		if currentChar == '_' {
			panic("[Lexer] Identificador não pode começar com '_'")
		}
		identStr := string(currentChar)
		l.Position++
		for l.Position < len(l.Source) && (unicode.IsLetter(rune(l.Source[l.Position])) || unicode.IsDigit(rune(l.Source[l.Position])) || rune(l.Source[l.Position]) == '_') {
			identStr += string(l.Source[l.Position])
			l.Position++
		}

		// Verificar palavras reservadas
		if identStr == "println" && l.Position < len(l.Source) && rune(l.Source[l.Position]) == '!' {
			l.Position++ // Consumir o "!"
			l.Next = token.Token{TokenType: token.PRINT, Value: "println"}
		} else if identStr == "scanln" && l.Position < len(l.Source) && rune(l.Source[l.Position]) == '!' {
			l.Position++ // Consumir o "!"
			l.Next = token.Token{TokenType: token.READ, Value: "scanln"}
		} else if identStr == "if" {
			l.Next = token.Token{TokenType: token.IF, Value: "if"}
		} else if identStr == "while" {
			l.Next = token.Token{TokenType: token.WHILE, Value: "while"}
		} else if identStr == "else" {
			l.Next = token.Token{TokenType: token.ELSE, Value: "else"}
		} else if identStr == "let" {
			l.Next = token.Token{TokenType: token.LET, Value: "let"}
		} else if identStr == "mut" {
			l.Next = token.Token{TokenType: token.MUT, Value: "mut"}
		} else if identStr == "fn" {
			l.Next = token.Token{TokenType: token.FUNC, Value: "fn"}
		} else if identStr == "return" {
			l.Next = token.Token{TokenType: token.RETURN, Value: "return"}
		} else if identStr == "true" {
			l.Next = token.Token{TokenType: token.BOOL, Value: true}
		} else if identStr == "false" {
			l.Next = token.Token{TokenType: token.BOOL, Value: false}
		} else if identStr == token.TYPE_I32 || identStr == token.TYPE_BOOL || identStr == token.TYPE_STR {
			l.Next = token.Token{TokenType: token.TYPE, Value: identStr}
		} else if identStr == "println" {
			// "println" sem "!" é um identificador normal
			l.Next = token.Token{TokenType: token.IDEN, Value: identStr}
		} else {
			l.Next = token.Token{TokenType: token.IDEN, Value: identStr}
		}
	} else {
		panic("[Lexer] Invalid Symbol: " + string(currentChar))
	}
}
