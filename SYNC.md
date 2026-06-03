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
