package parser

import (
	"lingpad/ast"
	"lingpad/lexer"
	"lingpad/token"
)

// parserLexer é o atributo estático (variável de pacote) do Parser
var parserLexer *lexer.Lexer

// ParseProgram é o entry point do parser
func ParseProgram() ast.Node {
	block := ast.NewBlock()

	for parserLexer.Next.TokenType != token.EOF {
		// Verifica se é uma declaração de função
		if parserLexer.Next.TokenType == token.FUNC {
			funcDec := ParseFunctionDec()
			block.AddChild(funcDec)
		} else {
			stmt := ParseStatement()
			block.AddChild(stmt)
		}
	}

	return block
}

func ParseFunctionDec() ast.Node {
	if parserLexer.Next.TokenType != token.FUNC {
		panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected fn")
	}
	parserLexer.SelectNext()

	// Nome da função
	if parserLexer.Next.TokenType != token.IDEN {
		panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected function name")
	}
	funcName := parserLexer.Next.Value.(string)
	parserLexer.SelectNext()

	// Parâmetros: ( [IDEN : TYPE [, IDEN : TYPE]*] )
	if parserLexer.Next.TokenType != token.OPEN_PAR {
		panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected OPEN_PAR")
	}
	parserLexer.SelectNext()

	params := []string{}
	paramTypes := []string{}

	for parserLexer.Next.TokenType != token.CLOSE_PAR {
		if parserLexer.Next.TokenType != token.IDEN {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected parameter name")
		}
		paramName := parserLexer.Next.Value.(string)
		params = append(params, paramName)
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.COLON {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected COLON")
		}
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.TYPE {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected TYPE")
		}
		paramType := parserLexer.Next.Value.(string)
		paramTypes = append(paramTypes, paramType)
		parserLexer.SelectNext()

		// Verifica se há mais parâmetros
		if parserLexer.Next.TokenType == token.COMMA || parserLexer.Next.TokenType == token.END {
			// Compatibilidade: alguns casos podem ter vírgula
			if parserLexer.Next.TokenType == token.COMMA {
				parserLexer.SelectNext()
			} else if parserLexer.Next.TokenType == token.END {
				break
			}
		}
	}

	if parserLexer.Next.TokenType != token.CLOSE_PAR {
		panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
	}
	parserLexer.SelectNext()

	// Verifica se há return type: -> TYPE ou -> ()
	if parserLexer.Next.TokenType == token.ARROW {
		parserLexer.SelectNext()
		// Pula o return type (pode ser TYPE ou OPEN_PAR para void)
		if parserLexer.Next.TokenType == token.OPEN_PAR {
			// Syntax: -> ()
			parserLexer.SelectNext()
			if parserLexer.Next.TokenType != token.CLOSE_PAR {
				panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR in return type")
			}
			parserLexer.SelectNext()
		} else if parserLexer.Next.TokenType == token.TYPE {
			// Syntax: -> i32, -> bool, etc
			parserLexer.SelectNext()
		} else {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected TYPE or () in return type")
		}
	}

	// Corpo da função
	body := ParseBlock()

	return ast.NewFunctionDec(funcName, params, paramTypes, body)
}

func ParseBlock() ast.Node {
	if parserLexer.Next.TokenType != token.OPEN_BRA {
		panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected OPEN_BRA")
	}
	parserLexer.SelectNext()

	block := ast.NewBlock()
	for parserLexer.Next.TokenType != token.CLOSE_BRA {
		if parserLexer.Next.TokenType == token.EOF {
			panic("[Parser] Unexpected EOF, expected CLOSE_BRA")
		}
		stmt := ParseStatement()
		block.AddChild(stmt)
	}

	parserLexer.SelectNext()
	return block
}

