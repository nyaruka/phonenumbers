# phonenumbers — AI agent guide

`phonenumbers` is a **vendored fork** of [nyaruka/phonenumbers](https://github.com/nyaruka/phonenumbers), itself a Go port of Google's [libphonenumber](https://github.com/google/libphonenumber). It is a pure library with no runtime infrastructure — treat changes with that in mind: prefer syncing fixes from upstream over hand-patching, see [`coding-guidelines.md`](coding-guidelines.md).

Agentic material for this repo is authored once in [`.agents/`](.agents/README.md) and consumed by Claude Code, Windsurf, and Devin.

## Where to find what

| What | Location |
|---|---|
| Upstream-first conventions, generated-file rules | [`coding-guidelines.md`](coding-guidelines.md) |
| Keeping docs current | [`.agents/rules/keep-docs-current.md`](.agents/rules/keep-docs-current.md) |
| Skills, workflows, sub-agents (none yet) | [`.agents/README.md`](.agents/README.md) |
| Architecture / package structure | [`docs/architecture.md`](docs/architecture.md) |
| Local setup, testing, upstream sync, releasing | [`docs/local-development.md`](docs/local-development.md) |
| Go module and upstream fork-chain dependencies | [`docs/dependencies.md`](docs/dependencies.md) |
| Common failure modes | [`docs/troubleshooting.md`](docs/troubleshooting.md) |
| Domain terminology | [`docs/glossary.md`](docs/glossary.md) |
| Repo metadata (machine-readable) | [`context.yaml`](context.yaml) |

## Editing rule

`.agents/` is the source of truth for repo-local rules, skills, workflows, and sub-agents. `.claude/` and `.windsurf/` (if present) contain **stubs** — frontmatter plus a pointer line, never content.

Shared skills (`ah-sdlc:*`, `ah-dev:*`, `ah-data:*`) come from the [ah-agents](https://github.com/Accompany-Health/ah-agents) marketplace via `/plugin install` and are never vendored here.

## Before you change library behavior

Read [`coding-guidelines.md`](coding-guidelines.md) first. This repo aims to be a strict port of libphonenumber — check whether a fix already exists upstream (`nyaruka/phonenumbers` or `google/libphonenumber`) before writing one locally, per [`docs/troubleshooting.md`](docs/troubleshooting.md).
