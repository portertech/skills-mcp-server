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
docker pull portertech/skills-mcp-server:latest
```

Or build locally:

```bash
docker build -t skills-mcp-server:latest .
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
docker run -i --mount type=bind,src=$HOME/.skills,dst=/skills,readonly skills-mcp-server:latest /skills

# List skills
docker run --mount type=bind,src=$HOME/.skills,dst=/skills,readonly skills-mcp-server:latest --list /skills
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
      "args": ["run", "-i", "--rm", "--mount", "type=bind,src=/path/to/skills,dst=/skills,readonly", "skills-mcp-server:latest", "/skills"]
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

## How It Works

1. **Discovery**: The server scans the skills directory for `SKILL.md` files
2. **Registration**: Each skill becomes an MCP tool named after the skill
3. **Invocation**: When a model calls the tool, it receives the skill's instructions
4. **Execution**: The model follows the instructions to complete the task

### Tool Naming

Skill names are converted to valid MCP tool names:

| Skill Name | Tool Name |
|------------|----------|
| `code-review` | `code_review` |
| `My Skill` | `my_skill` |
| `git-workflow` | `git_workflow` |

The conversion:
- Lowercases the name
- Replaces spaces and hyphens with underscores

**Note**: Skills that would produce the same tool name (e.g., `code-review` and `code_review`) are considered duplicates. The first one discovered is registered; subsequent collisions are skipped with a warning.

## Example Interaction

**User**: "Review my pull request"

**Model**: *Invokes `code_review` tool*

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
make test

# Run tests with coverage
make test-cover

# Build
make build

# Format code
make fmt

# Run all checks
make ci

# Test with sample skills
make list-test
```

See `Makefile` for all available targets.

## License

MIT
