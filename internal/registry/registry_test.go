package registry

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	skill1Dir := filepath.Join(tmpDir, "skill1")
	skill2Dir := filepath.Join(tmpDir, "nested", "skill2")

	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("failed to create skill1 dir: %v", err)
	}
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatalf("failed to create skill2 dir: %v", err)
	}

	skill1Content := `---
name: skill-one
description: First test skill
---

Instructions for skill one.
`
	skill2Content := `---
name: skill-two
description: Second test skill
---

Instructions for skill two.
`

	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644); err != nil {
		t.Fatalf("failed to write skill1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644); err != nil {
		t.Fatalf("failed to write skill2: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := NewRegistry(tmpDir, logger)

	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if reg.Count() != 2 {
		t.Errorf("Count() = %d, want 2", reg.Count())
	}

	skill1 := reg.Get("skill-one")
	if skill1 == nil {
		t.Error("Get(skill-one) returned nil")
	} else {
		if skill1.Name != "skill-one" {
			t.Errorf("skill1.Name = %q, want %q", skill1.Name, "skill-one")
		}
		if skill1.Path != skill1Dir {
			t.Errorf("skill1.Path = %q, want %q", skill1.Path, skill1Dir)
		}
	}

	skill2 := reg.Get("skill-two")
	if skill2 == nil {
		t.Error("Get(skill-two) returned nil")
	} else {
		if skill2.Name != "skill-two" {
			t.Errorf("skill2.Name = %q, want %q", skill2.Name, "skill-two")
		}
	}

	nonexistent := reg.Get("nonexistent")
	if nonexistent != nil {
		t.Error("Get(nonexistent) should return nil")
	}

	skills := reg.List()
	if len(skills) != 2 {
		t.Errorf("List() len = %d, want 2", len(skills))
	}
}

func TestRegistryEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := NewRegistry(tmpDir, logger)

	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if reg.Count() != 0 {
		t.Errorf("Count() = %d, want 0", reg.Count())
	}
}

func TestRegistryDuplicateNames(t *testing.T) {
	tmpDir := t.TempDir()

	skill1Dir := filepath.Join(tmpDir, "dir1")
	skill2Dir := filepath.Join(tmpDir, "dir2")

	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("failed to create dir1: %v", err)
	}
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatalf("failed to create dir2: %v", err)
	}

	content := `---
name: duplicate-name
description: Same name skill
---

Instructions.
`

	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write skill1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write skill2: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := NewRegistry(tmpDir, logger)

	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if reg.Count() != 1 {
		t.Errorf("Count() = %d, want 1 (duplicate should be skipped)", reg.Count())
	}
}
