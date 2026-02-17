package sync

import (
	"encoding/json"
	"testing"
	"time"
)

func entity(t time.Time, val string) *SyncEntity {
	raw, _ := json.Marshal(map[string]string{"value": val})
	return &SyncEntity{ModifiedAt: t, Config: raw}
}

func scalar(t time.Time, val string) *SyncScalar {
	return &SyncScalar{ModifiedAt: t, Value: val}
}

func TestMergeNewerWins(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	local := NewSyncPayload("dev1")
	local.Providers["p1"] = entity(t1, "old")

	remote := NewSyncPayload("dev2")
	remote.Providers["p1"] = entity(t2, "new")

	merged := Merge(local, remote)

	var got map[string]string
	json.Unmarshal(merged.Providers["p1"].Config, &got)
	if got["value"] != "new" {
		t.Fatalf("expected newer remote to win, got %v", got)
	}
}

func TestMergeLocalNewerWins(t *testing.T) {
	t1 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	local := NewSyncPayload("dev1")
	local.Providers["p1"] = entity(t1, "local")

	remote := NewSyncPayload("dev2")
	remote.Providers["p1"] = entity(t2, "remote")

	merged := Merge(local, remote)

	var got map[string]string
	json.Unmarshal(merged.Providers["p1"].Config, &got)
	if got["value"] != "local" {
		t.Fatalf("expected newer local to win, got %v", got)
	}
}

func TestMergeOnlyOnOneSide(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	local := NewSyncPayload("dev1")
	local.Providers["local-only"] = entity(t1, "a")

	remote := NewSyncPayload("dev2")
	remote.Providers["remote-only"] = entity(t1, "b")

	merged := Merge(local, remote)

	if _, ok := merged.Providers["local-only"]; !ok {
		t.Fatal("local-only provider should be kept")
	}
	if _, ok := merged.Providers["remote-only"]; !ok {
		t.Fatal("remote-only provider should be kept")
	}
}

func TestMergeTombstoneWins(t *testing.T) {
	now := time.Now().UTC()
	t1 := now.Add(-2 * time.Hour)
	t2 := now.Add(-1 * time.Hour)

	local := NewSyncPayload("dev1")
	local.Providers["p1"] = entity(t1, "alive")

	remote := NewSyncPayload("dev2")
	remote.Tombstones["provider:p1"] = &Tombstone{DeletedAt: t2}

	merged := Merge(local, remote)

	if _, ok := merged.Providers["p1"]; ok {
		t.Fatal("tombstone should have deleted provider p1")
	}
	if _, ok := merged.Tombstones["provider:p1"]; !ok {
		t.Fatal("tombstone should be preserved")
	}
}

func TestMergeEntityNewerThanTombstone(t *testing.T) {
	now := time.Now().UTC()
	t1 := now.Add(-2 * time.Hour)
	t2 := now.Add(-1 * time.Hour)

	local := NewSyncPayload("dev1")
	local.Tombstones["provider:p1"] = &Tombstone{DeletedAt: t1}

	remote := NewSyncPayload("dev2")
	remote.Providers["p1"] = entity(t2, "recreated")

	merged := Merge(local, remote)

	if _, ok := merged.Providers["p1"]; !ok {
		t.Fatal("entity newer than tombstone should survive")
	}
	if _, ok := merged.Tombstones["provider:p1"]; ok {
		t.Fatal("tombstone should be removed when entity is newer")
	}
}

func TestMergeScalars(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	local := NewSyncPayload("dev1")
	local.DefaultProfile = scalar(t1, "old-profile")
	local.DefaultClient = scalar(t2, "claude")

	remote := NewSyncPayload("dev2")
	remote.DefaultProfile = scalar(t2, "new-profile")
	remote.DefaultClient = scalar(t1, "codex")

	merged := Merge(local, remote)

	if merged.DefaultProfile.Value != "new-profile" {
		t.Fatalf("expected newer remote default_profile, got %s", merged.DefaultProfile.Value)
	}
	if merged.DefaultClient.Value != "claude" {
		t.Fatalf("expected newer local default_client, got %s", merged.DefaultClient.Value)
	}
}

func TestMergeNilPayloads(t *testing.T) {
	p := NewSyncPayload("dev1")
	if Merge(p, nil) != p {
		t.Fatal("merge with nil remote should return local")
	}
	if Merge(nil, p) != p {
		t.Fatal("merge with nil local should return remote")
	}
}

func TestMergeProfiles(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	local := NewSyncPayload("dev1")
	local.Profiles["default"] = entity(t1, "local-profile")

	remote := NewSyncPayload("dev2")
	remote.Profiles["default"] = entity(t2, "remote-profile")

	merged := Merge(local, remote)

	var got map[string]string
	json.Unmarshal(merged.Profiles["default"].Config, &got)
	if got["value"] != "remote-profile" {
		t.Fatalf("expected newer remote profile, got %v", got)
	}
}
