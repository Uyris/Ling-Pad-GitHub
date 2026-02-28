# Ling-Pad-GitHub

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)

## Diagrama Sintático

```
expression:
  ┌─────────────────────────────────────────────────────┐
  │                                                     │
  ▼                                                     │
 INT ──► ( '+' ──► INT )  ──────────────────────────► EOF
          ( '-' ──► INT )
```

O diagrama representa a gramática:

```
expression = INT ( ('+' | '-') INT )*
```
