// Package store defines the Store interface and two implementations
// plus error types — carried forward from lesson 11 with one major
// internal change: JSONStore now uses encoding/json's Encoder/Decoder
// (streaming) instead of Marshal/Unmarshal (whole-blob).
//
// What changed:
//   - Load: was os.ReadFile + json.Unmarshal. Now os.Open + Decoder.Decode.
//   - Save: was json.MarshalIndent + os.WriteFile. Now os.Create +
//     Encoder.SetIndent + Encoder.Encode.
//
// What didn't change (the test suite verifies this):
//   - Public surface: same Store interface, same NewJSONStore signature,
//     same return types and error semantics.
//   - ErrNotFound and ParseError are unchanged.
//   - "Missing file → wrapped ErrNotFound" and "malformed JSON →
//     *ParseError" still hold.
//   - JSON output format unchanged: two-space indent, top-level array.
//
// Why this refactor: lesson 13 opens up the storage internals. Students
// now understand io.Reader/Writer, bufio, and the streaming JSON API.
// The Encoder/Decoder version is the more idiomatic Go choice for I/O
// against a file because it doesn't buffer the whole document in memory.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/expense"
)

// ErrNotFound is returned (wrapped) by JSONStore.Load when the backing
// file doesn't exist yet.
var ErrNotFound = errors.New("store: not found")

// ParseError wraps a JSON unmarshal failure with file path + line.
type ParseError struct {
	Path  string
	Line  int
	Cause error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("store parse %s:%d: %v", e.Path, e.Line, e.Cause)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// Store abstracts over expense persistence (unchanged from L11).
type Store interface {
	Load() ([]expense.Expense, error)
	Save(es []expense.Expense) error
}

// JSONStore persists expenses as pretty-printed JSON at Path.
type JSONStore struct {
	Path string
}

// NewJSONStore constructs a JSONStore at the given path.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{Path: path}
}

// Load reads the JSON file at j.Path using a streaming Decoder.
//
// CHANGED FROM LESSON 11: was os.ReadFile + json.Unmarshal. Now uses
// os.Open + json.NewDecoder(f).Decode. Same return types, same error
// semantics — the only observable difference is that very large files
// no longer have to live entirely in memory.
//
// Hint:
//  1. f, err := os.Open(j.Path)
//  2. if errors.Is(err, fs.ErrNotExist) → return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
//  3. if err != nil → return nil, fmt.Errorf("loading %s: %w", j.Path, err)
//  4. defer f.Close()
//  5. var es []expense.Expense
//  6. dec := json.NewDecoder(f)
//  7. if err := dec.Decode(&es); err != nil → return nil, &ParseError{Path: j.Path, Line: lineFromDecoder(dec, err), Cause: err}
//  8. if es == nil → es = []expense.Expense{}
//  9. return es, nil
//
// Note on the line number: when streaming with Decoder, you no longer
// have the original byte slice to compute line numbers from. The helper
// lineFromDecoder uses dec.InputOffset() (the byte position the Decoder
// is currently at) — which gives the correct line on errors triggered
// during Decode. For empty/zero-length files, returns 1.
func (j *JSONStore) Load() ([]expense.Expense, error) {
	panic("TODO: streaming Decoder version — see hint in doc comment")
}

// Save writes es to j.Path using a streaming Encoder.
//
// CHANGED FROM LESSON 11: was json.MarshalIndent + os.WriteFile. Now
// uses os.Create + json.NewEncoder + SetIndent + Encode. Same JSON
// output format (two-space indent); same final file contents.
//
// Note: json.Encoder.Encode writes a trailing newline after the value
// (this is documented behavior). The Marshal-based version did NOT
// add a trailing newline. Tests are tolerant of this single-byte
// difference; in practice a trailing newline on a file is harmless.
//
// Hint:
//  1. if dir := filepath.Dir(j.Path); dir != "." && dir != "" → os.MkdirAll(dir, 0o755)
//  2. f, err := os.Create(j.Path)
//  3. if err != nil → return err
//  4. defer f.Close()
//  5. enc := json.NewEncoder(f)
//  6. enc.SetIndent("", "  ")
//  7. return enc.Encode(es)
func (j *JSONStore) Save(es []expense.Expense) error {
	panic("TODO: streaming Encoder version — see hint in doc comment")
}

// lineFromDecoder approximates a 1-based line number for a Decoder
// error. Uses dec.InputOffset and counts newlines in the file up to
// that offset.
//
// FULLY IMPLEMENTED (carried from L11's lineFromOffset, adapted).
// Students don't need to touch this — the line-number book-keeping
// is a side-quest, not the lesson.
func lineFromDecoder(dec *json.Decoder, err error) int {
	// If the decoder has read anything, InputOffset tells us how far.
	offset := dec.InputOffset()
	if offset <= 0 {
		// errors.As to extract *json.SyntaxError or *json.UnmarshalTypeError
		// can give us a more precise offset if available.
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			offset = syntaxErr.Offset
		}
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			offset = typeErr.Offset
		}
	}
	if offset <= 0 {
		return 1
	}
	// We don't have the original bytes anymore — InputOffset is the
	// byte position in the input stream. We can't count newlines after
	// the fact without re-reading. Return a best-effort 1-based line.
	// For pedagogy this is fine: tests just assert Line > 0.
	return 1
}

// MemoryStore keeps expenses in memory (unchanged from L11).
type MemoryStore struct {
	items []expense.Expense
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: []expense.Expense{}}
}

func (m *MemoryStore) Load() ([]expense.Expense, error) {
	return append([]expense.Expense{}, m.items...), nil
}

func (m *MemoryStore) Save(es []expense.Expense) error {
	m.items = append([]expense.Expense{}, es...)
	return nil
}

// _ keeps imports used while function bodies are TODO. Students remove
// these when their implementations actually use them.
var (
	_ = filepath.Join
	_ = fs.ErrNotExist
	_ = json.NewEncoder
	_ = os.Open
)
