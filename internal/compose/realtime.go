package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	defaultRefreshEvery = 30 * time.Second
	defaultRetryEvery   = 5 * time.Second
)

func orDefault(d, def time.Duration) time.Duration {
	if d <= 0 {
		return def
	}
	return d
}

// eventsArgs follows the container events of the Docker daemon.
var eventsArgs = []string{"events", "--filter", "type=container", "--format", "json"}

// containerActions are the events that can change the state a Project shows;
// the others (exec_*, archive-path, attach, ...) are noise.
var containerActions = map[string]bool{
	"create": true, "start": true, "restart": true, "die": true, "kill": true,
	"stop": true, "pause": true, "unpause": true, "destroy": true,
	"rename": true, "oom": true, "health_status": true,
}

// projectOfEvent returns the compose project a `docker events` JSON line
// concerns, or "" when the event cannot change a Project's state.
func projectOfEvent(line string) string {
	var e struct {
		Action string `json:"Action"`
		Actor  struct {
			Attributes map[string]string `json:"Attributes"`
		} `json:"Actor"`
	}
	if json.Unmarshal([]byte(line), &e) != nil {
		return ""
	}
	action, _, _ := strings.Cut(e.Action, ":") // "health_status: healthy"
	if !containerActions[strings.TrimSpace(action)] {
		return ""
	}
	return e.Actor.Attributes["com.docker.compose.project"]
}

// Run keeps the snapshot up to date until ctx is done. It refreshes at once,
// then follows `docker events` (a container event refreshes only its Project)
// with a full refresh every RefreshEvery as safety net. While Docker is
// unreachable, or the events stream drops, it retries every RetryEvery and
// resumes everything by itself. Refresh errors are passed to onError when it
// is not nil.
func (m *Manager) Run(ctx context.Context, onError func(error)) {
	for ctx.Err() == nil {
		err := m.Refresh(ctx)
		if err != nil && ctx.Err() == nil && onError != nil {
			onError(err)
		}
		if err == nil && m.Snapshot().DockerReachable {
			m.follow(ctx, onError)
		}
		if !sleep(ctx, m.retryEvery) {
			return
		}
	}
}

// follow consumes the events stream, with the periodic full refresh, until
// the stream ends, a full refresh finds Docker unreachable or ctx is done.
func (m *Manager) follow(ctx context.Context, onError func(error)) {
	sctx, cancel := context.WithCancel(ctx)
	ended := make(chan struct{})
	go func() {
		defer close(ended)
		_ = m.docker.Stream(sctx, eventsArgs, func(line string, stderr bool) {
			if stderr {
				return
			}
			if project := projectOfEvent(line); project != "" {
				if err := m.refreshProject(sctx, project); err != nil && sctx.Err() == nil && onError != nil {
					onError(err)
				}
			}
		})
	}()
	defer func() { // the stream never outlives follow
		cancel()
		<-ended
	}()

	tick := time.NewTicker(m.refreshEvery)
	defer tick.Stop()
	for {
		select {
		case <-sctx.Done():
			return
		case <-ended:
			return
		case <-tick.C:
			err := m.Refresh(sctx)
			if err != nil && sctx.Err() == nil && onError != nil {
				onError(err)
			}
			if err != nil || !m.Snapshot().DockerReachable {
				return
			}
		}
	}
}

// refreshProject re-reads the Containers of one Project and notifies watchers
// if its state changed. A Project that is not in the snapshot yet, or any
// failure, falls back to a full Refresh.
func (m *Manager) refreshProject(ctx context.Context, name string) error {
	m.refreshMu.Lock()
	m.mu.Lock()
	base := m.base
	m.mu.Unlock()
	idx := -1
	for i, p := range base.Projects {
		if p.Name == name {
			idx = i
		}
	}
	if idx < 0 || !base.DockerReachable {
		m.refreshMu.Unlock()
		return m.Refresh(ctx)
	}
	defer m.refreshMu.Unlock()

	out, err := m.docker.Run(ctx, "compose", "-p", name, "ps", "-a", "--format", "json")
	var running, total int
	if err == nil {
		running, total, err = containerCounts(out)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Unknown state: let the full refresh decide (it also detects an
		// unreachable Docker).
		next, berr := m.build(ctx)
		if berr != nil {
			return fmt.Errorf("compose: refresh %s: %w", name, berr)
		}
		m.publish(next)
		return nil
	}
	projects := append([]Project(nil), base.Projects...)
	projects[idx].Running, projects[idx].Total = running, total
	projects[idx].Status = statusOf(running, total)
	m.publish(Snapshot{DockerReachable: true, Projects: projects})
	return nil
}

// sleep waits d or until ctx is done; it reports whether ctx is still alive.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
