# Ling-Pad-GitHub

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)

## Diagrama Sintático
![alt text](public/img/image-2.png)


**Descricao:** A linguagem suporta atribuicao, impressao, condicionais, lacos, leitura de terminal e expressoes aritmeticas/booleanas.

**EBNF:**
```ebnf
PROGRAM = { STATEMENT } ;
STATEMENT = ((IF, "(", BOOLEXPRESSION, ")", STATEMENT, ("ELSE", STATEMENT) | ε) | (WHILE, "(", BOOLEXPRESSION, ")", STATEMENT) | (IDENTIFIER, "=", BOOLEXPRESSION) | (PRINT, "(", BOOLEXPRESSION, ")") | ε), EOL ;
BOOLEXPRESSION = BOOLTERM, { "||", BOOLTERM } ;
BOOLTERM = RELEXPRESSION, { "&&", RELEXPRESSION } ;
RELEXPRESSION = EXPRESSION, ("==" | "<" | ">"), EXPRESSION ;
EXPRESSION = TERM, { ("+" | "-"), TERM } ;
TERM = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR = ("+" | "-"), FACTOR | "(", BOOLEXPRESSION, ")" | NUMBER | READ, "(", ")" ;
NUMBER = DIGIT, {DIGIT} ;
DIGIT = 0 | 1 | ... | 9 ;
IDENTIFIER = LETTER, {LETTER | DIGIT | "_"} ;
LETTER = a | b | ... | z | A | B | ... | Z ;
```