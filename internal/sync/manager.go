package sync

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// SyncManager orchestrates sync operations between local config and remote backend.
type SyncManager struct {
	mu        sync.Mutex
	backend   Backend
	cfg       *config.SyncConfig
	meta      *SyncMeta
	isPulling bool // guard to prevent push-after-pull loops
}

// NewSyncManager creates a SyncManager from the current sync config.
func NewSyncManager(cfg *config.SyncConfig) (*SyncManager, error) {
	if cfg == nil || cfg.Backend == "" {
		return nil, fmt.Errorf("sync not configured")
	}
	backend, err := NewBackend(cfg)
	if err != nil {
		return nil, err
	}

	meta, err := LoadSyncMeta()
	if err != nil {
		return nil, fmt.Errorf("load sync meta: %w", err)
	}
	if meta == nil {
		meta = NewSyncMeta(generateDeviceID())
		if err := meta.Save(); err != nil {
			return nil, fmt.Errorf("save sync meta: %w", err)
		}
	}

	return &SyncManager{
		backend: backend,
		cfg:     cfg,
		meta:    meta,
	}, nil
}

// IsPulling returns true if a pull is currently in progress.
func (m *SyncManager) IsPulling() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isPulling
}

// TestConnection tests connectivity to the remote backend.
func (m *SyncManager) TestConnection(ctx context.Context) error {
	_, err := m.backend.Download(ctx)
	return err
}

// Status returns the current sync status.
func (m *SyncManager) Status() *SyncStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &SyncStatus{
		Configured: true,
		Backend:    m.backend.Name(),
		DeviceID:   m.meta.DeviceID,
		LastPullAt: m.meta.LastPullAt,
		LastPushAt: m.meta.LastPushAt,
	}
}

// Pull downloads remote payload, merges with local, and applies changes.
func (m *SyncManager) Pull(ctx context.Context) error {
	m.mu.Lock()
	m.isPulling = true
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.isPulling = false
		m.mu.Unlock()
	}()

	// Download remote
	remoteData, err := m.backend.Download(ctx)
	if err != nil {
		return fmt.Errorf("pull download: %w", err)
	}
	if remoteData == nil {
		return nil // nothing remote yet
	}

	// Decrypt and parse remote payload
	remote, err := m.decryptPayload(remoteData)
	if err != nil {
		return fmt.Errorf("pull decrypt: %w", err)
	}

	// Build local payload from current config
	local, err := m.buildLocalPayload()
	if err != nil {
		return fmt.Errorf("pull build local: %w", err)
	}

	// Merge
	merged := Merge(local, remote)

	// Apply merged state to local config
	if err := m.applyToLocal(merged); err != nil {
		return fmt.Errorf("pull apply: %w", err)
	}

	// Update meta
	m.mu.Lock()
	m.meta.LastPullAt = time.Now().UTC()
	m.updateMetaTimestamps(merged)
	m.mu.Unlock()

	return m.meta.Save()
}

// Push builds local payload, merges with remote, encrypts, and uploads.
func (m *SyncManager) Push(ctx context.Context) error {
	m.mu.Lock()
	if m.isPulling {
		m.mu.Unlock()
		return nil // skip push triggered by pull's config save
	}
	m.mu.Unlock()

	// Build local payload
	local, err := m.buildLocalPayload()
	if err != nil {
		return fmt.Errorf("push build local: %w", err)
	}

	// Download remote for merge
	remoteData, err := m.backend.Download(ctx)
	if err != nil {
		return fmt.Errorf("push download: %w", err)
	}

	var merged *SyncPayload
	if remoteData != nil {
		remote, err := m.decryptPayload(remoteData)
		if err != nil {
			return fmt.Errorf("push decrypt: %w", err)
		}
		merged = Merge(local, remote)
	} else {
		merged = local
	}

	// Encrypt and upload
	data, err := m.encryptPayload(merged)
	if err != nil {
		return fmt.Errorf("push encrypt: %w", err)
	}
	if err := m.backend.Upload(ctx, data); err != nil {
		return fmt.Errorf("push upload: %w", err)
	}

	// Update meta
	m.mu.Lock()
	m.meta.LastPushAt = time.Now().UTC()
	m.updateMetaTimestamps(merged)
	m.mu.Unlock()

	return m.meta.Save()
}

