# Skills MCP Server

An MCP (Model Context Protocol) server that exposes Claude-compatible skills as tools, enabling any MCP client to use them.

## Installation

```bash
go install github.com/portertech/skills/cmd/skills@latest
```

Or build from source:

```bash
git clone https://github.com/portertech/skills.git
cd skills
go build -o skills ./cmd/skills
```

### Docker

```bash
docker pull ghcr.io/portertech/skills:latest
```

Or build locally:

```bash
docker build -t skills:latest .
```

## Usage

```bash
# Start the server with default skills directory (~/.skills)
skills

# Start with a custom skills directory
skills /path/to/skills

# List discovered skills
skills --list /path/to/skills

# Enable verbose logging
skills --verbose /path/to/skills
```

### Docker

```bash
# Run with skills directory mounted
docker run -i -v ~/.skills:/skills:ro skills:latest /skills

# List skills
docker run -v ~/.skills:/skills:ro skills:latest --list /skills
```

## MCP Client Configuration

### Claude Code / Cursor / etc.

Add to your MCP configuration:

```json
{
  "mcpServers": {
    "skills": {
      "command": "skills",
      "args": ["/path/to/your/skills"]
    }
  }
}
```

### Docker

```json
{
  "mcpServers": {
    "skills": {
      "command": "docker",
      "args": ["run", "-i", "-v", "~/.skills:/skills:ro", "skills:latest", "/skills"]
    }
  }
}
```

## Creating Skills

Skills are directories containing a `SKILL.md` file with YAML frontmatter:

```
~/.skills/
├── code-review/
│   └── SKILL.md
└── git-workflow/
    └── SKILL.md
```

### SKILL.md Format

```markdown
---
name: my-skill
description: A brief description of what this skill does
license: MIT
allowed_tools:
  - view
  - grep
  - bash
---

# My Skill

Detailed instructions for the AI to follow when using this skill.

## Guidelines

- Specific rules and best practices
- Step-by-step procedures
- Examples and templates
```

### Required Fields

- `name`: Unique skill identifier
- `description`: Brief description shown in tool listings

### Optional Fields

- `license`: Skill license (e.g., MIT, Apache-2.0)
- `allowed_tools`: List of tools the skill may use

## How It Works

1. **Discovery**: The server scans the skills directory for `SKILL.md` files
2. **Registration**: Each skill becomes an MCP tool (e.g., `use_skill_code_review`)
3. **Invocation**: When a model calls the tool, it receives the skill's instructions
4. **Execution**: The model follows the instructions to complete the task

## Example Interaction

**User**: "Review my pull request"

**Model**: *Invokes `use_skill_code_review`*

**Server returns**:
```
# Skill: code-review

**Description:** Expert code review guidance...

---

## Instructions

When reviewing code, follow these guidelines...
```

**Model**: Uses the instructions to perform a thorough code review.

## Development

```bash
# Run tests
go test ./...

# Build
go build -o skills ./cmd/skills

# Test with sample skills
./skills --list ./testdata/skills
```

## License

MIT
