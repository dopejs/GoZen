package bot

import (
	"testing"
)

func TestNewSkill(t *testing.T) {
	s := &Skill{
		Name:        "process-control",
		Description: "控制进程的启停",
		Intent:      IntentControl,
		Priority:    10,
		Enabled:     true,
		Builtin:     true,
		Keywords: map[string][]string{
			"en": {"pause", "resume", "stop", "cancel"},
			"zh": {"暂停", "恢复", "停止", "取消"},
		},
	}

	if s.Name != "process-control" {
		t.Errorf("Name = %q, want %q", s.Name, "process-control")
	}
	if s.Intent != IntentControl {
		t.Errorf("Intent = %q, want %q", s.Intent, IntentControl)
	}
	if !s.Enabled {
		t.Error("Enabled should be true")
	}
	if !s.Builtin {
		t.Error("Builtin should be true")
	}
}

func TestSkillValidate(t *testing.T) {
	tests := []struct {
		name    string
		skill   Skill
		wantErr bool
	}{
		{
			name: "valid skill",
			skill: Skill{
				Name:     "test",
				Description: "test skill",
				Intent:   IntentControl,
				Priority: 10,
				Keywords: map[string][]string{"en": {"test"}},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			skill: Skill{
				Description: "test",
				Intent:      IntentControl,
				Priority:    10,
				Keywords:    map[string][]string{"en": {"test"}},
			},
			wantErr: true,
		},
		{
			name: "missing description",
			skill: Skill{
				Name:     "test",
				Intent:   IntentControl,
				Priority: 10,
				Keywords: map[string][]string{"en": {"test"}},
			},
			wantErr: true,
		},
		{
			name: "missing intent",
			skill: Skill{
				Name:        "test",
				Description: "test",
				Priority:    10,
				Keywords:    map[string][]string{"en": {"test"}},
			},
			wantErr: true,
		},
		{
			name: "priority too low",
			skill: Skill{
				Name:        "test",
				Description: "test",
				Intent:      IntentControl,
				Priority:    0,
				Keywords:    map[string][]string{"en": {"test"}},
			},
			wantErr: true,
		},
		{
			name: "priority too high",
			skill: Skill{
				Name:        "test",
				Description: "test",
				Intent:      IntentControl,
				Priority:    101,
				Keywords:    map[string][]string{"en": {"test"}},
			},
			wantErr: true,
		},
		{
			name: "no keywords",
			skill: Skill{
				Name:        "test",
				Description: "test",
				Intent:      IntentControl,
				Priority:    10,
				Keywords:    map[string][]string{},
			},
			wantErr: true,
		},
		{
			name: "empty keyword list",
			skill: Skill{
				Name:        "test",
				Description: "test",
				Intent:      IntentControl,
				Priority:    10,
				Keywords:    map[string][]string{"en": {}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.skill.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSkillRegistryRegisterAndGet(t *testing.T) {
	reg := NewSkillRegistry()

	s := &Skill{
		Name:        "test-skill",
		Description: "test",
		Intent:      IntentControl,
		Priority:    10,
		Enabled:     true,
		Keywords:    map[string][]string{"en": {"test"}},
	}

	if err := reg.Register(s); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got := reg.Get("test-skill")
	if got == nil {
		t.Fatal("Get() returned nil")
	}
	if got.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", got.Name, "test-skill")
	}
}

func TestSkillRegistryRegisterDuplicate(t *testing.T) {
	reg := NewSkillRegistry()

	s := &Skill{
		Name:        "dup",
		Description: "test",
		Intent:      IntentControl,
		Priority:    10,
		Enabled:     true,
		Keywords:    map[string][]string{"en": {"test"}},
	}

	if err := reg.Register(s); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := reg.Register(s); err == nil {
		t.Error("second Register() should return error for duplicate name")
	}
}

func TestSkillRegistryRegisterInvalid(t *testing.T) {
	reg := NewSkillRegistry()

	s := &Skill{Name: ""} // invalid: missing required fields
	if err := reg.Register(s); err == nil {
		t.Error("Register() should return error for invalid skill")
	}
}

func TestSkillRegistryGetMissing(t *testing.T) {
	reg := NewSkillRegistry()
	if got := reg.Get("nonexistent"); got != nil {
		t.Errorf("Get() = %v, want nil", got)
	}
}

func TestSkillRegistryList(t *testing.T) {
	reg := NewSkillRegistry()

	skills := []*Skill{
		{Name: "b-skill", Description: "b", Intent: IntentControl, Priority: 20, Enabled: true, Keywords: map[string][]string{"en": {"b"}}},
		{Name: "a-skill", Description: "a", Intent: IntentBind, Priority: 10, Enabled: true, Keywords: map[string][]string{"en": {"a"}}},
		{Name: "c-skill", Description: "c", Intent: IntentChat, Priority: 30, Enabled: true, Keywords: map[string][]string{"en": {"c"}}},
	}

	for _, s := range skills {
		if err := reg.Register(s); err != nil {
			t.Fatalf("Register(%s) error = %v", s.Name, err)
		}
	}

	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("List() len = %d, want 3", len(list))
	}

	// Should be sorted by priority (ascending)
	if list[0].Name != "a-skill" {
		t.Errorf("list[0].Name = %q, want %q", list[0].Name, "a-skill")
	}
	if list[1].Name != "b-skill" {
		t.Errorf("list[1].Name = %q, want %q", list[1].Name, "b-skill")
	}
	if list[2].Name != "c-skill" {
		t.Errorf("list[2].Name = %q, want %q", list[2].Name, "c-skill")
	}
}

func TestSkillRegistryEnableDisable(t *testing.T) {
	reg := NewSkillRegistry()

	s := &Skill{
		Name:        "toggle",
		Description: "test",
		Intent:      IntentControl,
		Priority:    10,
		Enabled:     true,
		Keywords:    map[string][]string{"en": {"test"}},
	}
	reg.Register(s)

	// Disable
	if err := reg.SetEnabled("toggle", false); err != nil {
		t.Fatalf("SetEnabled(false) error = %v", err)
	}
	got := reg.Get("toggle")
	if got.Enabled {
		t.Error("skill should be disabled")
	}

	// Enable
	if err := reg.SetEnabled("toggle", true); err != nil {
		t.Fatalf("SetEnabled(true) error = %v", err)
	}
	got = reg.Get("toggle")
	if !got.Enabled {
		t.Error("skill should be enabled")
	}

	// Missing skill
	if err := reg.SetEnabled("nonexistent", true); err == nil {
		t.Error("SetEnabled() should return error for missing skill")
	}
}

func TestSkillRegistryListEnabled(t *testing.T) {
	reg := NewSkillRegistry()

	reg.Register(&Skill{Name: "enabled1", Description: "e1", Intent: IntentControl, Priority: 10, Enabled: true, Keywords: map[string][]string{"en": {"e1"}}})
	reg.Register(&Skill{Name: "disabled1", Description: "d1", Intent: IntentBind, Priority: 20, Enabled: false, Keywords: map[string][]string{"en": {"d1"}}})
	reg.Register(&Skill{Name: "enabled2", Description: "e2", Intent: IntentChat, Priority: 30, Enabled: true, Keywords: map[string][]string{"en": {"e2"}}})

	list := reg.ListEnabled()
	if len(list) != 2 {
		t.Fatalf("ListEnabled() len = %d, want 2", len(list))
	}
	if list[0].Name != "enabled1" || list[1].Name != "enabled2" {
		t.Errorf("ListEnabled() = [%s, %s], want [enabled1, enabled2]", list[0].Name, list[1].Name)
	}
}
