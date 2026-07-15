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
- **Defensive slice bounds in `buildNationalNumberForParsing`** — the national number is taken
  from the substring between a leading `tel:` and `;phone-context=`; we only treat `tel:` as the
  start marker when it precedes the phone-context, keeping the substring bounds well-ordered on
  arbitrary input. This is a robustness guard for Go's slice semantics — preserve it across syncs
  rather than reconciling it away.
- **Clamped leading-zero allocation in `GetNationalSignificantNumber`** — the leading-zero count
  is taken from the `PhoneNumber.NumberOfLeadingZeros` field, which is an unbounded `int32` when a
  message is populated from outside `Parse`. We clamp it to `maxLengthForNSN` before allocating,
  since a national significant number cannot have more leading zeros than its maximum length. This
  is a robustness guard on the allocation size — preserve it across syncs.
- **Lookup-subpackage details** — `carrier`, `geocoding`, and `timezone` mirror the upstream
  `Get*` method names and behaviour, with two intentional shape differences:
  `getUnknownTimeZone()` is exposed as the `timezone.Unknown` const rather than a function, and
  `geocoding` renders the country name via `golang.org/x/text` rather than `java.util.Locale`.
