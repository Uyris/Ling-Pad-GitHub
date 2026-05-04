package compiler

// Parser realiza a analise sintatica consumindo tokens do Lexer
type Parser struct{}

// lexer e o atributo estatico (variavel de pacote) do Parser
var parserLexer *Lexer

func ParseProgram() Node {
	block := NewBlock()

	for parserLexer.next.tokenType != EOF {
		if parserLexer.next.tokenType == STRUCT {
			block.AddChild(ParseStructDeclaration())
			if parserLexer.next.tokenType == END {
				parserLexer.SelectNext()
			}
			continue
		}
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

func ParseTypeName() string {
	if parserLexer.next.tokenType == TYPE || parserLexer.next.tokenType == IDEN {
		typeName := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		return typeName
	}
	panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected TYPE or IDEN")
}

func ParseStructDeclaration() Node {
	// Declaracao: struct IDENTIFIER { FIELD* }
	if parserLexer.next.tokenType != STRUCT {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected STRUCT")
	}
	parserLexer.SelectNext()

	if parserLexer.next.tokenType != IDEN {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected IDEN")
	}
	name := parserLexer.next.value.(string)
	parserLexer.SelectNext()

	if parserLexer.next.tokenType != OPEN_BRA {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected OPEN_BRA")
	}
	parserLexer.SelectNext()

	fields := []*VarDec{}
	for parserLexer.next.tokenType != CLOSE_BRA {
		if parserLexer.next.tokenType == EOF {
			panic("[Parser] Unexpected EOF, expected CLOSE_BRA")
		}
		if parserLexer.next.tokenType != LET {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected LET")
		}
		fields = append(fields, ParseStructField())
	}

	parserLexer.SelectNext()
	return NewStructDec(name, fields)
}

func ParseStructField() *VarDec {
	// Campo: let [mut] IDENTIFIER : TYPE ;
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

	fieldType := ParseTypeName()

	if parserLexer.next.tokenType != END {
		panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
	}
	parserLexer.SelectNext()

	return NewVarDec(fieldType, isMutable, NewIdentifier(name), nil)
}

func ParseVarDeclaration() Node {
	// Declaracao: let [mut] IDENTIFIER : TYPE [= BOOLEXPRESSION] ;
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

	declaredType := ParseTypeName()

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

			paramType := ParseTypeName()

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
		if parserLexer.next.tokenType == TYPE || parserLexer.next.tokenType == IDEN {
			returnType = ParseTypeName()
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

func ParseFieldAccessChain(base Node) Node {
	current := base
	for parserLexer.next.tokenType == DOT {
		parserLexer.SelectNext()
		if parserLexer.next.tokenType != IDEN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected IDEN")
		}
		fieldName := parserLexer.next.value.(string)
		parserLexer.SelectNext()
		current = NewFieldAccess(current, fieldName)
	}
	return current
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

	// Declaracao: let [mut] IDENTIFIER : TYPE [= BOOLEXPRESSION] ;
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

	// Atribuicao: IDENTIFIER = BOOLEXPRESSION ;
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

		lvalue := ParseFieldAccessChain(NewIdentifier(name))
		if parserLexer.next.tokenType != ASSIGN {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected ASSIGN")
		}
		parserLexer.SelectNext()

		expr := ParseBoolExpression()

		if parserLexer.next.tokenType != END {
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
		}
		parserLexer.SelectNext()

		return NewAssignment(lvalue, expr)
	}

	// Impressao: PRINT ( BOOLEXPRESSION ) ;
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
			panic("[Parser] Unexpected token: " + parserLexer.next.tokenType + ", expected END (;) ")
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
	// Parenteses
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
		return ParseFieldAccessChain(NewIdentifier(name))
	}

	// Numero
	if parserLexer.next.tokenType == INT {
		value := parserLexer.next.value.(int)
		parserLexer.SelectNext()
		return NewIntVal(value)
	}

	panic("[Parser] Unexpected token in factor: " + parserLexer.next.tokenType)
}

// Run e o ponto de entrada do Parser. Retorna a raiz da AST.
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
