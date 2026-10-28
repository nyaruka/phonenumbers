# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Required reading

This repo is a **vendored fork** of [nyaruka/phonenumbers](https://github.com/nyaruka/phonenumbers) (itself a Go port of Google's [libphonenumber](https://github.com/google/libphonenumber)) — not an Accompany-authored service. Read these before making changes:

| Read this | For |
|---|---|
| [`coding-guidelines.md`](coding-guidelines.md) | Upstream-first conventions: when to sync vs. hand-patch, generated-file rules, error/constant style |
| [`.agents/rules/keep-docs-current.md`](.agents/rules/keep-docs-current.md) | Update docs in the same PR as the change that invalidates them |
| [`docs/architecture.md`](docs/architecture.md) | Package structure and public API surface |
| [`docs/dependencies.md`](docs/dependencies.md) | Go module deps and the upstream fork chain (google/libphonenumber → nyaruka/phonenumbers → this repo) |

Agentic material for this repo lives in [`.agents/`](.agents/README.md) and is shared with Windsurf and Devin. Files under `.claude/` (if any) are **stubs** that point back into `.agents/` — never edit content there directly; see [`.agents/README.md`](.agents/README.md).

## Development commands

```bash
# Tests
go test ./...
go test -race ./...

# Lint
golangci-lint run

# Build
go build ./...

# Refresh metadata from upstream Google libphonenumber
go install github.com/Accompany-Health/phonenumbers/cmd/buildmetadata
$GOPATH/bin/buildmetadata
```

Full setup, upstream-sync, and release instructions are in [`docs/local-development.md`](docs/local-development.md).

## Shared tooling

Ticket, PR, planning, and verification skills come from the [ah-agents](https://github.com/Accompany-Health/ah-agents) plugin marketplace (`ah-sdlc:*`, `ah-dev:*`, `ah-data:*`) and are installed per-developer with `/plugin install`. **Do not vendor them into this repo.**