// buildLocalPayload creates a SyncPayload from the current local config and meta.
func (m *SyncManager) buildLocalPayload() (*SyncPayload, error) {
	store := config.DefaultStore()
	providers := store.ProviderMap()
	profiles := store.ListProfiles()

	m.mu.Lock()
	payload := NewSyncPayload(m.meta.DeviceID)
	// Copy salt from config or meta
	payload.Salt = "" // will be set during encryption

	// Providers
	for name, pc := range providers {
		raw, err := json.Marshal(pc)
		if err != nil {
			m.mu.Unlock()
			return nil, err
		}
		modAt := m.meta.Providers[name]
		if modAt.IsZero() {
			modAt = time.Now().UTC()
			m.meta.Providers[name] = modAt
		}
		payload.Providers[name] = &SyncEntity{ModifiedAt: modAt, Config: raw}
	}

	// Profiles
	for _, name := range profiles {
		pc := store.GetProfileConfig(name)
		if pc == nil {
			continue
		}
		raw, err := json.Marshal(pc)
		if err != nil {
			m.mu.Unlock()
			return nil, err
		}
		modAt := m.meta.Profiles[name]
		if modAt.IsZero() {
			modAt = time.Now().UTC()
			m.meta.Profiles[name] = modAt
		}
		payload.Profiles[name] = &SyncEntity{ModifiedAt: modAt, Config: raw}
	}

	// Scalars
	dp := store.GetDefaultProfile()
	dpTime := m.meta.DefaultProfile
	if dpTime.IsZero() {
		dpTime = time.Now().UTC()
		m.meta.DefaultProfile = dpTime
	}
	payload.DefaultProfile = &SyncScalar{ModifiedAt: dpTime, Value: dp}

	// Tombstones
	for k, ts := range m.meta.Tombstones {
		payload.Tombstones[k] = ts
	}

	m.mu.Unlock()
	return payload, nil
}

// applyToLocal applies a merged payload to the local config store.
func (m *SyncManager) applyToLocal(payload *SyncPayload) error {
	store := config.DefaultStore()

	// Apply providers
	currentProviders := store.ProviderMap()
	for name, ent := range payload.Providers {
		var pc config.ProviderConfig
		if err := json.Unmarshal(ent.Config, &pc); err != nil {
			return fmt.Errorf("unmarshal provider %s: %w", name, err)
		}
		if err := store.SetProvider(name, &pc); err != nil {
			return err
		}
	}
	// Remove providers not in merged payload (deleted via tombstone)
	for name := range currentProviders {
		if _, ok := payload.Providers[name]; !ok {
			store.DeleteProvider(name)
		}
	}

	// Apply profiles
	currentProfiles := store.ListProfiles()
	for name, ent := range payload.Profiles {
		var pc config.ProfileConfig
		if err := json.Unmarshal(ent.Config, &pc); err != nil {
			return fmt.Errorf("unmarshal profile %s: %w", name, err)
		}
		if err := store.SetProfileConfig(name, &pc); err != nil {
			return err
		}
	}
	// Remove profiles not in merged payload
	for _, name := range currentProfiles {
		if _, ok := payload.Profiles[name]; !ok {
			store.DeleteProfile(name)
		}
	}

	// Apply scalars
	if payload.DefaultProfile != nil {
		store.SetDefaultProfile(payload.DefaultProfile.Value)
	}

	return nil
}

// updateMetaTimestamps updates local meta timestamps from a merged payload.
func (m *SyncManager) updateMetaTimestamps(payload *SyncPayload) {
	for name, ent := range payload.Providers {
		m.meta.Providers[name] = ent.ModifiedAt
	}
	// Clean up meta for providers no longer present
	for name := range m.meta.Providers {
		if _, ok := payload.Providers[name]; !ok {
			delete(m.meta.Providers, name)
		}
	}
	for name, ent := range payload.Profiles {
		m.meta.Profiles[name] = ent.ModifiedAt
	}
	for name := range m.meta.Profiles {
		if _, ok := payload.Profiles[name]; !ok {
			delete(m.meta.Profiles, name)
		}
	}
	if payload.DefaultProfile != nil {
		m.meta.DefaultProfile = payload.DefaultProfile.ModifiedAt
	}
	m.meta.Tombstones = payload.Tombstones
}

