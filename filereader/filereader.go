// Package filereader reads Turing machine programs from .tur files structured as follows:
// 1. Program comment section;
// 2. Program definition section;
// 3. State table comment section;
// 4. Saved tape section (optional).
//
// Program definition format:
// <Set of states>
// <Symbol from alphabet>\t<transition>\t<transition>...
// <Symbol from alphabet>\t<transition>...
// Where transitions are tab-delimited and each alphabet symbol begins a new line of
// its corresponding transitions.
package filereader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/asphodex/go-turing"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ReadFileCtx reads file from given filepath and returns turing.Program in case of success,
// else returns an error.
func ReadFileCtx(ctx context.Context, filePath string) (program turing.Program, err error) {
	path := filepath.Clean(filePath)

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("file %q does not exist: %w", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}

	defer func() {
		_ = file.Close()
	}()

	return ReadCtx(ctx, file)
}

var (
	// ErrParseTransition is returned when a transition field cannot be parsed correctly.
	ErrParseTransition = errors.New("parse transition")

	// ErrNoTransitions is returned when the program file contains no valid transitions.
	ErrNoTransitions = errors.New("no transitions")
)

// ParseTransition parse field like 1>Q2 and returns the decomposed parts of the field.
func ParseTransition(field string) (turing.Transition, error) {
	const transitionFieldsCount = 2

	directionTable := map[rune]turing.Direction{
		'>': turing.Right,
		'<': turing.Left,
		'.': turing.Stay,
	}

	for sep, dir := range directionTable {
		if strings.ContainsRune(field, sep) {
			fields := strings.Split(field, string(sep))

			// 1>Q1
			if len(fields) != transitionFieldsCount || fields[0] == "" || fields[1] == "" {
				return turing.Transition{}, fmt.Errorf("%w: %s", ErrParseTransition, field)
			}

			write, _ := utf8.DecodeRuneInString(fields[0])
			if write == '_' {
				write = ' '
			}

			return turing.Transition{
				NextState: "Q" + fields[1],
				Move:      dir,
				Write:     write,
			}, nil
		}
	}

	return turing.Transition{}, fmt.Errorf("%w: no direction found", ErrParseTransition)
}

// ReadCtx read .tur files from the given io.Reader.
func ReadCtx(ctx context.Context, r io.Reader) (turing.Program, error) {
	scanner := bufio.NewScanner(r)

	var (
		program = make(turing.Program)
		states  []string
		inScope bool
	)

	statePattern := regexp.MustCompile(`Q\d+`)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err() //nolint:wrapcheck
		}

		line := scanner.Text()

		fields := strings.Split(line, "\t")

		if len(fields) == 0 {
			continue
		}

		if !inScope {
			if len(fields) > 1 && statePattern.MatchString(strings.Join(fields[1:], " ")) {
				states = fields[1:]
				inScope = true
			}

			continue
		}

		if fields[0] == "" {
			continue
		}

		symbol, _ := utf8.DecodeRuneInString(fields[0])

		stateIndex := 0

		for i := 1; i < len(fields); i++ {
			if fields[i] == "" {
				stateIndex++
				continue
			}

			transition, err := ParseTransition(fields[i])
			if err != nil {
				return nil, err
			}

			if _, ok := program[states[stateIndex]]; !ok {
				program[states[stateIndex]] = make(map[rune]turing.Transition)
			}

			program[states[stateIndex]][symbol] = transition

			stateIndex++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read program: %w", err)
	}

	if len(program) == 0 {
		return nil, ErrNoTransitions
	}

	return program, nil
}
