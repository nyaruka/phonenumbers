# Upstream Sync

This package is a port of Google's [libphonenumber](https://github.com/google/libphonenumber),
tracking its **Java** reference implementation. This file is the source of truth for
*which upstream version each part is reconciled against* and *what we deliberately
diverge on*.

Per-file headers only record which Java source a file maps to (e.g. `// Port of
PhoneNumberUtil.java ...`); the upstream version lives here so it isn't duplicated
across the tree and can't go stale file-by-file.

## Baseline

- **Code reconciled against:** `9.0.32-SNAPSHOT` (upstream `master`, post-`v9.0.31`)
- **Metadata built from:** `9.0.32-SNAPSHOT` (regenerated 2026-06-01, v1.8.0)

> `9.0.32-SNAPSHOT` is upstream's *in-development* version on `master` after the
> `v9.0.31` release — there is no `v9.0.32` tag (yet). This baseline predates pinning:
> it was synced from an untagged `master` commit, so it is not exactly reproducible.
> From the next sync onward, `cmd/buildmetadata` pins to a release **tag** (see below),
> so each future baseline is a real, reproducible `vX.Y.Z`. Until then, the embedded
> metadata sits slightly *ahead* of the pinned tag.

## Pinning

`cmd/buildmetadata` clones a pinned upstream **release tag** (`git clone --branch`),
set by the `upstreamVersion` constant in [`cmd/buildmetadata/main.go`](cmd/buildmetadata/main.go)
(currently `v9.0.31`). To sync to a newer release:

1. Bump `upstreamVersion` to the new tag.
2. `go run ./cmd/buildmetadata` to regenerate the embedded metadata.
3. `go test ./...` and fix any reconciliation differences.
4. Update the **Baseline** and **Reconcile status** below, and add a **Sync log** entry.

## Reconcile status

Each ported Go file and the upstream Java source it mirrors. "Reconciled @" is the
upstream version the Go code was last checked against; test files track the same
version as the code they exercise.

| Go file | Upstream Java source | Reconciled @ |
| --- | --- | --- |
| `phonenumberutil.go` | `PhoneNumberUtil.java` | 9.0.32-SNAPSHOT |
| `enums.go` | `PhoneNumberUtil.java` (enums) | 9.0.32-SNAPSHOT |
| `errors.go` | `NumberParseException.java` | 9.0.32-SNAPSHOT |
| `shortnumberinfo.go` | `ShortNumberInfo.java` | 9.0.32-SNAPSHOT |
| `asyoutypeformatter.go` | `AsYouTypeFormatter.java` | 9.0.32-SNAPSHOT |
| `phonenumbermatch.go` | `PhoneNumberMatch.java` | 9.0.32-SNAPSHOT |
| `phonenumbermatcher.go` | `PhoneNumberMatcher.java` | 9.0.32-SNAPSHOT |
| `carrier.go` | `carrier/PhoneNumberToCarrierMapper.java` | 9.0.32-SNAPSHOT |
| `geocoding.go` | `geocoder/geocoding/PhoneNumberOfflineGeocoder.java` | 9.0.32-SNAPSHOT |
| `timezone.go` | `geocoder/PhoneNumberToTimeZonesMapper.java` | 9.0.32-SNAPSHOT |
| `prefixmapper.go` | `internal/prefixmapper/*` | 9.0.32-SNAPSHOT |
| `metadatasource.go` | `metadata/source/*` + `MetadataLoader` | 9.0.32-SNAPSHOT |
| `alternateformats.go` | `metadata/source/*` (alternate formats) | 9.0.32-SNAPSHOT |
| `builder.go` | `tools/.../BuildMetadataFromXml.java` | 9.0.32-SNAPSHOT |

Generated or built, not hand-ported:

| Go file / dir | Source |
| --- | --- |
| `phonemetadata.pb.go` | `protoc` from `phonemetadata.proto` |
| `phonenumber.pb.go` | `protoc` from `phonenumber.proto` |
| `data/*.gz` | `cmd/buildmetadata`, from upstream `resources/` |

Go-specific glue with no single upstream source (not tracked above): `doc.go`,
`metadata.go`, `metadata_util.go`, `serialize.go`, `stringbuilder.go`.

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

- **2026-06-02** — Established this file; moved per-file version headers here; pinned
  `cmd/buildmetadata` to release tags. Baseline `9.0.32-SNAPSHOT`.
- **2026-06-01** (v1.8.0) — Regenerated all metadata against `9.0.32-SNAPSHOT`;
  refactored to ease upstream syncing.
- Ported `PhoneNumberMatcher` + `PhoneNumberMatch` and their tests against
  `9.0.32-SNAPSHOT`.
