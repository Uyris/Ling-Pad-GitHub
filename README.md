# Ling-Pad-GitHub

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)

## Diagrama Sintático
<img width="300" height="300" alt="image" src="https://github.com/user-attachments/assets/a73e0f4a-3528-4ba8-bb39-da3deb9d65d0" />


**Descrição:** A expressão começa com um `NUMBER`, seguido de zero ou mais operações `+` ou `-`, cada uma seguida por outro `NUMBER`.

a EBNF equivalente do compilador é:
```bash
EXPRESSION = TERM, { ("+" | "-"), TERM } ;
TERM = FACTOR, { ("*" | "/"), FACTOR } ;
FACTOR = ("+" | "-"), FACTOR | "(", EXPRESSION, ")" | NUMBER ;
NUMBER = DIGIT, {DIGIT} ;
DIGIT = 0 | 1 | ... | 9 ;
```