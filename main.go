package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		panic("No input provided")
	}

	input := os.Args[1]
	// Remove all spaces
	input = strings.ReplaceAll(input, " ", "")

	if len(input) == 0 {
		panic("Empty input")
	}

	result := 0
	currentNumber := ""
	operation := '+'

	for _, ch := range input {
		if ch >= '0' && ch <= '9' {
			currentNumber += string(ch)
		} else if ch == '+' || ch == '-' {
			if currentNumber == "" {
				panic("Invalid expression: operator without number")
			}
			num, err := strconv.Atoi(currentNumber)
			if err != nil {
				panic(err)
			}
			if operation == '+' {
				result += num
			} else {
				result -= num
			}
			currentNumber = ""
			operation = ch
		} else {
			panic("Invalid character: " + string(ch))
		}
	}

	// Process the last number
	if currentNumber == "" {
		panic("Invalid expression: ends with operator")
	}
	num, err := strconv.Atoi(currentNumber)
	if err != nil {
		panic(err)
	}
	if operation == '+' {
		result += num
	} else {
		result -= num
	}

	fmt.Println(result)
}
