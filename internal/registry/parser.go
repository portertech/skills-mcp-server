// Package registry handles skill discovery and management.
package registry

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/portertech/skills/pkg/skill"
	"gopkg.in/yaml.v3"
)

var (
	ErrNoFrontmatter = errors.New("no YAML frontmatter found")
	ErrMissingName   = errors.New("skill name is required")
	ErrMissingDesc   = errors.New("skill description is required")
)

// ParseSkillMD parses a SKILL.md file and returns a Skill.
// The file must contain YAML frontmatter between --- markers.
func ParseSkillMD(path string) (*skill.Skill, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open skill file: %w", err)
	}
	defer f.Close()

	var (
		scanner       = bufio.NewScanner(f)
		inFrontmatter bool
		frontmatter   strings.Builder
		content       strings.Builder
		lineNum       int
		fmStart       = -1
		fmEnd         = -1
	)

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if lineNum == 1 && line == "---" {
			inFrontmatter = true
			fmStart = lineNum
			continue
		}

		if inFrontmatter && line == "---" {
			inFrontmatter = false
			fmEnd = lineNum
			continue
		}

		if inFrontmatter {
			frontmatter.WriteString(line)
			frontmatter.WriteString("\n")
		} else if fmEnd > 0 {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read skill file: %w", err)
	}

	if fmStart < 0 || fmEnd < 0 {
		return nil, ErrNoFrontmatter
	}

	var s skill.Skill
	if err := yaml.Unmarshal([]byte(frontmatter.String()), &s); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if s.Name == "" {
		return nil, ErrMissingName
	}
	if s.Description == "" {
		return nil, ErrMissingDesc
	}

	s.Instructions = strings.TrimSpace(content.String())

	return &s, nil
}
