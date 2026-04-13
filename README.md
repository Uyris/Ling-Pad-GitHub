# Ling-Pad-GitHub

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)

## Diagrama Sintático
![alt text](image.png)


**Descricao:** A linguagem suporta atribuicao, impressao, condicionais, lacos, leitura de terminal e expressoes aritmeticas/booleanas.

**EBNF:**
```ebnf
PROGRAM = { STATEMENT } ;
STATEMENT = BLOCK
          | IF, "(", BOOLEXPRESSION, ")", STATEMENT, [ELSE, STATEMENT]
          | WHILE, "(", BOOLEXPRESSION, ")", STATEMENT
          | LET, [MUT], IDENTIFIER, ":", TYPE, ["=", BOOLEXPRESSION], ";"
          | IDENTIFIER, "=", BOOLEXPRESSION, ";"
          | PRINT, "(", BOOLEXPRESSION, ")", ";"
          | ";" ;
BLOCK = "{", { STATEMENT }, "}" ;
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
       | IDENTIFIER
       | READ, "(", ")" ;

IF = "if" ;
ELSE = "else" ;
WHILE = "while" ;
LET = "let" ;
MUT = "mut" ;
PRINT = "println!" ;
READ = "scanln!" ;

BOOLEAN = "true" | "false" ;
TYPE = "i32" | "bool" | "str" ;
NUMBER = "number" ;
STRING = "string" ;
IDENTIFIER = "identifier" ;
```