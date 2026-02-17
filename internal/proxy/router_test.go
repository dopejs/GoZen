package proxy

import (
	"testing"
)

func TestParseRoutePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantProf  string
		wantSess  string
		wantRem   string
		wantErr   bool
	}{
		{
			name:     "standard path",
			path:     "/default/f47ac10b/v1/messages",
			wantProf: "default",
			wantSess: "f47ac10b",
			wantRem:  "/v1/messages",
		},
		{
			name:     "work profile",
			path:     "/work/a1b2c3d4/v1/chat/completions",
			wantProf: "work",
			wantSess: "a1b2c3d4",
			wantRem:  "/v1/chat/completions",
		},
		{
			name:     "temp profile",
			path:     "/_tmp_8f3a2b/abc123/v1/messages",
			wantProf: "_tmp_8f3a2b",
			wantSess: "abc123",
			wantRem:  "/v1/messages",
		},
		{
			name:     "no remainder",
			path:     "/default/sess123",
			wantProf: "default",
			wantSess: "sess123",
			wantRem:  "",
		},
		{
			name:    "empty path",
			path:    "/",
			wantErr: true,
		},
		{
			name:    "single segment",
			path:    "/default",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ri, err := ParseRoutePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ri.Profile != tt.wantProf {
				t.Errorf("Profile = %q, want %q", ri.Profile, tt.wantProf)
			}
			if ri.SessionID != tt.wantSess {
				t.Errorf("SessionID = %q, want %q", ri.SessionID, tt.wantSess)
			}
			if ri.Remainder != tt.wantRem {
				t.Errorf("Remainder = %q, want %q", ri.Remainder, tt.wantRem)
			}
		})
	}
}

func TestRouteInfoCacheKey(t *testing.T) {
	ri := &RouteInfo{Profile: "work", SessionID: "abc123"}
	if got := ri.CacheKey(); got != "work:abc123" {
		t.Errorf("CacheKey() = %q, want %q", got, "work:abc123")
	}
}

func TestRouteInfoIsTempProfile(t *testing.T) {
	tests := []struct {
		profile string
		want    bool
	}{
		{"default", false},
		{"work", false},
		{"_tmp_8f3a2b", true},
		{"_tmp_", true},
		{"tmp_abc", false},
	}
	for _, tt := range tests {
		ri := &RouteInfo{Profile: tt.profile}
		if got := ri.IsTempProfile(); got != tt.want {
			t.Errorf("IsTempProfile(%q) = %v, want %v", tt.profile, got, tt.want)
		}
	}
}
