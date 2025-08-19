package turing_test

import (
	"github.com/asphodex/go-turing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewMachine_Valid(t *testing.T) {
	t.Parallel()

	program := turing.Program{
		"Q1": {' ': {NextState: "Q0", Move: turing.Stay, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"ABC",
		"Q1",
		"Q0",
		program,
		7,
		10,
	)
	require.NoError(t, err)
	assert.NotNil(t, machine)
}

func TestNewMachine_EmptyStartState(t *testing.T) {
	t.Parallel()

	program := turing.Program{
		"Q1": {' ': {NextState: "Q0", Move: turing.Stay, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"ABC",
		"",
		"Q0",
		program,
		7,
		10,
	)
	require.ErrorIs(t, err, turing.ErrStartStateEmpty)
	assert.Nil(t, machine)
}

func TestNewMachine_InvalidTapeLength(t *testing.T) {
	t.Parallel()

	transitions := turing.Program{
		"Q1": {' ': {NextState: "Q0", Move: turing.Stay, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"ABC",
		"Q1",
		"Q0",
		transitions,
		0,
		10,
	)
	require.ErrorIs(t, err, turing.ErrInvalidMaxTapeLength)
	assert.Nil(t, machine)
}

func TestNewMachine_EmptyTerminalState(t *testing.T) {
	t.Parallel()

	program := turing.Program{
		"Q1": {' ': {NextState: "Q0", Move: turing.Stay, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"ABC",
		"Q1",
		"",
		program,
		7,
		10,
	)
	require.ErrorIs(t, err, turing.ErrTerminalStateEmpty)
	assert.Nil(t, machine)
}

func TestProgram_Validate(t *testing.T) {
	t.Parallel()

	alphabet := map[rune]struct{}{
		' ': {},
	}

	tt := []struct {
		name string
		p    turing.Program
		err  error
	}{
		{
			name: "valid program",
			p: turing.Program{
				"Q1": {' ': {NextState: "Q0", Move: turing.Stay, Write: ' '}},
			},
			err: nil,
		},
		{
			name: "return err on invalid next state",
			p: turing.Program{
				"Q1": {' ': {NextState: "Q2", Move: turing.Stay, Write: ' '}},
			},
			err: turing.ErrStateNotFound,
		},
		{
			name: "return err on invalid move",
			p: turing.Program{
				"Q1": {' ': {NextState: "Q0", Move: -2, Write: ' '}},
			},
			err: turing.ErrInvalidMoveDirection,
		},
		{
			name: "return err on unexpected symbol in write field",
			p: turing.Program{
				"Q1": {' ': {NextState: "Q2", Move: turing.Stay, Write: '1'}},
			},
			err: turing.ErrUnexpectedSymbol,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorIs(t, tc.p.Validate(alphabet, "Q0"), tc.err)
		})
	}
}

func TestMachine_Exec_Plus_One_Program(t *testing.T) {
	t.Parallel()

	program := turing.Program{
		"Q1": {
			' ': {NextState: "Q0", Move: turing.Stay, Write: '1'},
			'1': {NextState: "Q1", Move: turing.Left, Write: '1'},
		},
	}

	machine, err := turing.NewMachine(
		"1",
		"Q1",
		"Q0",
		program,
		20,
		10,
	)
	require.NoError(t, err)

	numberToUnary := func(number int) string {
		// in unary number system zero is '1', one is '11' and etc
		return strings.Repeat("1", number+1)
	}
	tapeToUnary := func(tape map[int]rune) (string, error) {
		var sb strings.Builder

		for _, symbol := range tape {
			if symbol != '1' {
				return "", turing.ErrUnexpectedSymbol
			}

			sb.WriteRune(symbol)
		}

		return sb.String(), nil
	}

	carriage := 0

	for i := 0; i < 10; i++ {
		input := make(map[int]rune)

		// build input 111...
		for k := 0; k <= i; k++ {
			input[k] = '1'
		}

		tape, err := machine.Exec(carriage, input)
		require.NoError(t, err)

		result, err := tapeToUnary(tape)
		require.NoError(t, err)

		// f(x)=x+1
		assert.Equal(t, numberToUnary(i+1), result)
	}
}

func TestMachine_Exec_Addition_Program(t *testing.T) {
	t.Parallel()

	// This program for a Turing machine,
	// calculates the function f(x)=x+y in the unary number system.
	// The numbers and are written in the specified order and
	// separated from each other by the symbol '+'.
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

	machine, err := turing.NewMachine(
		"1+",
		"Q1",
		"Q0",
		program,
		100,
		100,
	)
	require.NoError(t, err)

	numberToUnary := func(number int) string {
		// in unary number system zero is '1', one is '11' and etc
		return strings.Repeat("1", number+1)
	}
	tapeToUnary := func(tape map[int]rune) (string, error) {
		var sb strings.Builder

		for _, symbol := range tape {
			if symbol != '1' {
				if symbol != ' ' {
					return "", turing.ErrUnexpectedSymbol
				}

				continue
			}

			sb.WriteRune(symbol)
		}

		return sb.String(), nil
	}

	for i := 1; i < 10; i++ {
		for j := 1; j < 10; j++ {
			input := make(map[int]rune)

			// build input: 111...1+111...1
			for k := 1; k <= i; k++ {
				input[-k] = '1'
			}

			input[0] = '+'

			for k := 1; k <= j; k++ {
				input[k] = '1'
			}

			// by default carriage is looking at last left
			// non-empty cell
			tape, err := machine.Exec(-i, input)
			require.NoError(t, err)

			result, err := tapeToUnary(tape)
			require.NoError(t, err)

			assert.Equal(t, numberToUnary((i-1)+(j-1)), result)
		}
	}
}

func TestMachine_Exec_Multiply_Program(t *testing.T) {
	t.Parallel()

	// This program for a Turing machine calculates the
	// function f(x)=3*x in the unary number system.
	program := turing.Program{
		"Q1": {
			'1': {NextState: "Q2", Move: turing.Right, Write: '*'},
		},
		"Q2": {
			' ': {NextState: "Q3", Move: turing.Left, Write: ' '},
			'1': {NextState: "Q2", Move: turing.Right, Write: '1'},
		},
		"Q3": {
			'1': {NextState: "Q4", Move: turing.Left, Write: ' '},
			'*': {NextState: "Q0", Move: turing.Stay, Write: '1'},
		},
		"Q4": {
			' ': {NextState: "Q5", Move: turing.Left, Write: '1'},
			'1': {NextState: "Q4", Move: turing.Left, Write: '1'},
			'*': {NextState: "Q4", Move: turing.Left, Write: '*'},
		},
		"Q5": {
			' ': {NextState: "Q6", Move: turing.Left, Write: '1'},
		},
		"Q6": {
			' ': {NextState: "Q7", Move: turing.Right, Write: '1'},
		},
		"Q7": {
			' ': {NextState: "Q3", Move: turing.Left, Write: ' '},
			'1': {NextState: "Q7", Move: turing.Right, Write: '1'},
			'*': {NextState: "Q7", Move: turing.Right, Write: '*'},
		},
	}

	machine, err := turing.NewMachine(
		"1*",
		"Q1",
		"Q0",
		program,
		1000,
		1000,
	)
	require.NoError(t, err)

	numberToUnary := func(number int) string {
		// in unary number system zero is '1', one is '11' and etc
		return strings.Repeat("1", number+1)
	}
	tapeToUnary := func(tape map[int]rune) (string, error) {
		var sb strings.Builder

		for _, symbol := range tape {
			if symbol != '1' {
				if symbol != ' ' {
					return "", turing.ErrUnexpectedSymbol
				}

				continue
			}

			sb.WriteRune(symbol)
		}

		return sb.String(), nil
	}

	carriage := 0

	for i := 0; i < 10; i++ {
		input := make(map[int]rune, i)

		// build input 111...
		for k := 0; k <= i; k++ {
			input[k] = '1'
		}

		tape, err := machine.Exec(carriage, input)
		require.NoError(t, err)

		result, err := tapeToUnary(tape)
		require.NoError(t, err)

		// f(x)=3*x
		assert.Equal(t, numberToUnary(i*3), result)
	}
}

func TestMachine_Exec_StepsExceed(t *testing.T) {
	t.Parallel()

	// Go left infinitely.
	program := turing.Program{
		"Q1": {' ': {NextState: "Q1", Move: turing.Left, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"",
		"Q1",
		"Q0",
		program,
		10000,
		10,
	)
	require.NoError(t, err)

	tape, err := machine.Exec(0, map[int]rune{})
	require.ErrorIs(t, err, turing.ErrStepsExceeded)
	assert.Nil(t, tape)
}

func TestMachine_Exec_TapeOver(t *testing.T) {
	t.Parallel()

	// Go left infinitely.
	program := turing.Program{
		"Q1": {' ': {NextState: "Q1", Move: turing.Left, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"",
		"Q1",
		"Q0",
		program,
		10,
		10000,
	)
	require.NoError(t, err)

	tape, err := machine.Exec(0, map[int]rune{})
	require.ErrorIs(t, err, turing.ErrTapeOver)
	assert.Nil(t, tape)
}

func TestMachine_Exec_InfiniteLoop(t *testing.T) {
	t.Parallel()

	// Just infinite loop.
	program := turing.Program{
		"Q1": {' ': {NextState: "Q1", Move: turing.Stay, Write: ' '}},
	}

	machine, err := turing.NewMachine(
		"",
		"Q1",
		"Q0",
		program,
		10000,
		0, // disable maxSteps constraint
	)
	require.NoError(t, err)

	tape, err := machine.Exec(0, map[int]rune{})
	require.ErrorIs(t, err, turing.ErrInfiniteLoop)
	assert.Nil(t, tape)
}
