---
name: git-workflow
description: Best practices for Git workflows and commit hygiene
---

# Git Workflow Skill

Guidelines for maintaining clean Git history and following best practices.

## Commit Messages

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation only
- **style**: Formatting, missing semicolons, etc.
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **test**: Adding or updating tests
- **chore**: Maintenance tasks

### Guidelines
- Use imperative mood ("Add feature" not "Added feature")
- Keep subject line under 50 characters
- Wrap body at 72 characters
- Explain what and why, not how

## Branching Strategy

- `main` - Production-ready code
- `develop` - Integration branch
- `feature/*` - New features
- `fix/*` - Bug fixes
- `release/*` - Release preparation

## Pull Request Best Practices

1. Keep PRs small and focused
2. Write descriptive titles and descriptions
3. Link related issues
4. Request appropriate reviewers
5. Respond to feedback promptly
