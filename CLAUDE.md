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

**lazyas aims to support multiple AI agent backends, not just Claude.**

The "Lazy Agent Skills" name reflects this broader scope. While the initial implementation targets Claude Code's skills directory (`~/.claude/skills/`), the architecture is designed to support other backends as they adopt the Agent Skills format.

Current backend: Claude Code
Future backends: TBD (as Agent Skills adoption grows)

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

- Panel-based TUI (lazygit-style) with left/right split
- Registry is a git repo containing `index.yaml`
- Skills are cloned via git sparse-checkout when possible
- Local state stored in `.lazyas/` within the skills directory