func ParseStatement() ast.Node {
	// Bloco: { STATEMENT* }
	if parserLexer.Next.TokenType == token.OPEN_BRA {
		return ParseBlock()
	}

	// Declaração: let [mut] IDENTIFIER : TYPE [= BOOLEXPRESSION] ;
	if parserLexer.Next.TokenType == token.LET {
		parserLexer.SelectNext()

		isMutable := false
		if parserLexer.Next.TokenType == token.MUT {
			isMutable = true
			parserLexer.SelectNext()
		}

		if parserLexer.Next.TokenType != token.IDEN {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected IDEN")
		}
		name := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.COLON {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected COLON")
		}
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.TYPE {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected TYPE")
		}
		declaredType := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()

		var expr ast.Node = nil
		if parserLexer.Next.TokenType == token.ASSIGN {
			parserLexer.SelectNext()
			expr = ParseBoolExpression()
		}

		if parserLexer.Next.TokenType != token.END {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected END (;) ")
		}
		parserLexer.SelectNext()

		return ast.NewVarDec(declaredType, isMutable, ast.NewIdentifier(name), expr)
	}

	// RETURN: return [BOOLEXPRESSION] ;
	if parserLexer.Next.TokenType == token.RETURN {
		parserLexer.SelectNext()

		var expr ast.Node = nil
		if parserLexer.Next.TokenType != token.END {
			expr = ParseBoolExpression()
		}

		if parserLexer.Next.TokenType != token.END {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected END (;)")
		}
		parserLexer.SelectNext()

		return ast.NewReturnNode(expr)
	}

	// IF: if ( BOOLEXPRESSION ) STATEMENT [else STATEMENT]
	if parserLexer.Next.TokenType == token.IF {
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		condition := ParseBoolExpression()

		if parserLexer.Next.TokenType != token.CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		thenBranch := ParseStatement()

		if parserLexer.Next.TokenType == token.ELSE {
			parserLexer.SelectNext()
			elseBranch := ParseStatement()
			return ast.NewIfNode(condition, thenBranch, elseBranch)
		}

		return ast.NewIfNode(condition, thenBranch, nil)
	}

	// WHILE: while ( BOOLEXPRESSION ) STATEMENT
	if parserLexer.Next.TokenType == token.WHILE {
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		condition := ParseBoolExpression()

		if parserLexer.Next.TokenType != token.CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		body := ParseStatement()
		return ast.NewWhileNode(condition, body)
	}

	// Atribuição ou chamada de função: IDENTIFIER [= BOOLEXPRESSION | ( ... ) ]  ;
	if parserLexer.Next.TokenType == token.IDEN {
		name := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()

		// Verifica se é chamada de função
		if parserLexer.Next.TokenType == token.OPEN_PAR {
			// Chamada de função como statement
			parserLexer.SelectNext()

			// Paisa os argumentos
			args := []ast.Node{}
			for parserLexer.Next.TokenType != token.CLOSE_PAR {
				args = append(args, ParseBoolExpression())

				if parserLexer.Next.TokenType == token.COMMA {
					parserLexer.SelectNext()
				} else if parserLexer.Next.TokenType != token.CLOSE_PAR {
					panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected COMMA or CLOSE_PAR")
				}
			}

			if parserLexer.Next.TokenType != token.CLOSE_PAR {
				panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
			}
			parserLexer.SelectNext()

			if parserLexer.Next.TokenType != token.END {
				panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected END (;)")
			}
			parserLexer.SelectNext()

			return ast.NewFunctionCall(name, args)
		}

		// Atribuição: IDENTIFIER = BOOLEXPRESSION ;
		if parserLexer.Next.TokenType != token.ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected ASSIGN")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()

		if parserLexer.Next.TokenType != token.END {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected END (;)")
		}
		parserLexer.SelectNext()

		return ast.NewAssignment(ast.NewIdentifier(name), expr)
	}

	// Impressão: PRINT ( BOOLEXPRESSION ) ;
	if parserLexer.Next.TokenType == token.PRINT {
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()

		if parserLexer.Next.TokenType != token.CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.END {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected END (;)")
		}
		parserLexer.SelectNext()

		return ast.NewPrint(expr)
	}

	// Linha vazia: ;
	if parserLexer.Next.TokenType == token.END {
		parserLexer.SelectNext()
		return ast.NewNoOp()
	}

	panic("[Parser] Unexpected token in statement: " + parserLexer.Next.TokenType)
}

func ParseBoolExpression() ast.Node {
	result := ParseBoolTerm()

	for parserLexer.Next.TokenType == token.OR {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		right := ParseBoolTerm()
		result = ast.NewBinOp(op, result, right)
	}

	return result
}

func ParseBoolTerm() ast.Node {
	result := ParseRelExpression()

	for parserLexer.Next.TokenType == token.AND {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		right := ParseRelExpression()
		result = ast.NewBinOp(op, result, right)
	}

	return result
}

