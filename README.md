# lazyas - Lazy Agent Skills Manager

A lazygit/lazydocker-style TUI for managing Agent Skills with an upstream registry.

## Overview

**lazyas** is a package manager for Agent Skills - it handles browsing, installing, updating, and removing skills from a centralized registry. Think `apt`, `brew`, or `npm` for AI agent skills.

### Multi-Backend Support

lazyas manages ONE central skills directory (`~/.lazyas/skills/`) and uses symlinks to connect multiple AI agent backends. Each backend symlinks its skills directory to the central location, so all backends share the same skills without duplication.

Built-in backends: Claude Code, OpenAI Codex, Gemini CLI, Cursor, GitHub Copilot, Amp, Goose, OpenCode, and Mistral Vibe.

### Scope

lazyas manages **Agent Skills only**. It does not handle MCP servers, model configurations, or other plugin systems.

### Not a Skill Creator

lazyas is a management tool, not a skill authoring tool. For creating skills, see [skillcreator.ai](https://www.skillcreator.ai/).

## Installation

```bash
cd lazyas
make build
make install
```

## Usage

### Interactive Browser

Launch the TUI to browse and manage skills:

```bash
lazyas
# or
lazyas browse
```

The interface features a two-panel layout:
- **Left Panel**: Skills grouped by Installed/Available with collapsible sections
- **Right Panel**: Detail view with Info and SKILL.md tabs

Key bindings:
- `j/k` or `↑/↓` - Navigate up/down in current panel
- `Tab` or `h/l` - Switch focus between panels
- `[/]` - Switch tabs in detail panel
- `z` - Collapse/expand group
- `i` - Install selected skill
- `r` - Remove selected skill
- `U` - Update all installed skills
- `S` - Sync repositories (force refresh)
- `b` - Backend management
- `/` - Search skills
- `Esc` - Clear search
- `a` - Add repository
- `q` - Quit

### CLI Commands

```bash
# Install a skill
lazyas install <name>[@version]
lazyas install my-skill
lazyas install my-skill@v1.2.0
lazyas install --force my-skill    # Overwrite modified

# Remove a skill
lazyas remove <name>
lazyas rm my-skill

# List skills
lazyas list              # List installed skills
lazyas list --available  # List available skills
lazyas list --all        # List all with install status

# Search skills
lazyas search <query>

# Update skills
lazyas update                # Update all
lazyas update <name>         # Update specific skill
lazyas update --dry-run      # Preview updates
lazyas update --force        # Update even modified skills

# Sync registry
lazyas sync                  # Force refresh from all repos

# Show skill info
lazyas info <name>

# Backend management
lazyas backend list              # Show backends and link status
lazyas backend link              # Link all unlinked backends
lazyas backend link claude       # Link specific backend
lazyas backend unlink claude     # Remove symlink
lazyas backend add myai ~/.myai/skills
lazyas backend remove myai

# Configuration
lazyas config show
lazyas config repo add <name> <url>
lazyas config repo remove <name>
lazyas config repo list
```

## Architecture

### Directory Structure

```
~/.lazyas/
├── skills/              # Symlinks into repo worktrees
│   ├── my-skill → repos/anthropics-skills/skills/my-skill
│   ├── helper-skill → repos/anthropics-skills/skills/helper-skill
│   └── ...
├── repos/               # Per-repo sparse clones
│   └── anthropics-skills/
├── config.toml          # Configuration
├── manifest.yaml        # Installed skills tracking
└── cache.yaml           # Registry cache

# Symlinks (created by lazyas)
~/.claude/skills → ~/.lazyas/skills
~/.codex/skills → ~/.lazyas/skills
~/.gemini/skills → ~/.lazyas/skills
~/.cursor/skills → ~/.lazyas/skills
~/.copilot/skills → ~/.lazyas/skills
$XDG_CONFIG_HOME/agents/skills → ~/.lazyas/skills    # Amp
$XDG_CONFIG_HOME/goose/skills → ~/.lazyas/skills     # Goose
$XDG_CONFIG_HOME/opencode/skills → ~/.lazyas/skills  # OpenCode
~/.vibe/skills → ~/.lazyas/skills               # Mistral Vibe
```

### Code Structure

```
internal/
├── tui/
│   ├── app.go              # Main TUI application (panel-based)
│   ├── layout/
│   │   └── panels.go       # Two-panel layout manager
│   ├── panels/
│   │   ├── skills.go       # Left panel: grouped skill list
│   │   └── detail.go       # Right panel: detail with tabs
│   ├── styles/             # Lipgloss styles
│   └── testing/            # Test harness and mocks
├── registry/               # Index fetching and caching
├── manifest/               # Local manifest management
├── config/                 # Configuration (backends, repos)
├── symlink/                # Symlink management for backends
├── skillmd/                # Shared SKILL.md parsing helpers
├── git/                    # Git operations (repo clones, sparse checkout)
└── cli/                    # Cobra CLI commands
```

## Configuration

Configuration is stored in `~/.lazyas/config.toml`:

```toml
[[repos]]
name = "official"
url = "https://github.com/example/skills-index"

[[backends]]
name = "work-tool"
path = "~/work/.ai/skills"
description = "Internal AI tool"
```

Built-in backends (claude, codex, gemini, cursor, copilot, amp, goose, opencode, vibe) are configured automatically. Custom backends can be added via `lazyas backend add` or the config file.

## Popular Skill Repositories

These repositories contain Agent Skills that can be added as lazyas registry sources:

```bash
lazyas config repo add <name> <url>
```

| Repository | Stars | Skills | Description |
|---|---|---|---|
| [anthropics/skills](https://github.com/anthropics/skills) | 63k | 17 | Anthropic's official skills - webapp testing, canvas design, document generation |
| [vercel-labs/agent-skills](https://github.com/vercel-labs/agent-skills) | 19k | 5 | Vercel's official collection - React best practices, web design, deploy |
| [muratcankoylan/Agent-Skills-for-Context-Engineering](https://github.com/muratcankoylan/Agent-Skills-for-Context-Engineering) | 8k | 19 | Context engineering, multi-agent architectures, memory management |
| [sickn33/antigravity-awesome-skills](https://github.com/sickn33/antigravity-awesome-skills) | 7k | 600+ | Largest collection - security, React, autonomous coding, and more |
| [Orchestra-Research/AI-research-SKILLs](https://github.com/Orchestra-Research/AI-research-SKILLs) | 2.3k | 82 | AI research & engineering - ML, NLP, computer vision, scientific computing |
| [alirezarezvani/claude-skills](https://github.com/alirezarezvani/claude-skills) | 1.6k | 42 | Practical Claude workflows, subagents, and commands |
| [skillcreatorai/Ai-Agent-Skills](https://github.com/skillcreatorai/Ai-Agent-Skills) | 716 | 47 | General purpose skills from the skillcreator.ai ecosystem |
| [microsoft/agent-skills](https://github.com/microsoft/agent-skills) | 542 | 133 | Azure, Cosmos DB, SDKs - Microsoft ecosystem skills |

Repos without an `index.yaml` are auto-scanned for `SKILL.md` files during sync.

## Registry Format

The registry is a git repository containing an `index.yaml`:

```yaml
version: 1
metadata:
  name: "skills-registry"
  updated_at: "2025-01-30T12:00:00Z"

skills:
  - name: "my-skill"
    description: "Example skill for AI agents"
    source:
      repo: "https://github.com/example/skills"
      path: "skills/my-skill"  # subdirectory support
      tag: "v1.2.0"
    author: "example-author"
    tags: [example, utility]
```

## Skill Format

Each skill must contain a `SKILL.md` file that describes the skill's capabilities and triggers.

## UI Design

The UI follows the lazy* tool design pattern:

```
lazyas  Lazy Agent Skills      claude ✓ codex ✓ gemini ✓ ...
┌──────────────────┬──────────────────────────────┐
│   Skills Panel   │        Detail Panel          │
│                  │  ┌────┬─────────┐            │
│ ▼ Installed (2)  │  │Info│ SKILL.md│            │
│   ● my-skill     │  └────┴─────────┘            │
│   ● helper       │                              │
│ ▼ Available      │  Name: my-skill              │
│   ○ new-skill    │  Author: example             │
│   ○ other-skill  │  Installed: Yes (v1.0.0)     │
│                  │  Modified: No                │
├──────────────────┴──────────────────────────────┤
│ j/k navigate  i install  r remove  U update all │
│ S sync  b backends  / search  q quit            │
└─────────────────────────────────────────────────┘
```

- Purple borders indicate the active panel
- `●` = installed, `○` = available, `◉` = modified
- Collapsible groups with `▼`/`▶` indicators
- Backend status shown in header

## Development

```bash
# Build
make build

# Run in development
make run ARGS="browse"
make browse

# Format code
make fmt

# Run tests
make test
```
