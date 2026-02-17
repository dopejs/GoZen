package sync

import (
	"time"
)

// Merge merges local and remote payloads using per-entity timestamp resolution.
// Returns the merged payload that should be uploaded and applied locally.
func Merge(local, remote *SyncPayload) *SyncPayload {
	if remote == nil {
		return local
	}
	if local == nil {
		return remote
	}

	merged := &SyncPayload{
		FormatVersion: FormatVersion,
		UpdatedAt:     time.Now().UTC(),
		DeviceID:      local.DeviceID,
		Salt:          local.Salt,
		Providers:     make(map[string]*SyncEntity),
		Profiles:      make(map[string]*SyncEntity),
		Tombstones:    make(map[string]*Tombstone),
	}

	// Merge tombstones first (union, keep newest)
	mergeTombstones(merged.Tombstones, local.Tombstones)
	mergeTombstones(merged.Tombstones, remote.Tombstones)
	expireTombstones(merged.Tombstones)

	// Merge providers
	mergeEntities(merged.Providers, local.Providers, remote.Providers, "provider", merged.Tombstones)

	// Merge profiles
	mergeEntities(merged.Profiles, local.Profiles, remote.Profiles, "profile", merged.Tombstones)

	// Merge scalars
	merged.DefaultProfile = mergeScalar(local.DefaultProfile, remote.DefaultProfile)
	merged.DefaultClient = mergeScalar(local.DefaultClient, remote.DefaultClient)

	return merged
}

// mergeEntities merges two entity maps. Tombstones override live entities if newer.
func mergeEntities(dst, local, remote map[string]*SyncEntity, prefix string, tombstones map[string]*Tombstone) {
	// Collect all keys
	keys := make(map[string]struct{})
	for k := range local {
		keys[k] = struct{}{}
	}
	for k := range remote {
		keys[k] = struct{}{}
	}

	for name := range keys {
		tombKey := prefix + ":" + name
		l := local[name]
		r := remote[name]

		winner := pickNewer(l, r)
		if winner == nil {
			continue
		}

		// Check tombstone
		if ts, ok := tombstones[tombKey]; ok {
			if ts.DeletedAt.After(winner.ModifiedAt) || ts.DeletedAt.Equal(winner.ModifiedAt) {
				// Tombstone wins — entity stays deleted
				continue
			}
			// Entity is newer than tombstone — remove tombstone, keep entity
			delete(tombstones, tombKey)
		}

		dst[name] = winner
	}
}

// pickNewer returns the entity with the later ModifiedAt. If only one exists, returns it.
func pickNewer(a, b *SyncEntity) *SyncEntity {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if b.ModifiedAt.After(a.ModifiedAt) {
		return b
	}
	return a
}

// mergeScalar returns the scalar with the later ModifiedAt.
func mergeScalar(a, b *SyncScalar) *SyncScalar {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if b.ModifiedAt.After(a.ModifiedAt) {
		return b
	}
	return a
}

// mergeTombstones merges src tombstones into dst, keeping the newer DeletedAt.
func mergeTombstones(dst, src map[string]*Tombstone) {
	for k, ts := range src {
		if existing, ok := dst[k]; ok {
			if ts.DeletedAt.After(existing.DeletedAt) {
				dst[k] = ts
			}
		} else {
			dst[k] = ts
		}
	}
}

// expireTombstones removes tombstones older than TombstoneMaxAge.
func expireTombstones(tombstones map[string]*Tombstone) {
	cutoff := time.Now().UTC().Add(-TombstoneMaxAge)
	for k, ts := range tombstones {
		if ts.DeletedAt.Before(cutoff) {
			delete(tombstones, k)
		}
	}
}
