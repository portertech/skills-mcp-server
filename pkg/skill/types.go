// Package skill defines the core types for Claude-compatible skills.
package skill

// Skill represents a Claude-compatible skill parsed from a SKILL.md file.
type Skill struct {
	// Name is the unique identifier for the skill (required).
	Name string `yaml:"name"`

	// Description explains what the skill does (required).
	Description string `yaml:"description"`

	// Instructions contains the markdown content after the YAML frontmatter.
	Instructions string `yaml:"-"`

	// Path is the filesystem path to the skill directory.
	Path string `yaml:"-"`
}
