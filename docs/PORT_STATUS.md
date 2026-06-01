# Upstream test-port status

Tracks the ongoing effort to port Google libphonenumber's unit tests (Java is the
canonical reference) into this Go library, run against upstream's frozen
*synthetic* test metadata so the assertions never go stale when real-world
metadata is refreshed.

## Approach

Upstream's `PhoneNumberUtilTest` (and the other test classes) run against
`resources/PhoneNumberMetadataForTesting.xml` — a hand-authored set of regions
with made-up number ranges. We commit that XML under `testdata/`, compile it
through `builder.go`, and swap it in via the metadata-injection seam
(`useMetadata`) so the public package API behaves exactly as in production but
against frozen data.

- Faithful ports live in `phonenumberutil_*_test.go` and start each test with
  `useTestMetadata(t)`.
- The legacy ad-hoc tests (run against the real embedded metadata, and therefore
  break whenever upstream refreshes data) live in `phonenumberutil_adhoc_test.go`
  and are deleted as their faithful ports land.

Harness files: `testdata/PhoneNumberMetadataForTesting.xml`,
`testmetadata_test.go` (`useTestMetadata` + `regionCode`, the analogue of
upstream's `TestMetadataTestCase` + `RegionCode`), fixtures in
`phonenumberutil_test.go`. Last reconciled against upstream v9.0.32.

## Status

### PhoneNumberUtilTest — substantially ported
- ✅ Metadata loading, region / country-code lookup, number type, validation,
  example numbers, normalize / alpha, supported-types, possibility-by-type
- ✅ Parsing (national, international prefixes, non-ASCII, extensions, the
  invalid-number error table, keep-raw, phone-context)
- ✅ Formatting (per-country, by-pattern, out-of-country, carrier, mobile-dialing,
  in-original-format)
- Remaining: 2 skipped absent-type tests (see TODO below), the 2 Mockito
  missing-metadata tests, and a couple of string-overload possibility cases.

### Bugs the port surfaced and fixed
- Builder nil-deref on regions lacking a mobile / fixed-line pattern
- `GetNationalSignificantNumber` panic on a negative `numberOfLeadingZeros`
- `noInternationalDialling` XML struct-tag typo (the descriptor was silently
  dropped for every region)
- `$FG` / national-prefix formatting-rule application
- `UNIQUE_INTERNATIONAL_PREFIX` unanchored (out-of-country IDD prefix resolution)

## Remaining work (roughly in order)

1. **Absent-type metadata representation** (see TODO). Small; unblocks 2 tests and
   removes a divergence.
2. **Port `ShortNumberInfoTest`** (`shortnumberinfo_test.go` is still ad-hoc).
   Open question: upstream `resources/` has no `ShortNumberMetadataForTesting.xml`
   — confirm how upstream's `ShortNumberInfoTest` sources its test metadata before
   porting.
3. **Implement `AsYouTypeFormatter`** (currently absent) and port
   `AsYouTypeFormatterTest`.
4. **Implement `PhoneNumberMatcher` / `findNumbers`** (currently a `nil` stub in
   `phonenumbermatcher.go`) and port `PhoneNumberMatcherTest`.
5. **Add `ExampleNumbersTest`** — a real-metadata regression that validates every
   shipped region's example numbers parse and validate.
6. **Automate**: a scheduled task that detects new upstream releases, regenerates
   metadata, runs the (now-stable) synthetic tests, opens a PR for data-only
   deltas, and flags logic-touching changes for manual porting. See
   `docs/2.0-restructure.md`.

## Known TODOs / documented divergences

- **Absent-type representation.** Upstream marks a type with no numbers using
  `possibleLength = [-1]` (so `testNumberLength` returns `INVALID_LENGTH`); this
  builder instead leaves `possibleLength` empty with an `"NA"` pattern, so
  unsupported types fall back to the general desc's lengths. Aligning it un-skips
  `testIsPossibleNumberForType[WithReason]_NumberTypeNotSupportedForRegion` in
  `phonenumberutil_types_test.go` and removes the `"NA"` assertion in
  `TestGetInstanceLoadUSMetadata`. A naive fix (set `[-1]` in
  `processPhoneNumberDescElement` for absent elements) has subtle
  `FIXED_LINE_OR_MOBILE`-merge interactions in `testNumberLength`; it needs care
  plus full-suite verification.
- **Parsing edge cases:** `normalizeDigits` doesn't map non-ASCII / non-Arabic
  unicode digits (e.g. Mongolian) to ASCII; the extension regex doesn't tolerate
  trailing whitespace after the extension digits. Both are characterized in the
  parsing tests.
- `getExampleNumberForType` has no region-less overload; the relevant test uses a
  local helper.

## Conventions for continuing the port

- Mirror upstream method names and assertions. Start each test with
  `useTestMetadata(t)` (it is **not** safe to run such tests with `t.Parallel()`).
  Use `regionCode.*` and the fixture helpers in `phonenumberutil_test.go`.
- Map `NumberParseException` error types to the package's error sentinels
  (`ErrInvalidCountryCode`, `ErrNotANumber`, `ErrTooShortAfterIDD`,
  `ErrTooShortNSN`, `ErrNumTooLong`) and assert with `errors.Is`. Compare
  `*PhoneNumber` values with `proto.Equal`.
- When a divergence appears, **verify it against the current tree** before
  assuming it is real — don't characterize stale behavior.
- Delete the ad-hoc equivalent from `phonenumberutil_adhoc_test.go` as each
  upstream test lands.
- Verify with `go build ./...`, `go test -count=1 ./...`, `go test -race`,
  `go vet ./...`, and `gofmt`.
- Prefer working in a separate `git worktree` so concurrent work doesn't disturb a
  shared checkout.
