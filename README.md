# go-turing

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)

A simple Go implementation of a Turing machine.

## Features

- Turing machine implementation with configurable alphabet and states.
- Built-in verification mechanisms (infinite loop detection, step limits, tape size limits)
- File-based program loading from `.tur` files
- Examples (addition, multiplication, increment)
- Zero external dependencies (except testing)

## Installation

```bash
go get github.com/asphodex/go-turing
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/asphodex/go-turing"
)

func main() {
    // Create a simple "add one" program
    program := turing.Program{
        "Q1": {
            ' ': {NextState: "Q0", Move: turing.Stay, Write: '1'},
            '1': {NextState: "Q1", Move: turing.Left, Write: '1'},
        },
    }

    // Create machine with alphabet "1", states Q1->Q0, max 20 tape cells, 10 steps
    machine, err := turing.NewMachine("1", "Q1", "Q0", program, 20, 10)
    if err != nil {
        panic(err)
    }

    // Input: "111" (represents number 2 in unary)
    input := map[int]rune{0: '1', 1: '1', 2: '1'}
    
    // Execute starting at position 0
    result, err := machine.Exec(0, input)
    if err != nil {
        panic(err)
    }

    fmt.Println("Result:", result) // Output: "1111" (number 3 in unary)
}
```

## Examples

### Addition Program

Computes `x + y` in unary number system, where numbers are separated by '+':

```go
program := turing.Program{
    "Q1": {
        '1': {NextState: "Q2", Move: turing.Right, Write: ' '},
    },
    "Q2": {
        ' ': {NextState: "Q3", Move: turing.Left, Write: ' '},
        '1': {NextState: "Q2", Move: turing.Right, Write: '1'},
        '+': {NextState: "Q2", Move: turing.Right, Write: '1'},
    },
    "Q3": {
        '1': {NextState: "Q0", Move: turing.Stay, Write: ' '},
    },
}

machine, _ := turing.NewMachine("1+", "Q1", "Q0", program, 100, 100)

// Input: "11+111" (1 + 2 in unary)
input := map[int]rune{-2: '1', -1: '1', 0: '+', 1: '1', 2: '1', 3: '1'}
result, _ := machine.Exec(-2, input) // Result: "1111" (3 in unary)
```

### Loading from File

```go
import "github.com/asphodex/go-turing/filereader"

// Load program from .tur file
program, err := filereader.ReadFileCtx(context.Background(), "program.tur")
if err != nil {
    panic(err)
}

machine, err := turing.NewMachine("1xy", "Q1", "Q0", program, 1000, 1000)
// ... use machine
```

### Error Types

- `ErrStartStateEmpty`: Start state parameter is empty
- `ErrTerminalStateEmpty`: Terminal state parameter is empty  
- `ErrInvalidMaxTapeLength`: Max tape length is zero
- `ErrInvalidMoveDirection`: Invalid move direction in transition
- `ErrStateNotFound`: Transition references non-existent state
- `ErrTransitionNotFound`: No transition defined for current state/symbol
- `ErrInfiniteLoop`: Machine detected infinite loop
- `ErrStepsExceeded`: Execution exceeded maximum steps
- `ErrUnexpectedSymbol`: Symbol not in machine's alphabet
- `ErrTapeOver`: Tape exceeded maximum length

## File Format (.tur files)

Program file support https://kpolyakov.spb.ru/prog/turing.htm

## Testing

Run the test suite:

```bash
go test ./...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```