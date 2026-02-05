# ADR-0001: CLI Tool Name

**Status:** Accepted
**Date:** 2026-02-01
**Deciders:** Team

## Context

We need a name for the CLI tool that manages agent skills. The tool handles browsing, installation, updates, and removal of skills across multiple AI agent backends.

The name should be:
- Short enough for frequent command-line use
- Descriptive enough to indicate purpose
- Unlikely to conflict with existing system commands

## Decision

We will name the CLI tool **`lazyas`** (lazy agent skills).

The name follows the convention of the **lazy\* TUI tool family** — lazygit, lazydocker, lazynpm — which share a common design philosophy:

- **Panel-based TUI** with keyboard-driven navigation (j/k, enter, tab)
- **Package-manager workflow**: browse, install, remove, update
- **Opinionated defaults** that minimise configuration for the common case
- **Git-aware operations** (lazyas uses sparse-checkout, per-repo clones)

lazyas applies this same workflow and intent to agent skills: a lazygit-style interface for managing skill repositories, with the familiar browse → select → install loop.

## Alternatives Considered

| Name | Pros | Cons |
|------|------|------|
| `lazyas` | Fits lazy\* family, clear intent, 6 chars | Requires knowing the convention |
| `skillpkg` | Descriptive, matches `pkg` pattern | Long (8 chars), no TUI connotation |
| `skpkg` | Short (5 chars) | Cryptic |
| `sk` | Ultra-short | Very generic, potential conflicts |
| `askill` | "Agent skill", clear | No TUI connotation |
| `skillctl` | Follows kubectl pattern | Overkill for scope |

## Consequences

### Positive
- Immediately signals TUI-first, keyboard-driven design to users familiar with lazygit/lazydocker
- Short and memorable for daily command-line use
- Clear association with agent skills (`as` = agent skills)
- No known conflicts with common CLI tools

### Negative
- Users unfamiliar with the lazy\* family may not recognise the convention at first glance

## References

- Storage location: `~/.lazyas/`
- Config: `~/.lazyas/config.toml`
- Manifest: `~/.lazyas/manifest.yaml`
- Inspiration: [lazygit](https://github.com/jesseduffield/lazygit), [lazydocker](https://github.com/jesseduffield/lazydocker), [lazynpm](https://github.com/jesseduffield/lazynpm)
