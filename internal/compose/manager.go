package compose

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
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
	// OnActivity, when not nil, receives the events of every Activity (see
	// Do). Calls are serialised; it must not block for long.
	OnActivity func(ActivityEvent)
	// RefreshEvery is the interval of the safety-net full refresh done by Run;
	// 30 seconds when zero.
	RefreshEvery time.Duration
	// RetryEvery is how often Run retries while Docker is unreachable or its
	// events stream is down; 5 seconds when zero.
	RetryEvery time.Duration
}

// Manager is the interface of the compose module.
type Manager struct {
	docker Docker
	store  store

	refreshEvery, retryEvery time.Duration

	refreshMu sync.Mutex              // serialises Refresh
	profiles  map[string]profileCache // available Profiles per Project; guarded by refreshMu

	onActivity func(ActivityEvent)
	emitMu     sync.Mutex // serialises onActivity calls

	mu       sync.Mutex // guards the fields below
	base     Snapshot   // what Docker and the store say
	snap     Snapshot   // base plus the running Activities
	running  map[string]ActivityRef
	removing map[string]bool // Projects being removed from the list: no Activity may start on them
	nextID   uint64
	watchers []chan Snapshot
}

// New creates a Manager. The snapshot is empty until the first Refresh.
func New(cfg Config) (*Manager, error) {
	if cfg.Docker == nil || cfg.DB == nil {
		return nil, errors.New("compose: Docker and DB are required")
	}
	return &Manager{
		docker:     cfg.Docker,
		store:      store{db: cfg.DB},
		base:       Snapshot{Loading: true, Projects: []Project{}, Groups: []string{}},
		snap:       Snapshot{Loading: true, Projects: []Project{}, Groups: []string{}},
		running:    map[string]ActivityRef{},
		removing:   map[string]bool{},
		profiles:   map[string]profileCache{},
		onActivity: cfg.OnActivity,

		refreshEvery: orDefault(cfg.RefreshEvery, defaultRefreshEvery),
		retryEvery:   orDefault(cfg.RetryEvery, defaultRetryEvery),
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

	type observed struct {
		running, total int
		services       []Service
	}
	seen := map[string]observed{}
	for _, d := range found {
		psOut, err := m.docker.Run(ctx, "compose", "-p", d.Name, "ps", "-a", "--format", "json")
		if err != nil {
			return Snapshot{}, fmt.Errorf("compose: ps %s: %w", d.Name, err)
		}
		cs, err := parsePs(psOut)
		if err != nil {
			return Snapshot{}, err
		}
		seen[d.Name] = observed{running: runningCount(cs), total: len(cs), services: groupByService(cs)}
		if err := m.store.upsertDiscovered(ctx, d); err != nil {
			return Snapshot{}, err
		}
	}

	projects, err := m.store.list(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	for i := range projects {
		o := seen[projects[i].Name]
		projects[i].Running, projects[i].Total = o.running, o.total
		projects[i].Status = statusOf(o.running, o.total)
		projects[i].Services = o.services
		if projects[i].Services == nil {
			projects[i].Services = []Service{}
		}
		projects[i].ComposeMissing = composeMissing(projects[i].ComposeFiles)
		projects[i].Profiles = m.availableProfiles(ctx, projects[i])
	}
	return Snapshot{DockerReachable: true, Projects: projects}, nil
}

// composeMissing reports whether any of the compose files is gone. A file that
// cannot be inspected for another reason (permissions) is not reported missing.
func composeMissing(files []string) bool {
	for _, f := range files {
		if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
			return true
		}
	}
	return false
}

// unreachable keeps the last known Projects (or the saved ones, as Rimosso,
// when nothing is known yet) and flags Docker as not reachable.
func (m *Manager) unreachable(ctx context.Context) (Snapshot, error) {
	m.mu.Lock()
	prev := m.base
	m.mu.Unlock()
	if len(prev.Projects) > 0 {
		return Snapshot{DockerReachable: false, Projects: prev.Projects}, nil
	}
	projects, err := m.store.list(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	for i := range projects {
		projects[i].Status = StatusRemoved
		projects[i].Services = []Service{}
		projects[i].ComposeMissing = composeMissing(projects[i].ComposeFiles)
		projects[i].Profiles = orEmpty(m.profiles[projects[i].Name].names)
	}
	return Snapshot{DockerReachable: false, Projects: projects}, nil
}

// publish stores a new base snapshot and notifies watchers if the result
// changed.
func (m *Manager) publish(next Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.base = next
	m.recompose()
}

// recompose rebuilds snap from base and the running Activities and notifies
// watchers if it changed. m.mu must be held.
func (m *Manager) recompose() {
	next := Snapshot{Loading: m.base.Loading, DockerReachable: m.base.DockerReachable, Projects: make([]Project, len(m.base.Projects))}
	copy(next.Projects, m.base.Projects)
	next.Groups = groupsOf(next.Projects)
	for i := range next.Projects {
		if a, ok := m.running[next.Projects[i].Name]; ok {
			next.Projects[i].Activity = &a
		}
	}
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
