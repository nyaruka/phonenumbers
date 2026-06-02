# Upstream Sync

This package is a port of Google's [libphonenumber](https://github.com/google/libphonenumber),
tracking its **Java** reference implementation. This file is the source of truth for
*which upstream version each part is reconciled against* and *what we deliberately
diverge on*.

Per-file headers only record which Java source a file maps to (e.g. `// Port of
PhoneNumberUtil.java ...`); the upstream version lives here so it isn't duplicated
across the tree and can't go stale file-by-file.

## Baseline

- **Code reconciled against:** `v9.0.31`

The **metadata** version isn't tracked here — `cmd/buildmetadata` records the upstream
release it built `data/` from in the generated
[`metadataversion.go`](metadataversion.go) (the exported `MetadataVersion` constant).

## Building metadata

`cmd/buildmetadata` regenerates the embedded `data/` from upstream libphonenumber:

1. `go run ./cmd/buildmetadata` — resolves the **latest** libphonenumber release tag,
   clones it, rebuilds `data/`, and records the tag in the generated
   [`metadataversion.go`](metadataversion.go). To rebuild from a specific release
   instead (e.g. to reproduce older embedded data), pass the tag:
   `go run ./cmd/buildmetadata v9.0.31`.
2. `go test ./...` and fix any reconciliation differences.
3. Review the `data/` and `metadataversion.go` diff. Reconciling the *code* against the
   new release is a human judgment, so it isn't automated: bump **Code reconciled
   against** in **Baseline** above and add a **Sync log** entry.

## Deliberate divergences

Places where this port intentionally differs from upstream:

- **`MaybeSeparateExtensionFromPhone`** (`phonenumberutil.go`) — a nyaruka-only helper
  with no upstream equivalent. Revisit for API purity.
- **v2 API cleanups** — the deprecated `FormatWithBuf` and the exported `StringBuilder`
  type exist only for backwards compatibility and are slated for removal in a future
  `/v2` module. See "Planned for v2" in the [README](README.md).
- **Unported upstream tests** — three `PhoneNumberUtilTest` cases have no Go
  counterpart: `testGetMetadataForRegionForMissingMetadata` and its non-geographical
  variant (both rely on Mockito-style metadata-source injection), and
  `testRemovalNotSupported` (Go's range-over-func iterator has no `remove()`).

## Sync log

Newest first. Each entry: what was reconciled or pulled, and the upstream version.

- **2026-06-02** — Established this file; moved per-file version headers here.
  `cmd/buildmetadata` now resolves the latest upstream release tag itself (pass a tag to
  pin a specific release) and records it in the generated `metadataversion.go`, so the
  embedded metadata's version is no longer hand-maintained here. `go test ./...` green.
- **2026-06-01** (v1.8.0) — Regenerated all metadata against `v9.0.31`; refactored to
  ease upstream syncing.
- Ported `PhoneNumberMatcher` + `PhoneNumberMatch` and their tests against `v9.0.31`.
