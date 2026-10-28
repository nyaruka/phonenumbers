---
name: keep-docs-current
description: Documentation and agentic material must be updated in the same change that invalidates it
target: both
---

# Keep documentation current

Stale documentation is worse than missing documentation: an agent reads it with the same confidence as correct documentation and acts on it.

**Update the docs in the same PR as the change that invalidates them.** Not a follow-up ticket. A reviewer should be able to reject a PR for leaving a doc wrong, the same way they would for a missing test.

Org-level requirements are in `ah-agents/docs/repository-standards.md` → **Documentation Maintenance**. This rule is the local enforcement.

## What to update, and when

| If you change… | Update |
|---|---|
| Package structure, public API surface, or metadata embedding | [`docs/architecture.md`](../../docs/architecture.md) |
| Setup or testing/linting instructions | [`docs/local-development.md`](../../docs/local-development.md) |
| Go module dependencies, or the upstream fork chain | [`docs/dependencies.md`](../../docs/dependencies.md) |
| A new failure mode, or the fix for one | [`docs/troubleshooting.md`](../../docs/troubleshooting.md) |
| Phone-number or libphonenumber terminology | [`docs/glossary.md`](../../docs/glossary.md) |
| A convention or anti-pattern a reviewer would otherwise re-explain | [`coding-guidelines.md`](../../coding-guidelines.md) |
| Ownership, team, Slack, tier | [`context.yaml`](../../context.yaml), the README's Ownership section, and `CODEOWNERS` |
| A `.proto` definition | regenerate the corresponding `.pb.go` file — never hand-edit generated code |
| The upstream metadata (`data/`) | run `cmd/buildmetadata` and note the upstream version in `CHANGELOG.md` |
| A procedure encoded in a script or workflow | the matching file under [`.agents/`](../README.md) |

## Files the context manifest promises exist

[`context.yaml`](../../context.yaml) declares these under `docs:` and `important_paths:`. If you move or delete one, update the manifest in the same change — it is machine-readable and other tooling consumes it:

`README.md` · `docs/architecture.md` · `docs/local-development.md` · `docs/dependencies.md` · `docs/troubleshooting.md` · `docs/glossary.md`

## Don't hardcode values that drift

Reference the file that holds a value instead of copying the value into prose — dependency versions (`go.mod`), the upstream metadata version (`CHANGELOG.md`), and the current release tag all go stale the moment they're pasted into a doc.

## Agentic material is documentation too

Everything under `.agents/` is subject to this rule. When a rule describes code or docs that changed, fix it in the same PR.
