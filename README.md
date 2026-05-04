# Ling-Pad-GitHub

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)

## Diagrama Sintático

**Descricao:** A linguagem suporta atribuicao, impressao, condicionais, lacos, leitura de terminal, funcoes com retorno tipado, escopo de variaveis e expressoes aritmeticas/booleanas.

**EBNF:**
```ebnf
PROGRAM = { FUNCDEC | VARDEC } ;
FUNCDEC = "fn", IDENTIFIER, "(", [PARAMS], ")", ["->", (TYPE | "(", ")")], BLOCK ;
PARAMS = IDENTIFIER, ":", TYPE, {",", IDENTIFIER, ":", TYPE} ;
VARDEC = "let", [MUT], IDENTIFIER, ":", TYPE, ["=", BOOLEXPRESSION], ";" ;
BLOCK = "{", { STATEMENT }, "}" ;
STATEMENT = VARDEC
          | BLOCK
          | IF, "(", BOOLEXPRESSION, ")", STATEMENT, [ELSE, STATEMENT]
          | WHILE, "(", BOOLEXPRESSION, ")", STATEMENT
          | RETURN, BOOLEXPRESSION, ";"
          | PRINT, "(", BOOLEXPRESSION, ")", ";"
          | IDENTIFIER, ("=", BOOLEXPRESSION | "(", [BOOLEXPRESSION, {",", BOOLEXPRESSION}], ")"), ";"
          | ";" ;
BOOLEXPRESSION = BOOLTERM, { "||", BOOLTERM } ;
BOOLTERM = RELEXPRESSION, { "&&", RELEXPRESSION } ;
RELEXPRESSION = EXPRESSION, [("==" | "<" | ">"), EXPRESSION] ;
EXPRESSION = TERM, { ("+" | "-" | "^"), TERM } ;
TERM = UNARY, { ("*" | "/"), UNARY } ;
UNARY = ("+" | "-" | "!"), UNARY | POWER ;
POWER = FACTOR, ["**", UNARY] ;
FACTOR = "(", BOOLEXPRESSION, ")"
       | NUMBER
       | BOOLEAN
       | STRING
       | IDENTIFIER, ["(", [BOOLEXPRESSION, {",", BOOLEXPRESSION}], ")"]
       | READ, "(", ")" ;

IF = "if" ;
ELSE = "else" ;
WHILE = "while" ;
LET = "let" ;
MUT = "mut" ;
PRINT = "println!" ;
READ = "scanln!" ;
FUNC = "fn" ;
RETURN = "return" ;

BOOLEAN = "true" | "false" ;
TYPE = "i32" | "bool" | "str" ;
NUMBER = "number" ;
STRING = "string" ;
IDENTIFIER = "identifier" ;
```