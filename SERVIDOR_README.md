# Instruções para Compilar no Servidor Linux

## Problema
O compilador foi desenvolvido no Windows, gerando um executável `.exe`. No servidor Linux, é necessário compilar um novo binário específico para o sistema operacional Linux.

## Solução

### Opção 1: Build Local no Servidor (Recomendado)

Execute no servidor Linux:

```bash
cd /path/to/ling-pad

# Compilar para Linux 32-bit
go build -o main main.go

# Testar com um arquivo
./main /tmp/compiler-testing-lib/compiler_testing_lib/languages/Rust/x3.0/test001.ling
```

**Resultado esperado:**
- O programa executará e imprimirá: `5 4 3 2 120`
- Um arquivo `/tmp/compiler-testing-lib/compiler_testing_lib/languages/Rust/x3.0/test001.asm` será criado

### Opção 2: Build Automático via Script

Se você estiver no servidor com bash disponível:

```bash
chmod +x build-linux.sh
./build-linux.sh
```

## Estrutura de Diretórios no Servidor

O projeto deve estar neste formato:

```
ling-pad/
├── main.go
├── go.mod
├── token/
│   └── token.go
├── lexer/
│   └── lexer.go
├── preprocessor/
│   └── preprocessor.go
├── semantic/
│   └── semantic.go
├── ast/
│   └── ast.go
├── parser/
│   └── parser.go
└── codegen/
    └── codegen.go
```

## Troubleshooting

### Erro: "cannot find package"

Verifique se o `go.mod` está configurado corretamente:

```bash
cat go.mod
# Deve conter: module lingpad
```

### Erro: "permission denied"

Certifique-se que o binário tem permissão de execução:

```bash
chmod +x main
```

### Arquivo ASM não é gerado

Execute com mais verbosidade:

```bash
./main /tmp/test.ling 2>&1
```

Você verá logs [Main] DEBUG mostrando exatamente onde o processo falha.

## Fluxo de Execução

1. **Entrada**: `test001.ling` 
2. **Processamento**:
   - Leitura do arquivo
   - Pré-processamento (remove comentários)
   - Análise léxica
   - Análise sintática
   - Avaliação semântica + Execução
   - Geração de código assembly
3. **Saída**: `test001.asm` no mesmo diretório da entrada
4. **Montagem**: 
   ```bash
   nasm -f elf32 -o test001.o test001.asm
   gcc -m32 -no-pie -nostartfiles -o test001 test001.o -e _start
   ```

## Verificação Final

Para confirmar que tudo funciona:

```bash
# Compilar o ASM em objeto
nasm -f elf32 -o test001.o test001.asm

# Linkar em executável
gcc -m32 -no-pie -nostartfiles -o test001 test001.o -e _start

# Executar
./test001
# Deve imprimir: 5 4 3 2 120
```
