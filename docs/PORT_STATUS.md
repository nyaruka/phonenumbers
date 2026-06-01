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
- Remaining: the 2 Mockito missing-metadata tests, and a couple of
  string-overload possibility cases.

### ShortNumberInfoTest — ported
- ✅ Faithful port in `shortnumberinfo_test.go`, reconciled against v9.0.32.
  Open question resolved: upstream has **no** `ShortNumberMetadataForTesting.xml`
  because `ShortNumberInfo.getInstance()` uses the production short metadata; only
  its `parse()` uses test metadata. The Go port therefore runs against the
  embedded short metadata (a real-metadata regression, like upstream).
- New public API implemented to support the port: `GetExpectedCost` /
  `GetExpectedCostForRegion` (+ the `ShortNumberCost` type), `IsCarrierSpecific` /
  `IsCarrierSpecificForRegion`, `IsSmsServiceForRegion`, and the
  `getExampleShortNumber[ForCost]` helpers.
- One skipped test: `TestIsSmsService` (see TODO below).

### ExampleNumbersTest — ported
- ✅ Faithful port in `examplenumbers_test.go`, reconciled against v9.0.32. A
  real-metadata regression (upstream uses the production singletons), so it runs
  against the embedded metadata: every supported region's per-type example numbers
  parse, validate, and classify correctly; every region has example / invalid-
  example numbers; every type has an example; non-geo and short-number (cost,
  emergency, carrier-specific, SMS) examples check out. Exercises all 245 regions.
- Test names mirror the upstream method names. The sole exception is
  `TestCanBeInternationallyDialledExampleNumbers`: `testCanBeInternationallyDialled`
  exists in both `ExampleNumbersTest` and `PhoneNumberUtilTest`, and the latter is
  already ported as `TestCanBeInternationallyDialled`, so this one is suffixed to
  disambiguate within the flat Go package.
- `getExampleNumberForType` still has no region-less overload, so the
  every-type test uses a local helper (`getExampleNumberForTypeAnyRegion`).

### Bugs the port surfaced and fixed
- Builder nil-deref on regions lacking a mobile / fixed-line pattern
- `GetNationalSignificantNumber` panic on a negative `numberOfLeadingZeros`
- `noInternationalDialling` XML struct-tag typo (the descriptor was silently
  dropped for every region)
- `$FG` / national-prefix formatting-rule application
- `UNIQUE_INTERNATIONAL_PREFIX` unanchored (out-of-country IDD prefix resolution)
- Absent-type metadata representation: the builder now marks a type with no
  numbers using `possibleLength = [-1]` (matching upstream) instead of an `"NA"`
  national pattern, so `testNumberLength` reports `INVALID_LENGTH` for
  unsupported types. `descHasPossibleNumberData` was aligned with upstream
  (empty ⇒ inherits general desc; `[-1]` ⇒ unsupported) while still treating the
  legacy `"NA"` sentinel in the committed embedded metadata as "no data".
- Short-number builder dropped `<smsServices>`: the short-metadata branch of
  `setRelevantDescPatterns` never processed the element, so `IsSmsServiceForRegion`
  could never match. The builder now reads it (verified by
  `TestBuilderProcessesSmsServices`); the embedded data still needs regenerating
  for it to carry the data (see TODO).

## Remaining work (roughly in order)

1. **Implement `AsYouTypeFormatter`** (currently absent) and port
   `AsYouTypeFormatterTest`.
2. **Implement `PhoneNumberMatcher` / `findNumbers`** (currently a `nil` stub in
   `phonenumbermatcher.go`) and port `PhoneNumberMatcherTest`.
3. **Automate**: a scheduled task that detects new upstream releases, regenerates
   metadata, runs the (now-stable) synthetic tests, opens a PR for data-only
   deltas, and flags logic-touching changes for manual porting. See
   `docs/2.0-restructure.md`.

## Known TODOs / documented divergences

- **Embedded metadata still uses the legacy `"NA"` sentinel.** The builder now
  emits `possibleLength = [-1]` for absent types (matching upstream), but the
  committed `data/metadata.xml.gz` was generated before that change and still
  marks absent types with an `"NA"` national pattern. `descHasData` /
  `descHasPossibleNumberData` handle both representations, so behaviour is
  correct; the `"NA"` special-cases become dead code only once the embedded
  metadata is regenerated (a separate data-refresh concern — see the automate item).
- **`TestIsSmsService` is skipped.** `IsSmsServiceForRegion` is implemented and the
  builder now reads `<smsServices>`, but the committed embedded short metadata was
  generated before that builder support, so it carries no smsServices data and the
  test would always see `false`. Un-skips once the short metadata is regenerated.
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
