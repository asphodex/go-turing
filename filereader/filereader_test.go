package filereader_test

import (
	"context"
	"github.com/asphodex/go-turing"
	"github.com/asphodex/go-turing/filereader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//nolint:paralleltest
func TestReadFileCtx_ValidFile(t *testing.T) {
	testFilePath := filepath.Join("testdata", "valid_turing.tur")
	assert.FileExists(t, testFilePath)

	ctx := context.Background()
	transitions, err := filereader.ReadFileCtx(ctx, testFilePath)
	require.NoError(t, err)
	assert.Len(t, transitions, 34)
}

//nolint:paralleltest
func TestReadFileCtx_ValidFileWithInput(t *testing.T) {
	testFilePath := filepath.Join("testdata", "valid_turing_with_input.tur")
	assert.FileExists(t, testFilePath)

	ctx := context.Background()
	transitions, err := filereader.ReadFileCtx(ctx, testFilePath)
	require.NoError(t, err)
	assert.Len(t, transitions, 71)
}

//nolint:paralleltest
func TestReadFileCtx_NoFile(t *testing.T) {
	ctx := context.Background()
	transitions, err := filereader.ReadFileCtx(ctx, "invalid_path")
	require.ErrorIs(t, err, os.ErrNotExist)
	assert.Nil(t, transitions)
}

func TestReadCtx_InvalidData(t *testing.T) {
	t.Parallel()

	data := "Q1 Q2"

	ctx := context.Background()
	transitions, err := filereader.ReadCtx(ctx, strings.NewReader(data))
	require.ErrorIs(t, err, filereader.ErrNoTransitions)
	assert.Empty(t, transitions)
}

func TestParseTransition(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name  string
		field string

		transition turing.Transition
		err        error
	}{
		{
			name:  "parse valid field with valid right direction",
			field: "1>2",
			transition: turing.Transition{
				NextState: "Q2",
				Write:     '1',
				Move:      turing.Right,
			},
			err: nil,
		},
		{
			name:  "parse valid field with valid left direction",
			field: "1<3",
			transition: turing.Transition{
				NextState: "Q3",
				Write:     '1',
				Move:      turing.Left,
			},
			err: nil,
		},
		{
			name:  "parse valid field with valid stay direction",
			field: "1.2",
			transition: turing.Transition{
				NextState: "Q2",
				Write:     '1',
				Move:      turing.Stay,
			},
			err: nil,
		},
		{
			name:       "return error on invalid direction",
			field:      "1!2",
			transition: turing.Transition{},
			err:        filereader.ErrParseTransition,
		},
		{
			name:       "return error on empty field",
			field:      "",
			transition: turing.Transition{},
			err:        filereader.ErrParseTransition,
		},
		{
			name:       "return error on invalid field without direction",
			field:      "Q2",
			transition: turing.Transition{},
			err:        filereader.ErrParseTransition,
		},
		{
			name:       "return error on invalid transition fields count",
			field:      "Q2>",
			transition: turing.Transition{},
			err:        filereader.ErrParseTransition,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transition, err := filereader.ParseTransition(tc.field)
			assert.Equal(t, tc.transition, transition)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
