package turing

import (
	"context"
	"errors"
	"fmt"
)

// Direction of movement of the carriage along the tape.
type Direction int

// Available carriage move directions.
const (
	Left  Direction = -1
	Right Direction = 1
	Stay  Direction = 0
)

// Program for Turing machine.
type Program map[string]map[rune]Transition

// Validate checks move, write and next state fields for program.
func (tp Program) Validate(alphabet map[rune]struct{}, terminalState string) error {
	for state, stateTransitions := range tp {
		for symbol, transition := range stateTransitions {
			if transition.Move != Left && transition.Move != Right && transition.Move != Stay {
				return fmt.Errorf("%w: %d for state %q, symbol %q", ErrInvalidMoveDirection, transition.Move, state, symbol)
			}

			if _, ok := alphabet[transition.Write]; !ok {
				return fmt.Errorf("%w: %q for state %q", ErrUnexpectedSymbol, transition.Write, state)
			}

			if transition.NextState == terminalState {
				continue
			}

			if _, ok := tp[transition.NextState]; !ok {
				return fmt.Errorf("%w: %q", ErrStateNotFound, transition.NextState)
			}
		}
	}

	return nil
}

type Machine struct {
	// current carriage position
	carriage int

	//           ↓
	// [ ][ ][ ][A][!][ ][ ]
	// infinite tape with carriage
	tape map[int]rune

	// current state (Q1 for example)
	state string

	// the startState from which the algorithm will start
	startState string

	// the terminalState that the algorithm will terminate
	terminalState string

	// "A!" for example
	alphabet map[rune]struct{}

	// program[Q1][A] for example
	program map[string]map[rune]Transition

	// number of executed steps
	steps uint

	// to understand whether the algorithm will go along the tape infinitely
	maxTapeLength uint

	maxSteps uint
}

// A! - alphabet
//		Q1	Q2	Q3	<- states
//	A	*
//	!
//	^ symbols, * - field (A>Q2 for example)

type Transition struct {
	NextState string
	Move      Direction

	// symbol to write in a cell that carriage look at
	Write rune
}

var (
	// ErrStartStateEmpty is returned when the start state parameter is empty.
	ErrStartStateEmpty = errors.New("start state is empty")

	// ErrTerminalStateEmpty is returned when the terminal state parameter is empty.
	ErrTerminalStateEmpty = errors.New("terminal state is empty")

	// ErrInvalidMaxTapeLength is returned when maxTapeLength is zero.
	ErrInvalidMaxTapeLength = errors.New("invalid max tape length")

	// ErrInvalidMoveDirection is returned when a transition has an invalid move direction.
	ErrInvalidMoveDirection = errors.New("invalid move direction")

	// ErrStateNotFound is returned when a transition references a non-existent state.
	ErrStateNotFound = errors.New("state not found")
)

// NewMachine creates a new Turing machine with the specified configuration.
// Space character is automatically included in the alphabet.
// To avoid max steps constraint pass 0.
func NewMachine(
	alphabet, // "ABC" for example, space is already included
	startState, // Q1 for example
	terminalState string, // Q0 for example
	program Program,
	maxTapeLength,
	maxSteps uint, // pass 0 to disable
) (*Machine, error) {
	a := make(map[rune]struct{}, len(alphabet))
	for _, sym := range alphabet {
		a[sym] = struct{}{}
	}

	a[' '] = struct{}{}

	if startState == "" {
		return nil, ErrStartStateEmpty
	}

	if terminalState == "" {
		return nil, ErrTerminalStateEmpty
	}

	if maxTapeLength == 0 {
		return nil, ErrInvalidMaxTapeLength
	}

	if err := program.Validate(a, terminalState); err != nil {
		return nil, err
	}

	return &Machine{
		tape:          make(map[int]rune),
		startState:    startState,
		terminalState: terminalState,
		alphabet:      a,
		program:       program,
		maxTapeLength: maxTapeLength,
		maxSteps:      maxSteps,
	}, nil
}

