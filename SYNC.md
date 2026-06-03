# Upstream Sync

This package is a port of Google's [libphonenumber](https://github.com/google/libphonenumber),
tracking its **Java** reference implementation. This file is the **state** of the sync:
*which upstream version the code is reconciled against* and *what we deliberately diverge on*.

The **procedure** for performing a sync lives in the `sync-upstream` skill
([`.claude/skills/sync-upstream/SKILL.md`](.claude/skills/sync-upstream/SKILL.md)) — run it
with "sync with upstream" or `/sync-upstream`.

Per-file headers carry the Go→Java mapping: each ported file records the full upstream path it
maps to (e.g. `// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberUtil.java`).
The reconciled version isn't duplicated into those headers — it lives here, so it can't go stale
file-by-file.

## Baseline

- **Code reconciled against:** `v9.0.31`

The **metadata** version isn't tracked here — `cmd/buildmetadata` records the upstream release it
built `data/` from in the generated [`metadata/version.go`](metadata/version.go) (the
`metadata.Version` constant).

## Deliberate divergences

Places where this port intentionally differs from upstream:

- **Unported upstream tests** — three `PhoneNumberUtilTest` cases have no Go counterpart:
  `testGetMetadataForRegionForMissingMetadata` and its non-geographical variant (both rely on
  Mockito-style metadata-source injection), and `testRemovalNotSupported` (Go's range-over-func
  iterator has no `remove()`).

## Sync log

Newest first. Each entry: what was reconciled or pulled, and the upstream version.

- **2026-06-03** — Moved the sync *procedure* into the `sync-upstream` skill; SYNC.md is now
  pure state. Per-file `// Port of` headers upgraded to full upstream paths so reconciliation
  diffs are mechanical.
- **2026-06-02** — Established this file; moved per-file version headers here.
  `cmd/buildmetadata` now resolves the latest upstream release tag itself (pass a tag to
  pin a specific release) and records it in the generated `metadata/version.go`, so the
  embedded metadata's version is no longer hand-maintained here. `go test ./...` green.
- **2026-06-01** (v1.8.0) — Regenerated all metadata against `v9.0.31`; refactored to
  ease upstream syncing.
- Ported `PhoneNumberMatcher` + `PhoneNumberMatch` and their tests against `v9.0.31`.
