package compose

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Config wires a Manager.
type Config struct {
	// Docker is the port towards the Docker CLI.
	Docker Docker
	// DB is the app database, already migrated (see package db).
	DB *sql.DB
}

// Manager is the interface of the compose module.
type Manager struct {
	docker Docker
	store  store

	refreshMu sync.Mutex // serialises Refresh

	mu       sync.Mutex // guards snap and watchers
	snap     Snapshot
	watchers []chan Snapshot
}

// New creates a Manager. The snapshot is empty until the first Refresh.
func New(cfg Config) (*Manager, error) {
	if cfg.Docker == nil || cfg.DB == nil {
		return nil, errors.New("compose: Docker and DB are required")
	}
	return &Manager{
		docker: cfg.Docker,
		store:  store{db: cfg.DB},
		snap:   Snapshot{Projects: []Project{}},
	}, nil
}

// Snapshot returns the last computed snapshot. It must be treated as
// read-only.
func (m *Manager) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snap
}

// Watch subscribes to snapshots. The current snapshot is delivered first, then
// a new one only when something changed. A slow reader only ever misses
// intermediate snapshots, never the latest. The channel is closed when ctx is
// done.
func (m *Manager) Watch(ctx context.Context) <-chan Snapshot {
	ch := make(chan Snapshot, 1)
	m.mu.Lock()
	ch <- m.snap
	m.watchers = append(m.watchers, ch)
	m.mu.Unlock()

	go func() {
		<-ctx.Done()
		m.mu.Lock()
		defer m.mu.Unlock()
		for i, w := range m.watchers {
			if w == ch {
				m.watchers = append(m.watchers[:i], m.watchers[i+1:]...)
				break
			}
		}
		close(ch)
	}()
	return ch
}

// Refresh asks Docker for the current state, saves the newly discovered
// Projects, rebuilds the snapshot and notifies watchers if it changed. An
// unreachable Docker is reported in the snapshot, not as an error.
func (m *Manager) Refresh(ctx context.Context) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	next, err := m.build(ctx)
	if err != nil {
		return err
	}
	m.publish(next)
	return nil
}

func (m *Manager) build(ctx context.Context) (Snapshot, error) {
	out, err := m.docker.Run(ctx, "compose", "ls", "-a", "--format", "json")
	if err != nil {
		if ctx.Err() != nil {
			return Snapshot{}, ctx.Err()
		}
		return m.unreachable(ctx)
	}
	found, err := parseLs(out)
	if err != nil {
		return Snapshot{}, err
	}

	counts := map[string][2]int{} // name -> running, total
	for _, d := range found {
		psOut, err := m.docker.Run(ctx, "compose", "-p", d.Name, "ps", "-a", "--format", "json")
		if err != nil {
			return Snapshot{}, fmt.Errorf("compose: ps %s: %w", d.Name, err)
		}
		running, total, err := containerCounts(psOut)
		if err != nil {
			return Snapshot{}, err
		}
		counts[d.Name] = [2]int{running, total}
		if err := m.store.upsertDiscovered(ctx, d); err != nil {
			return Snapshot{}, err
		}
	}

	projects, err := m.store.list(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	for i := range projects {
		c := counts[projects[i].Name]
		projects[i].Running, projects[i].Total = c[0], c[1]
		projects[i].Status = statusOf(c[0], c[1])
	}
	return Snapshot{DockerReachable: true, Projects: projects}, nil
}

// unreachable keeps the last known Projects (or the saved ones, as Rimosso,
// when nothing is known yet) and flags Docker as not reachable.
func (m *Manager) unreachable(ctx context.Context) (Snapshot, error) {
	prev := m.Snapshot()
	if len(prev.Projects) > 0 {
		return Snapshot{DockerReachable: false, Projects: prev.Projects}, nil
	}
	projects, err := m.store.list(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	for i := range projects {
		projects[i].Status = StatusRemoved
	}
	return Snapshot{DockerReachable: false, Projects: projects}, nil
}

func (m *Manager) publish(next Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if reflect.DeepEqual(m.snap, next) {
		return
	}
	m.snap = next
	for _, w := range m.watchers {
		select { // keep only the latest snapshot in the 1-slot buffer
		case <-w:
		default:
		}
		w <- next
	}
}

// refreshEvery is the interval of the full refresh done by Run.
const refreshEvery = 30 * time.Second

// Run keeps the snapshot up to date until ctx is done: an immediate refresh,
// then a full one every 30 seconds. Refresh errors are passed to onError when
// it is not nil.
func (m *Manager) Run(ctx context.Context, onError func(error)) {
	tick := time.NewTicker(refreshEvery)
	defer tick.Stop()
	for {
		if err := m.Refresh(ctx); err != nil && ctx.Err() == nil && onError != nil {
			onError(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}
