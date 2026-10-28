# `.agents/` — agentic material for `phonenumbers`

This directory is the **single source of truth** for every repo-local rule, skill, workflow, and sub-agent in this repo. It is tool-neutral: Claude Code, Windsurf, and Devin all consume it.

## Edit here, never in `.claude/` or `.windsurf/`

Agent tools only discover content from fixed paths — Claude Code from `.claude/{skills,agents,commands}`, Windsurf and Devin from `.windsurf/{rules,skills,workflows}`. A prose pointer cannot create a slash command or a dispatchable sub-agent, so those directories hold **stubs**: real frontmatter plus a single line pointing back here.

```markdown
---
name: example-skill
description: <copied verbatim from the .agents/ source>
---

Canonical definition: `.agents/skills/example-skill/SKILL.md`. Read that file and follow it.
```

A stub never contains content. If you find yourself editing prose inside `.claude/` or `.windsurf/`, you are editing the wrong file.

## Layout

| Path | What goes here |
|---|---|
| `rules/` | Always-on conventions and guardrails. Short, imperative, non-negotiable. |
| `skills/<name>/SKILL.md` | On-demand procedural knowledge — "how do I do X in this repo". |
| `workflows/<name>.md` | Multi-step procedures, surfaced as slash commands. |
| `agents/<name>/AGENT.md` | Sub-agent personas. |

This repo currently has no skills, workflows, or sub-agents — only the always-on rule in `rules/`. Add directories as real repo-local agentic material accumulates; don't pre-create empty stubs.

## Frontmatter

```yaml
---
name: kebab-case          # must equal the containing directory (or filename, for rules/workflows)
description: One line     # copied verbatim into stubs — make it self-contained
target: both              # claude-code | windsurf | both — governs which stubs exist
---
```

## Adding something

1. Create the canonical file here with complete frontmatter.
2. Add the matching stub(s) under `.claude/` and `.windsurf/` for whatever `target:` declares.
3. Keep stub bodies to the pointer line only — no content.

## What does *not* belong here

Anything reusable across Accompany Health repos. Shared skills and agents (`ah-sdlc:*`, `ah-dev:*`, `ah-data:*`) are installed with `/plugin install` from [ah-agents](https://github.com/Accompany-Health/ah-agents) and must never be vendored into this repo.

**The test:** could another Go service use this unchanged? If yes, it belongs in `ah-agents`, not here. Only `phonenumbers`-specific knowledge (upstream sync, metadata refresh, the vendored-library constraints) stays local.

Prose documentation belongs in `docs/`; rules and skills link to it rather than restating it.

## Standard

This layout implements the **Repo-Local Agentic Material** section of `ah-agents/docs/repository-standards.md`, and keeps instruction files thin per `ah-agents/docs/agentic-sdlc-standards.md` first principle #4. `mercury` is the reference implementation of this pattern.
