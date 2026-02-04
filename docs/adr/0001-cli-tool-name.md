# ADR-0001: CLI Tool Name

**Status:** Accepted
**Date:** 2026-02-01
**Deciders:** Team

## Context

We need a name for the CLI tool that manages Claude Code agent skills. The tool handles installation, updates, listing, and removal of skills from `~/.claude/skills/`.

The name should be:
- Short enough for frequent command-line use
- Descriptive enough to indicate purpose
- Unlikely to conflict with existing system commands

## Decision

We will name the CLI tool **`skillpkg`** (skill package).

## Alternatives Considered

| Name | Pros | Cons |
|------|------|------|
| `skillpkg` | Descriptive, matches internal directory name | Long (8 chars) |
| `skpkg` | Short (5 chars), still readable | Cryptic, requires learning abbreviation |
| `sk` | Ultra-short | Very generic, potential conflicts |
| `csk` | Short, "Claude skill" | Unclear pronunciation |
| `askill` | "Agent skill", clear | 6 chars |
| `skill` | Clean, obvious | Likely conflicts with system commands |
| `skillctl` | Follows kubectl pattern | Overkill for scope |

## Consequences

### Positive
- Self-documenting name, clear purpose
- Matches existing internal directory name (`.skillpkg/`)
- Clear association with package management (`pkg` suffix)
- No known conflicts with common CLI tools

### Negative
- Slightly longer than minimal (8 chars)

## References

- Storage location: `~/.claude/skills/.skillpkg/`
- Manifest: `~/.claude/skills/.skillpkg/manifest.yaml`
