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
	holds     map[string]chan struct{}
	follows   map[string]bool
	calls     [][]string
}

type response struct {
	out    []byte
	stderr []byte // streamed after out, flagged as stderr
	err    error
}

// New returns an empty Fake: any call without a configured response fails.
func New() *Fake {
	return &Fake{responses: map[string]response{}, holds: map[string]chan struct{}{}, follows: map[string]bool{}}
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

// RespondStream configures a streamed command: stdout lines, then stderr
// lines, then err as the command's exit error.
func (f *Fake) RespondStream(stdout, stderr []byte, err error, args ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses[key(args)] = response{out: stdout, stderr: stderr, err: err}
}

// Hold makes the command (streamed or not) with exactly these arguments block,
// once it has started, until the returned release function is called or its context is
// cancelled. It lets a test observe an Activity while it is running.
func (f *Fake) Hold(args ...string) (release func()) {
	ch := make(chan struct{})
	f.mu.Lock()
	f.holds[key(args)] = ch
	f.mu.Unlock()
	var once sync.Once
	return func() { once.Do(func() { close(ch) }) }
}

// Follow makes the streamed command with exactly these arguments stay open
// after replaying its output, until its context is cancelled, like
// `docker events` does.
func (f *Fake) Follow(args ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.follows[key(args)] = true
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
func (f *Fake) Run(ctx context.Context, args ...string) ([]byte, error) {
	r, err := f.lookup(args)
	f.mu.Lock()
	hold := f.holds[key(args)]
	f.mu.Unlock()
	if hold != nil {
		select {
		case <-hold:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return r.out, err
}

// Stream implements "run streaming": it replays the canned stdout then stderr
// line by line, after waiting for a Hold if one is set.
func (f *Fake) Stream(ctx context.Context, args []string, onLine func(line string, stderr bool)) error {
	r, err := f.lookup(args)
	f.mu.Lock()
	hold := f.holds[key(args)]
	follow := f.follows[key(args)]
	f.mu.Unlock()
	if hold != nil {
		select {
		case <-hold:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	for _, part := range []struct {
		data   []byte
		stderr bool
	}{{r.out, false}, {r.stderr, true}} {
		sc := bufio.NewScanner(bytes.NewReader(part.data))
		sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for sc.Scan() {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			onLine(sc.Text(), part.stderr)
		}
	}
	if follow {
		<-ctx.Done()
		return ctx.Err()
	}
	return err
}
