package bot

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// ProcessInfo represents a registered zen process.
type ProcessInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Alias       string    `json:"alias,omitempty"`
	Path        string    `json:"path"`
	PID         int       `json:"pid"`
	SocketPath  string    `json:"socket_path"`
	Status      string    `json:"status"`
	CurrentTask string    `json:"current_task,omitempty"`
	StartTime   time.Time `json:"start_time"`
	LastSeen    time.Time `json:"last_seen"`
	conn        net.Conn
}

// Registry manages registered zen processes.
type Registry struct {
	mu        sync.RWMutex
	processes map[string]*ProcessInfo // ID -> ProcessInfo
	byName    map[string]string       // name -> ID
	byAlias   map[string]string       // alias -> ID
	byPath    map[string]string       // path -> ID
	aliases   map[string]string       // user-defined alias -> path (from config)
}

// NewRegistry creates a new process registry.
func NewRegistry(aliases map[string]string) *Registry {
	if aliases == nil {
		aliases = make(map[string]string)
	}
	return &Registry{
		processes: make(map[string]*ProcessInfo),
		byName:    make(map[string]string),
		byAlias:   make(map[string]string),
		byPath:    make(map[string]string),
		aliases:   aliases,
	}
}

// Register adds a new process to the registry.
func (r *Registry) Register(info *ProcessInfo, conn net.Conn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate unique name
	info.Name = r.generateName(info.Path)
	info.LastSeen = time.Now()
	info.conn = conn

	// Check for user-defined alias
	for alias, path := range r.aliases {
		if path == info.Path {
			info.Alias = alias
			break
		}
	}

	r.processes[info.ID] = info
	r.byName[info.Name] = info.ID
	r.byPath[info.Path] = info.ID
	if info.Alias != "" {
		r.byAlias[info.Alias] = info.ID
	}

	return nil
}

// Unregister removes a process from the registry.
func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, ok := r.processes[id]
	if !ok {
		return
	}

	delete(r.byName, info.Name)
	delete(r.byPath, info.Path)
	if info.Alias != "" {
		delete(r.byAlias, info.Alias)
	}
	delete(r.processes, id)

	if info.conn != nil {
		info.conn.Close()
	}
}

// Get returns a process by ID.
func (r *Registry) Get(id string) *ProcessInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processes[id]
}

// Find finds a process by alias, name, or path.
func (r *Registry) Find(identifier string) *ProcessInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try alias first
	if id, ok := r.byAlias[identifier]; ok {
		return r.processes[id]
	}
	// Try name
	if id, ok := r.byName[identifier]; ok {
		return r.processes[id]
	}
	// Try path
	if id, ok := r.byPath[identifier]; ok {
		return r.processes[id]
	}
	// Try ID directly
	if info, ok := r.processes[identifier]; ok {
		return info
	}
	return nil
}

// List returns all registered processes.
func (r *Registry) List() []*ProcessInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*ProcessInfo, 0, len(r.processes))
	for _, info := range r.processes {
		list = append(list, info)
	}
	return list
}

// Count returns the number of registered processes.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.processes)
}

// UpdateStatus updates a process status.
func (r *Registry) UpdateStatus(id, status, task string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info, ok := r.processes[id]; ok {
		info.Status = status
		info.CurrentTask = task
		info.LastSeen = time.Now()
	}
}

// GetConnection returns the connection for a process.
func (r *Registry) GetConnection(id string) net.Conn {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if info, ok := r.processes[id]; ok {
		return info.conn
	}
	return nil
}

// generateName generates a unique name for a process based on its path.
func (r *Registry) generateName(path string) string {
	// Extract directory name
	dirName := extractDirName(path)

	// Check for duplicates
	count := 0
	for _, info := range r.processes {
		existingDir := extractDirName(info.Path)
		if existingDir == dirName {
			count++
		}
	}

	if count == 0 {
		return dirName
	}
	return fmt.Sprintf("%s#%d", dirName, count+1)
}

// extractDirName extracts the last component of a path.
func extractDirName(path string) string {
	if path == "" {
		return "unknown"
	}
	// Remove trailing slash
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	// Find last slash
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// CleanupStale removes processes that haven't been seen recently.
func (r *Registry) CleanupStale(timeout time.Duration) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var removed []string
	now := time.Now()
	for id, info := range r.processes {
		if now.Sub(info.LastSeen) > timeout {
			delete(r.byName, info.Name)
			delete(r.byPath, info.Path)
			if info.Alias != "" {
				delete(r.byAlias, info.Alias)
			}
			delete(r.processes, id)
			if info.conn != nil {
				info.conn.Close()
			}
			removed = append(removed, info.Name)
		}
	}
	return removed
}

// SetAlias sets a user-defined alias for a path.
func (r *Registry) SetAlias(alias, path string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.aliases[alias] = path

	// Update existing process if registered
	if id, ok := r.byPath[path]; ok {
		if info, ok := r.processes[id]; ok {
			// Remove old alias mapping
			if info.Alias != "" {
				delete(r.byAlias, info.Alias)
			}
			info.Alias = alias
			r.byAlias[alias] = id
		}
	}
}

// ToJSON serializes process info to JSON.
func (p *ProcessInfo) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
