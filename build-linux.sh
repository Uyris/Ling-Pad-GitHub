#!/bin/bash
# Script para compilar o compilador LingPad para Linux

set -e

echo "[Build] Compilando LingPad para Linux..."

# Navegar até o diretório do projeto (adaptado para seu caminho)
cd "$(dirname "$0")"

# Compilar com Go
export GOOS=linux
export GOARCH=386

go build -o main main.go

if [ -f main ]; then
    echo "[Build] ✓ Compilação bem-sucedida!"
    echo "[Build] Arquivo executável: main"
    echo "[Build] Para usar: ./main /caminho/para/arquivo.ling"
else
    echo "[Build] ✗ Erro: executável não foi criado"
    exit 1
fi
