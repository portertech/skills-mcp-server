package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillMD(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantLicense string
		wantTools   []string
		wantInstr   string
		wantErr     bool
	}{
		{
			name: "valid skill",
			content: `---
name: test-skill
description: A test skill
license: MIT
allowed_tools:
  - view
  - grep
---

# Test Skill

These are the instructions.
`,
			wantName:    "test-skill",
			wantDesc:    "A test skill",
			wantLicense: "MIT",
			wantTools:   []string{"view", "grep"},
			wantInstr:   "# Test Skill\n\nThese are the instructions.",
		},
		{
			name: "minimal skill",
			content: `---
name: minimal
description: Minimal skill
---

Instructions here.
`,
			wantName:  "minimal",
			wantDesc:  "Minimal skill",
			wantInstr: "Instructions here.",
		},
		{
			name: "no frontmatter",
			content: `# Just Markdown

No YAML frontmatter here.
`,
			wantErr: true,
		},
		{
			name: "missing name",
			content: `---
description: Missing name
---

Content.
`,
			wantErr: true,
		},
		{
			name: "missing description",
			content: `---
name: no-desc
---

Content.
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			skillPath := filepath.Join(tmpDir, "SKILL.md")

			if err := os.WriteFile(skillPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			skill, err := ParseSkillMD(skillPath)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if skill.Name != tt.wantName {
				t.Errorf("name = %q, want %q", skill.Name, tt.wantName)
			}

			if skill.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", skill.Description, tt.wantDesc)
			}

			if skill.License != tt.wantLicense {
				t.Errorf("license = %q, want %q", skill.License, tt.wantLicense)
			}

			if len(skill.AllowedTools) != len(tt.wantTools) {
				t.Errorf("allowed_tools len = %d, want %d", len(skill.AllowedTools), len(tt.wantTools))
			} else {
				for i, tool := range skill.AllowedTools {
					if tool != tt.wantTools[i] {
						t.Errorf("allowed_tools[%d] = %q, want %q", i, tool, tt.wantTools[i])
					}
				}
			}

			if skill.Instructions != tt.wantInstr {
				t.Errorf("instructions = %q, want %q", skill.Instructions, tt.wantInstr)
			}
		})
	}
}
