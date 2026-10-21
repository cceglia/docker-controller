// Package composefake is the test adapter of the compose Docker port: it
// answers with canned (real, captured) output and records the arguments it
// receives. It is safe for concurrent use.
package composefake

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
)

// Fake implements compose.Docker. It does not import compose: the port uses
// only primitive types so any adapter can satisfy it structurally.
type Fake struct {
	mu        sync.Mutex
	responses map[string]response
	calls     [][]string
}

type response struct {
	out []byte
	err error
}

// New returns an empty Fake: any call without a configured response fails.
func New() *Fake {
	return &Fake{responses: map[string]response{}}
}

func key(args []string) string { return strings.Join(args, "\x00") }

// Respond configures the output returned for the exact argument list.
func (f *Fake) Respond(out []byte, args ...string) {
	f.RespondErr(out, nil, args...)
}

// RespondErr configures output plus an error for the exact argument list.
func (f *Fake) RespondErr(out []byte, err error, args ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses[key(args)] = response{out: out, err: err}
}

// Calls returns every argument list received so far, in order.
func (f *Fake) Calls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]string, len(f.calls))
	for i, c := range f.calls {
		out[i] = append([]string(nil), c...)
	}
	return out
}

func (f *Fake) lookup(args []string) (response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, append([]string(nil), args...))
	r, ok := f.responses[key(args)]
	if !ok {
		return response{}, fmt.Errorf("composefake: unexpected docker call %q", strings.Join(args, " "))
	}
	return r, r.err
}

// Run implements "run and return the output".
func (f *Fake) Run(_ context.Context, args ...string) ([]byte, error) {
	r, err := f.lookup(args)
	return r.out, err
}

// Stream implements "run streaming": it replays the canned output line by
// line. stderr is always false in the fake.
func (f *Fake) Stream(ctx context.Context, args []string, onLine func(line string, stderr bool)) error {
	r, err := f.lookup(args)
	sc := bufio.NewScanner(bytes.NewReader(r.out))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		onLine(sc.Text(), false)
	}
	return err
}
