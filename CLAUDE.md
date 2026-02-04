# lazyas - Project Guidelines

## What lazyas IS

**lazyas is a MANAGEMENT TOOL for Agent Skills.**

It handles:
- **Browsing** skills from a registry
- **Installing** skills to the appropriate skills directory
- **Removing** installed skills
- **Updating** skills to newer versions
- **Searching** for skills by name, tags, or description
- **Tracking** installed skills and their versions via manifest

Think of it like `apt`, `brew`, or `npm` - a package manager for Agent Skills.

## Multi-Backend Support

**lazyas supports multiple AI agent backends through a symlinked central directory.**

All skills live in one place (`~/.lazyas/skills/`), with symlinks from each backend's expected location pointing back to it. This means every linked backend shares the same skills without duplication.

Built-in backends:
- Claude Code (`~/.claude/skills/`)
- OpenAI Codex (`~/.codex/skills/`)
- Gemini CLI (`~/.gemini/skills/`)
- Cursor (`~/.cursor/skills/`)
- GitHub Copilot (`~/.copilot/skills/`)
- Amp (`$XDG_CONFIG_HOME/agents/skills/`)
- Goose (`$XDG_CONFIG_HOME/goose/skills/`)
- OpenCode (`$XDG_CONFIG_HOME/opencode/skills/`)
- Mistral Vibe (`~/.vibe/skills/`)

Custom backends can be added via `lazyas backend add <name> <path>`.

## Scope: Agent Skills ONLY

**lazyas manages Agent Skills exclusively.**

It does NOT manage:
- MCP (Model Context Protocol) servers
- Other plugin/extension systems
- AI model configurations
- API keys or credentials

If someone asks about MCP servers or other integrations, those are separate concerns with their own tooling.

## What lazyas IS NOT

**lazyas is NOT a skill creation tool.**

Skill creation is a separate domain with its own tooling:
- Skill authoring: https://www.skillcreator.ai/
- Skill format documentation: Agent Skills specification

lazyas assumes skills already exist in a registry. It does not:
- Generate SKILL.md files
- Scaffold new skill projects
- Validate skill content beyond checking SKILL.md exists
- Provide skill authoring guidance

If someone asks about creating skills, point them to skillcreator.ai instead.

## Architecture Notes

- Central directory: `~/.lazyas/` holds config, manifest, cache, and skills
- Symlinks from backend paths (e.g., `~/.claude/skills/`) to `~/.lazyas/skills/`
- Config: `~/.lazyas/config.toml` (TOML, parsed by BurntSushi/toml)
- Manifest: `~/.lazyas/manifest.yaml` (YAML, tracks installed skills)
- Registry is a git repo containing `index.yaml`
- Skills are cloned via git sparse-checkout when possible
- Panel-based TUI (lazygit-style) with left/right split using Bubble Tea
- CLI built with Cobra

## Build

```
export GOROOT=/opt/go
make lint    # gofmt -s check + go vet (CI gate)
make test    # all unit tests (CI gate)
make build   # dev binary to bin/lazyas
make release # multi-arch release (requires tagged HEAD)
```

## Key Directories

```
cmd/lazyas/          Entry point, version injection
internal/cli/        Cobra commands (root, install, remove, update, sync, backend, ...)
internal/config/     Config struct, Backend struct, ExpandPath(), KnownBackends
internal/symlink/    Symlink lifecycle (check, create, remove, migrate)
internal/tui/        Bubble Tea app, panels (skills, detail), layout
internal/manifest/   Installed skills tracking
internal/registry/   Registry fetching, caching, index parsing
internal/git/        Clone and sparse-checkout operations
```