// encryptPayload encrypts provider tokens in the payload and marshals to JSON.
func (m *SyncManager) encryptPayload(payload *SyncPayload) ([]byte, error) {
	if m.cfg.Passphrase == "" {
		return json.MarshalIndent(payload, "", "  ")
	}

	// Generate or reuse salt
	var salt []byte
	if payload.Salt != "" {
		var err error
		salt, err = base64.StdEncoding.DecodeString(payload.Salt)
		if err != nil {
			salt = nil
		}
	}
	if salt == nil {
		var err error
		salt, err = GenerateSalt()
		if err != nil {
			return nil, err
		}
	}
	payload.Salt = base64.StdEncoding.EncodeToString(salt)
	key := DeriveKey(m.cfg.Passphrase, salt)

	// Encrypt each provider's config (contains auth tokens)
	for name, ent := range payload.Providers {
		encrypted, err := Encrypt(ent.Config, key)
		if err != nil {
			return nil, fmt.Errorf("encrypt provider %s: %w", name, err)
		}
		payload.Providers[name] = &SyncEntity{
			ModifiedAt: ent.ModifiedAt,
			Config:     json.RawMessage(fmt.Sprintf("%q", encrypted)),
		}
	}

	return json.MarshalIndent(payload, "", "  ")
}

// decryptPayload parses JSON and decrypts provider tokens.
func (m *SyncManager) decryptPayload(data []byte) (*SyncPayload, error) {
	var payload SyncPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if payload.Providers == nil {
		payload.Providers = make(map[string]*SyncEntity)
	}
	if payload.Profiles == nil {
		payload.Profiles = make(map[string]*SyncEntity)
	}
	if payload.Tombstones == nil {
		payload.Tombstones = make(map[string]*Tombstone)
	}

	if m.cfg.Passphrase == "" || payload.Salt == "" {
		return &payload, nil
	}

	salt, err := base64.StdEncoding.DecodeString(payload.Salt)
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}
	key := DeriveKey(m.cfg.Passphrase, salt)

	// Decrypt each provider's config
	for name, ent := range payload.Providers {
		// Config is stored as a JSON string (quoted encrypted base64)
		var encrypted string
		if err := json.Unmarshal(ent.Config, &encrypted); err != nil {
			// Not encrypted, skip
			continue
		}
		decrypted, err := Decrypt(encrypted, key)
		if err != nil {
			return nil, fmt.Errorf("decrypt provider %s: %w", name, err)
		}
		ent.Config = decrypted
	}

	return &payload, nil
}

// generateDeviceID creates a short random device identifier.
func generateDeviceID() string {
	b := make([]byte, 4)
	salt, _ := GenerateSalt()
	if len(salt) >= 4 {
		copy(b, salt[:4])
	}
	return fmt.Sprintf("%x", b)
}

// MarkDeleted records a tombstone for a deleted entity.
// prefix is "provider" or "profile", name is the entity name.
func (m *SyncManager) MarkDeleted(prefix, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := prefix + ":" + name
	m.meta.Tombstones[key] = &Tombstone{DeletedAt: time.Now().UTC()}
	// Remove from meta timestamps
	switch prefix {
	case "provider":
		delete(m.meta.Providers, name)
	case "profile":
		delete(m.meta.Profiles, name)
	}
	m.meta.Save()
}

// MarkModified updates the modification timestamp for an entity.
func (m *SyncManager) MarkModified(prefix, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	switch prefix {
	case "provider":
		m.meta.Providers[name] = now
	case "profile":
		m.meta.Profiles[name] = now
	case "default_profile":
		m.meta.DefaultProfile = now
	}
	m.meta.Save()
}
