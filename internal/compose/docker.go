package compose

import "context"

// Docker is the port towards the Docker CLI (ADR-0002). args never contain the
// "docker" executable itself: the adapter prepends it.
type Docker interface {
	// Run executes the command and returns its stdout. When the command fails
	// the error carries the process stderr.
	Run(ctx context.Context, args ...string) ([]byte, error)
	// Stream executes the command and delivers its output line by line until
	// it ends or ctx is cancelled.
	Stream(ctx context.Context, args []string, onLine func(line string, stderr bool)) error
}
