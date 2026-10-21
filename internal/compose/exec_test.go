package compose_test

import (
	"context"
	"strings"
	"testing"

	"dockercontroller/internal/compose"
)

// The adapter is exercised with `sh` standing in for the docker binary.
func TestExecDocker_RunReturnsStdoutAndStderrOnFailure(t *testing.T) {
	d := compose.ExecDocker{Binary: "sh"}

	out, err := d.Run(context.Background(), "-c", "echo hello")
	if err != nil || strings.TrimSpace(string(out)) != "hello" {
		t.Fatalf("out=%q err=%v", out, err)
	}

	_, err = d.Run(context.Background(), "-c", "echo boom >&2; exit 3")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error should carry stderr, got %v", err)
	}
}

func TestExecDocker_StreamDeliversLinesWithOrigin(t *testing.T) {
	d := compose.ExecDocker{Binary: "sh"}
	var got []string
	err := d.Stream(context.Background(), []string{"-c", "echo a; echo b >&2"}, func(line string, stderr bool) {
		got = append(got, map[bool]string{false: "out:", true: "err:"}[stderr]+line)
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, ",")
	if !strings.Contains(joined, "out:a") || !strings.Contains(joined, "err:b") {
		t.Fatalf("got %v", got)
	}
}
