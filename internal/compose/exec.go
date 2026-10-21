package compose

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// ExecDocker is the production adapter of the Docker port: it runs the Docker
// CLI as an external process (ADR-0002).
type ExecDocker struct {
	// Binary is the executable to run; "docker" when empty.
	Binary string
}

func (d ExecDocker) binary() string {
	if d.Binary == "" {
		return "docker"
	}
	return d.Binary
}

// Run implements Docker.
func (d ExecDocker) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, d.binary(), args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), fmt.Errorf("%s %s: %w: %s", d.binary(), strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// Stream implements Docker. The process is killed when ctx is cancelled.
func (d ExecDocker) Stream(ctx context.Context, args []string, onLine func(line string, stderr bool)) error {
	cmd := exec.CommandContext(ctx, d.binary(), args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s %s: %w", d.binary(), strings.Join(args, " "), err)
	}

	var mu sync.Mutex // onLine is never called concurrently
	var wg sync.WaitGroup
	pump := func(r io.Reader, isErr bool) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for sc.Scan() {
			mu.Lock()
			onLine(sc.Text(), isErr)
			mu.Unlock()
		}
	}
	wg.Add(2)
	go pump(stdout, false)
	go pump(stderr, true)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%s %s: %w", d.binary(), strings.Join(args, " "), err)
	}
	return nil
}
