package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

const (
	FormatVersion    = 1
	SyncMetaFile     = "sync_meta.json"
	TombstoneMaxAge  = 30 * 24 * time.Hour // 30 days
)

// SyncPayload is the remote JSON structure uploaded/downloaded from backends.
type SyncPayload struct {
	FormatVersion  int                                `json:"format_version"`
	UpdatedAt      time.Time                          `json:"updated_at"`
	DeviceID       string                             `json:"device_id"`
	Salt           string                             `json:"salt"`
	Providers      map[string]*SyncEntity             `json:"providers"`
	Profiles       map[string]*SyncEntity             `json:"profiles"`
	DefaultProfile *SyncScalar                        `json:"default_profile,omitempty"`
	DefaultClient  *SyncScalar                        `json:"default_client,omitempty"`
	Tombstones     map[string]*Tombstone              `json:"tombstones,omitempty"`
}

// SyncEntity wraps a config entity with a modification timestamp.
// Config holds the raw JSON so we can defer typed deserialization.
type SyncEntity struct {
	ModifiedAt time.Time        `json:"modified_at"`
	Config     json.RawMessage  `json:"config"`
}

// SyncScalar wraps a scalar value with a modification timestamp.
type SyncScalar struct {
	ModifiedAt time.Time `json:"modified_at"`
	Value      string    `json:"value"`
}

// Tombstone marks a deleted entity.
type Tombstone struct {
	DeletedAt time.Time `json:"deleted_at"`
}

// SyncMeta is local-only metadata stored at ~/.zen/sync_meta.json.
type SyncMeta struct {
	DeviceID       string                `json:"device_id"`
	LastPullAt     time.Time             `json:"last_pull_at,omitempty"`
	LastPushAt     time.Time             `json:"last_push_at,omitempty"`
	Providers      map[string]time.Time  `json:"providers"`
	Profiles       map[string]time.Time  `json:"profiles"`
	DefaultProfile time.Time             `json:"default_profile,omitempty"`
	DefaultClient  time.Time             `json:"default_client,omitempty"`
	Tombstones     map[string]*Tombstone `json:"tombstones,omitempty"`
}

// SyncStatus is returned by the status API.
type SyncStatus struct {
	Configured bool      `json:"configured"`
	Backend    string    `json:"backend,omitempty"`
	DeviceID   string    `json:"device_id,omitempty"`
	LastPullAt time.Time `json:"last_pull_at,omitempty"`
	LastPushAt time.Time `json:"last_push_at,omitempty"`
}

// NewSyncPayload creates an empty payload with initialized maps.
func NewSyncPayload(deviceID string) *SyncPayload {
	return &SyncPayload{
		FormatVersion: FormatVersion,
		UpdatedAt:     time.Now().UTC(),
		DeviceID:      deviceID,
		Providers:     make(map[string]*SyncEntity),
		Profiles:      make(map[string]*SyncEntity),
		Tombstones:    make(map[string]*Tombstone),
	}
}

// NewSyncMeta creates an empty meta with initialized maps.
func NewSyncMeta(deviceID string) *SyncMeta {
	return &SyncMeta{
		DeviceID:   deviceID,
		Providers:  make(map[string]time.Time),
		Profiles:   make(map[string]time.Time),
		Tombstones: make(map[string]*Tombstone),
	}
}

// LoadSyncMeta reads sync_meta.json from the config directory.
func LoadSyncMeta() (*SyncMeta, error) {
	path := filepath.Join(config.ConfigDirPath(), SyncMetaFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var meta SyncMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	if meta.Providers == nil {
		meta.Providers = make(map[string]time.Time)
	}
	if meta.Profiles == nil {
		meta.Profiles = make(map[string]time.Time)
	}
	if meta.Tombstones == nil {
		meta.Tombstones = make(map[string]*Tombstone)
	}
	return &meta, nil
}

// Save writes sync_meta.json to the config directory.
func (m *SyncMeta) Save() error {
	path := filepath.Join(config.ConfigDirPath(), SyncMetaFile)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}
