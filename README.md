# Ling-Pad-GitHub

[![Compilation Status](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)](https://compiler-tester.insper-comp.com.br/svg/Uyris/Ling-Pad-GitHub)

## Diagrama Sintático

```
EXPRESSION:

         ┌──────────────────────────────────────┐
         │                                      │
         ▼                                      │
──▶ [ NUMBER ] ──▶ [ + ] ──▶ [ NUMBER ] ──▶─────┤
                  │                             │
                  └──▶ [ - ] ──▶ [ NUMBER ] ──▶─┘
                                                │
                                                ▼
                                              (end)
```

**Descrição:** A expressão começa com um `NUMBER`, seguido de zero ou mais operações `+` ou `-`, cada uma seguida por outro `NUMBER`.
