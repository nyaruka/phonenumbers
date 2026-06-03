# Upstream Sync

This package is a port of Google's [libphonenumber](https://github.com/google/libphonenumber),
tracking its **Java** reference implementation. This file records the **state** of the port:
the upstream version the code is reconciled against, and the deliberate divergences from
upstream.

For *how* to perform a sync, see the **Updating Metadata** / **Updating Code** sections of the
[README](README.md), or the `sync-upstream` skill
([`.claude/skills/sync-upstream/SKILL.md`](.claude/skills/sync-upstream/SKILL.md)). The
procedure is not duplicated here.

## Code reconciled against

- **v9.0.31**

This is the upstream release whose **Java logic** the Go code is reconciled against. It is
tracked separately from the embedded **metadata** version — recorded in the generated
[`metadata/version.go`](metadata/version.go) (`metadata.Version`) — which can move ahead of
this when metadata is regenerated without a code change.

## Deliberate divergences

Places where this port intentionally differs from upstream:

- **Unported upstream tests** — three `PhoneNumberUtilTest` cases have no Go counterpart:
  `testGetMetadataForRegionForMissingMetadata` and its non-geographical variant (both rely on
  Mockito-style metadata-source injection), and `testRemovalNotSupported` (Go's range-over-func
  iterator has no `remove()`).
- **Lookup-subpackage naming** — `carrier`, `geocoding`, and `timezone` keep the upstream
  method *behaviour* but drop the Java `get` prefix (`getNameForNumber` → `carrier.NameForNumber`,
  `getDescriptionForNumber` → `geocoding.DescriptionForNumber`, `getTimeZonesForNumber` →
  `timezone.TimeZonesForNumber`), and `getUnknownTimeZone()` is the `timezone.Unknown` const.
  `geocoding` renders the country name via `golang.org/x/text` rather than `java.util.Locale`.
  Don't "restore" the `Get` prefix on a sync.