func ParseRelExpression() ast.Node {
	left := ParseExpression()

	if parserLexer.Next.TokenType == token.EQ || parserLexer.Next.TokenType == token.GT || parserLexer.Next.TokenType == token.LT {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		right := ParseExpression()
		return ast.NewBinOp(op, left, right)
	}

	return left
}

func ParseExpression() ast.Node {
	result := ParseTerm()

	for parserLexer.Next.TokenType == token.PLUS || parserLexer.Next.TokenType == token.MINUS || parserLexer.Next.TokenType == token.XOR {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		right := ParseTerm()
		result = ast.NewBinOp(op, result, right)
	}

	return result
}

func ParseTerm() ast.Node {
	result := ParseUnary()

	for parserLexer.Next.TokenType == token.MULT || parserLexer.Next.TokenType == token.DIV {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		right := ParseUnary()
		result = ast.NewBinOp(op, result, right)
	}

	return result
}

func ParseUnary() ast.Node {
	if parserLexer.Next.TokenType == token.PLUS || parserLexer.Next.TokenType == token.MINUS || parserLexer.Next.TokenType == token.NOT {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		operand := ParseUnary()
		return ast.NewUnOp(op, operand)
	}

	return ParsePower()
}

func ParsePower() ast.Node {
	base := ParseFactor()

	if parserLexer.Next.TokenType == token.POW {
		op := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		exponent := ParseUnary()
		return ast.NewBinOp(op, base, exponent)
	}

	return base
}

func ParseFactor() ast.Node {
	// Parênteses
	if parserLexer.Next.TokenType == token.OPEN_PAR {
		parserLexer.SelectNext()
		result := ParseBoolExpression()
		if parserLexer.Next.TokenType != token.CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()
		return result
	}

	// Leitura: scanln!()
	if parserLexer.Next.TokenType == token.READ {
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.OPEN_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected OPEN_PAR")
		}
		parserLexer.SelectNext()

		if parserLexer.Next.TokenType != token.CLOSE_PAR {
			panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
		}
		parserLexer.SelectNext()

		return ast.NewReadNode()
	}

	// Booleano
	if parserLexer.Next.TokenType == token.BOOL {
		value := parserLexer.Next.Value.(bool)
		parserLexer.SelectNext()
		return ast.NewBoolVal(value)
	}

	// String
	if parserLexer.Next.TokenType == token.STR {
		value := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()
		return ast.NewStringVal(value)
	}

	// Identificador ou chamada de função
	if parserLexer.Next.TokenType == token.IDEN {
		name := parserLexer.Next.Value.(string)
		parserLexer.SelectNext()

		// Verifica se é uma chamada de função
		if parserLexer.Next.TokenType == token.OPEN_PAR {
			parserLexer.SelectNext()

			// Paisa os argumentos
			args := []ast.Node{}
			for parserLexer.Next.TokenType != token.CLOSE_PAR {
				args = append(args, ParseBoolExpression())

				if parserLexer.Next.TokenType == token.COMMA {
					parserLexer.SelectNext()
				} else if parserLexer.Next.TokenType != token.CLOSE_PAR {
					panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected COMMA or CLOSE_PAR")
				}
			}

			if parserLexer.Next.TokenType != token.CLOSE_PAR {
				panic("[Parser] Unexpected token: " + parserLexer.Next.TokenType + ", expected CLOSE_PAR")
			}
			parserLexer.SelectNext()

			return ast.NewFunctionCall(name, args)
		}

		// Caso contrário, é apenas um identificador (variável)
		return ast.NewIdentifier(name)
	}

	// Número
	if parserLexer.Next.TokenType == token.INT {
		value := parserLexer.Next.Value.(int)
		parserLexer.SelectNext()
		return ast.NewIntVal(value)
	}

	panic("[Parser] Unexpected token in factor: " + parserLexer.Next.TokenType)
}

// Parse é o ponto de entrada do Parser. Retorna a raiz da AST.
func Parse(code string) ast.Node {
	lex := lexer.NewLexer(code)
	parserLexer = lex
	parserLexer.SelectNext()

	result := ParseProgram()

	if parserLexer.Next.TokenType != token.EOF {
		panic("[Parser] Unexpected token after program: " + parserLexer.Next.TokenType)
	}

	return result
}