func (m *Machine) Copy() *Machine {
	return &Machine{
		carriage:      m.carriage,
		tape:          m.tape,
		state:         m.state,
		startState:    m.startState,
		terminalState: m.terminalState,
		alphabet:      m.alphabet,
		program:       m.program,
		steps:         m.steps,
		maxTapeLength: m.maxTapeLength,
		maxSteps:      m.maxSteps,
	}
}

// Exec executes the Turing machine program with the starting carriage position
// and input tape, returning the final tape state upon completion or an error
// if execution fails.
func (m *Machine) Exec(carriage int, input map[int]rune) (map[int]rune, error) {
	return m.ExecCtx(context.Background(), carriage, input)
}

// ExecCtx executes the Turing machine program with the given context, starting carriage position,
// and input tape. The method initializes the machine state, copies the input to the internal tape,
// and runs the computation step by step until the machine halts or encounters an error.
// The context allows for cancellation of long-running computations.
// Returns the final tape state upon successful completion, or an error if the execution fails
// or the context is cancelled.
func (m *Machine) ExecCtx(ctx context.Context, carriage int, input map[int]rune) (map[int]rune, error) {
	m.carriage = carriage
	m.state = m.startState
	m.steps = 0

	tape := make(map[int]rune, len(input))
	for i, symbol := range input {
		tape[i] = symbol
	}

	m.tape = tape

	var (
		ok  = true
		err error
	)

	for ok {
		if ctx.Err() != nil {
			return nil, ctx.Err() //nolint:wrapcheck
		}

		ok, err = m.step()
		if err != nil {
			return nil, err
		}
	}

	return m.tape, nil
}

var (
	// ErrTransitionNotFound is returned when no transition is defined for the current state and symbol.
	ErrTransitionNotFound = errors.New("transition not found")

	// ErrInfiniteLoop is returned when the machine enters an infinite loop.
	ErrInfiniteLoop = errors.New("infinite loop")

	// ErrStepsExceeded is returned when the machine exceeds the maximum number of execution steps.
	ErrStepsExceeded = errors.New("steps exceeded")

	// ErrUnexpectedSymbol is returned when the machine reads a symbol not in its alphabet.
	ErrUnexpectedSymbol = errors.New("unexpected symbol")

	// ErrTapeOver is returned when the tape exceeds its maximum allowed length.
	ErrTapeOver = errors.New("tape is over")
)

func (m *Machine) step() (bool, error) {
	if m.state == m.terminalState {
		return false, nil
	}

	sym := m.read()

	if _, ok := m.alphabet[sym]; !ok {
		return false, fmt.Errorf("%w: %q", ErrUnexpectedSymbol, sym)
	}

	transition, ok := m.program[m.state][sym]
	if !ok {
		return false, fmt.Errorf("%w: state %q, symbol %q", ErrTransitionNotFound, m.state, sym)
	}

	// is current transition an infinite loop?
	if transition.Move == Stay && transition.NextState == m.state && transition.Write == sym {
		return false, fmt.Errorf("%w: state %q, symbol %q", ErrInfiniteLoop, m.state, sym)
	}

	m.write(transition.Write)
	m.move(transition.Move)
	m.state = transition.NextState
	m.steps++

	if uint(len(m.tape)) >= m.maxTapeLength {
		return false, fmt.Errorf("%w, carriage: %d", ErrTapeOver, m.carriage)
	}

	if m.maxSteps > 0 && m.steps >= m.maxSteps {
		return false, ErrStepsExceeded
	}

	return true, nil
}

// The read method allows reading a symbol from the cell the carriage points to.
// If there is no symbol at this position, it returns ' '.
func (m *Machine) read() rune {
	if sym, ok := m.tape[m.carriage]; ok {
		return sym
	}

	return ' '
}

func (m *Machine) write(sym rune) {
	m.tape[m.carriage] = sym
}

func (m *Machine) move(d Direction) {
	m.carriage += int(d)
}
