package registry

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/portertech/skills/pkg/skill"
)

const skillFileName = "SKILL.md"

// Registry manages skill discovery and retrieval.
type Registry struct {
	root   string
	skills map[string]*skill.Skill
	mu     sync.RWMutex
	logger *slog.Logger
}

// NewRegistry creates a new skill registry rooted at the given directory.
func NewRegistry(root string, logger *slog.Logger) *Registry {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return &Registry{
		root:   root,
		skills: make(map[string]*skill.Skill),
		logger: logger,
	}
}

// Scan discovers all skills in the registry root directory.
func (r *Registry) Scan() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.skills = make(map[string]*skill.Skill)

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

// List returns all discovered skills.
func (r *Registry) List() []*skill.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*skill.Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
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
