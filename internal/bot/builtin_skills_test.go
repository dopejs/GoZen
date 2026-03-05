package bot

import (
	"testing"
)

func TestBuiltinSkillsExist(t *testing.T) {
	skills := BuiltinSkills()
	if len(skills) == 0 {
		t.Fatal("BuiltinSkills() returned empty list")
	}

	// All existing intents should be covered
	requiredIntents := map[Intent]bool{
		IntentControl:     false,
		IntentBind:        false,
		IntentApprove:     false,
		IntentSendTask:    false,
		IntentPersona:     false,
		IntentForget:      false,
		IntentQueryStatus: false,
		IntentQueryList:   false,
	}

	for _, s := range skills {
		if _, ok := requiredIntents[s.Intent]; ok {
			requiredIntents[s.Intent] = true
		}
	}

	for intent, found := range requiredIntents {
		if !found {
			t.Errorf("missing builtin skill for intent %q", intent)
		}
	}
}

func TestBuiltinSkillsValidate(t *testing.T) {
	skills := BuiltinSkills()
	for _, s := range skills {
		if err := s.Validate(); err != nil {
			t.Errorf("builtin skill %q failed validation: %v", s.Name, err)
		}
	}
}

func TestBuiltinSkillsAreBuiltin(t *testing.T) {
	skills := BuiltinSkills()
	for _, s := range skills {
		if !s.Builtin {
			t.Errorf("builtin skill %q has Builtin=false", s.Name)
		}
		if !s.Enabled {
			t.Errorf("builtin skill %q has Enabled=false", s.Name)
		}
	}
}

func TestBuiltinSkillsHaveMultiLangKeywords(t *testing.T) {
	skills := BuiltinSkills()
	for _, s := range skills {
		en := s.Keywords["en"]
		zh := s.Keywords["zh"]
		if len(en) == 0 {
			t.Errorf("builtin skill %q missing English keywords", s.Name)
		}
		if len(zh) == 0 {
			t.Errorf("builtin skill %q missing Chinese keywords", s.Name)
		}
	}
}

func TestBuiltinSkillsUniqueNames(t *testing.T) {
	skills := BuiltinSkills()
	seen := make(map[string]bool)
	for _, s := range skills {
		if seen[s.Name] {
			t.Errorf("duplicate builtin skill name: %q", s.Name)
		}
		seen[s.Name] = true
	}
}

func TestBuiltinSkillsUniquePriorities(t *testing.T) {
	skills := BuiltinSkills()
	seen := make(map[int]string)
	for _, s := range skills {
		if prev, ok := seen[s.Priority]; ok {
			t.Errorf("duplicate priority %d: %q and %q", s.Priority, prev, s.Name)
		}
		seen[s.Priority] = s.Name
	}
}
