package registry

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/portertech/skills/pkg/skill"
)

const skillFileName = "SKILL.md"

// Registry manages skill discovery and retrieval.
type Registry struct {
	root     string
	skills   map[string]*skill.Skill
	toolName map[string]string // maps tool name -> skill name for collision detection
	mu       sync.RWMutex
	logger   *slog.Logger
}

// NewRegistry creates a new skill registry rooted at the given directory.
func NewRegistry(root string, logger *slog.Logger) *Registry {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return &Registry{
		root:     root,
		skills:   make(map[string]*skill.Skill),
		toolName: make(map[string]string),
		logger:   logger,
	}
}

// Scan discovers all skills in the registry root directory.
func (r *Registry) Scan() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.skills = make(map[string]*skill.Skill)
	r.toolName = make(map[string]string)

	return filepath.WalkDir(r.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			r.logger.Warn("walk error", "path", path, "error", err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() != skillFileName {
			return nil
		}

		s, err := ParseSkillMD(path)
		if err != nil {
			r.logger.Warn("parse skill", "path", path, "error", err)
			return nil
		}

		s.Path = filepath.Dir(path)

		if existing, ok := r.skills[s.Name]; ok {
			r.logger.Warn("duplicate skill name",
				"name", s.Name,
				"path", path,
				"existing", existing.Path,
			)
			return nil
		}

		// Check for tool name collision after normalization
		toolName := ToolNameForSkill(s.Name)
		if existingName, ok := r.toolName[toolName]; ok {
			r.logger.Warn("tool name collision",
				"tool_name", toolName,
				"skill", s.Name,
				"existing_skill", existingName,
			)
			return nil
		}
		r.toolName[toolName] = s.Name

		r.skills[s.Name] = s
		r.logger.Debug("discovered skill", "name", s.Name, "path", s.Path)

		return nil
	})
}

// Get retrieves a skill by name.
func (r *Registry) Get(name string) *skill.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.skills[name]
}

// List returns all discovered skills sorted by name.
func (r *Registry) List() []*skill.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*skill.Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills
}

// Root returns the registry root directory.
func (r *Registry) Root() string {
	return r.root
}

// Count returns the number of discovered skills.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}

// String returns a human-readable summary.
func (r *Registry) String() string {
	return fmt.Sprintf("Registry{root=%s, skills=%d}", r.root, r.Count())
}

// ToolNameForSkill converts a skill name to a valid MCP tool name.
// Lowercases the name and replaces spaces and hyphens with underscores.
func ToolNameForSkill(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}
