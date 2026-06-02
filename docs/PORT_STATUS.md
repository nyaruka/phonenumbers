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

### PhoneNumberUtilTest — ported (except 2 Mockito tests)
- ✅ Metadata loading, region / country-code lookup, number type, validation,
  example numbers, supported-types, possibility (incl. by-type, with-reason,
  not-possible)
- ✅ Parsing (national, international prefixes, non-ASCII, extensions, the
  invalid-number error table, keep-raw, phone-context, Italian leading zeros,
  national-prefix / international-prefix stripping, country-code extraction)
- ✅ Formatting (per-country, by-pattern, out-of-country, carrier, mobile-dialing,
  in-original-format)
- ✅ Number matching — the full `testIsNumberMatch*` family (8 methods) in
  `phonenumberutil_isnumbermatch_test.go`
- ✅ Normalization / viability / extraction (`testConvertAlphaCharactersInNumber`,
  the `testNormalise*` family, `testIsViablePhoneNumber` (+ `NonAscii`),
  `testExtractPossibleNumber`) in `phonenumberutil_normalize_test.go`; truncation
  and possibility in `phonenumberutil_possibility_test.go`.
- ⚠️ Not ported: only the 2 Mockito missing-metadata tests
  (`testGetMetadataForRegionForMissingMetadata` and the non-geographical variant),
  which rely on Mockito-style metadata-source injection.

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
- `TestIsSmsService` is active: the embedded short metadata was regenerated, so it
  now carries `<smsServices>` data (see the metadata-refresh note below).

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

### AsYouTypeFormatterTest — ported
- ✅ Faithful port in `asyoutypeformatter_test.go` (all 33 upstream test methods),
  run against the synthetic test metadata via `useTestMetadata(t)`, reconciled
  against v9.0.32. Covers invalid region / plus sign, the too-long multiple-leading-
  digits case, national-prefix-with-space (incl. long NDD), per-region formatting
  (US incl. full-width input, mobile short codes and vanity numbers; GB / NZ / DE /
  AR / KR / MX / JP / AU / SG / IT / AO / CN), remembered-position tracking, long
  IDD / NDD extraction, the short-number formatting fix, NANPA leading-1 handling,
  and NDD-cleared-after-IDD-extraction.
- New implementation in `asyoutypeformatter.go` (the `AsYouTypeFormatter` type and
  the `GetAsYouTypeFormatter` constructor). It reuses the existing helpers
  (`extractCountryCode`, `formattingRuleHasFirstGroupOnly`,
  `normalizeDiallableCharsOnly`, region/country-code lookups, the `regexFor` cache,
  the `arabicIndicNumberals` digit map). A few unexported `java.lang.StringBuilder`-
  style methods were added to `stringbuilder.go` (`setLength`, `delete`,
  `substring`, `lastIndexOf`, `charAt`); `formattingTemplate` is held as a `[]rune`
  (not a `StringBuilder`) so the multi-byte `DIGIT_PLACEHOLDER` is indexed by
  character the way Java indexes by UTF-16 char. Position tracking
  (`GetRememberedPosition`) uses rune counts, which equal Java's UTF-16 char
  positions for every (BMP) character a phone number can contain — supplementary
  code points would differ, but phone input never contains them. The port required
  no behaviour fixes; it passed the upstream assertions as written.

### PhoneNumberMatcherTest — ported
- ✅ Faithful port in `phonenumbermatcher_test.go` (all 46 upstream test methods
  except `testRemovalNotSupported`, which has no Go analogue — range-over-func has
  no `remove()`), run against the synthetic test metadata via `useTestMetadata(t)`,
  reconciled against v9.0.32. Covers national/international finding, extensions,
  the `isLatinLetter` table, money/percentage/date/timestamp rejection, surrounding
  Latin/Chinese/punctuation contexts, all four leniency levels with the
  impossible/possible/valid/strict-grouping/exact-grouping case tables, bracket
  matching, max-tries bounding, and the iterator contract (idempotent `hasNext`,
  `next` after exhaustion).
- New implementation: `phonenumbermatch.go` (`PhoneNumberMatch` with
  `Number`/`Start`/`End`/`RawString`) and a full rewrite of `phonenumbermatcher.go`
  (the `find`/`extractMatch`/`extractInnerMatch`/`parseAndVerify` engine, the bounded
  `PATTERN` / `MATCHING_BRACKETS` / inner-match regexes, `isLatinLetter`,
  `checkNumberGroupingIsValid`, and faithful rewrites of the previously-stubbed/buggy
  grouping helpers). The public entry points are `FindNumbers(text, region)` and
  `FindNumbersWithLeniency(text, region, leniency, maxTries)`, returning a Go 1.23
  `iter.Seq[*PhoneNumberMatch]` (range with `for m := range …`). Match offsets are
  **byte** offsets (Go-idiomatic), so `text[m.Start():m.End()] == m.RawString()`;
  upstream uses UTF-16 char offsets, which differ only for multi-byte input.
- **Alternate-formats metadata** was added to support the grouping leniencies
  (`STRICT_GROUPING` / `EXACT_GROUPING`): `PhoneNumberAlternateFormats.xml` is now
  compiled (`BuildAlternateFormatsMetadataCollection`, a new `cmd/buildmetadata`
  step) into `data/alternateformats_metadata.xml.gz`, embedded, and looked up by
  country calling code (`getAlternateFormatsForCountryCallingCode`, loaded lazily).
  As upstream does, this uses the production alternate-formats data even though the
  matcher's main metadata is the synthetic test set.
- Regex notes: `REGEX_FLAGS` (UNICODE|CASE_INSENSITIVE) maps to RE2's already-Unicode
  `\p{…}` classes plus a `(?i)` prefix on `PATTERN`; Java's `.matches()` /
  `.lookingAt()` are emulated with anchored patterns / leftmost-at-0 index checks;
  `String.split(\D+)` is replicated (drop trailing empties) for the exact-grouping
  check.

### Bugs the port surfaced and fixed
- `extractPossibleNumber` never trimmed trailing junk: `UNWANTED_END_CHARS` was
  copied verbatim from Java (`[[\P{N}&&\P{L}]&&[^#]]+$`), but Go's RE2 engine has
  no character-class intersection (`&&`), so the pattern compiled to something
  meaningless and matched nothing. Rewrote it as the equivalent negated class
  `[^\p{N}\p{L}#]+$`. This also resolved the documented trailing-whitespace
  divergence in `testParseExtensions` (`"+44 2034567890 x 456  "` now extracts
  extension `456` like upstream, instead of folding the digits into the NSN).
- `IsNumberMatch` leading-zeros equality: the matcher compared the two numbers
  with `reflect.DeepEqual` on the raw protos instead of upstream's
  `copyCoreFieldsOnly` value comparison, so a `numberOfLeadingZeros` that was the
  proto default (1) on one side but unset on the other — or set at all when
  `italianLeadingZero` is false — wrongly demoted an `EXACT_MATCH` to
  `SHORT_NSN_MATCH`. Added `copyCoreFieldsOnly` and compare with `proto.Equal`.
- `IsNumberMatchWithOneNumber` nil-pointer panic: when the first number's country
  calling code maps to no region, the fallback branch passed a nil `*PhoneNumber`
  into `parseHelper` (which writes into an out-param, unlike upstream's
  return-value form). The branch had no test coverage until the port added the
  `randomNumber` (cc 41) case. Now allocates the proto first.
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
  `TestBuilderProcessesSmsServices`). The embedded short metadata has since been
  regenerated, so all 241 regions now carry `<smsServices>` data and the ported
  `TestIsSmsService` is active.

## Remaining work (roughly in order)

1. **Automate**: a scheduled task that detects new upstream releases, regenerates
   metadata, runs the (now-stable) synthetic tests, opens a PR for data-only
   deltas, and flags logic-touching changes for manual porting. See
   `docs/2.0-restructure.md`.

## Known TODOs / documented divergences

- **Embedded metadata regenerated against upstream (legacy `"NA"` sentinel gone).**
  `data/metadata.xml.gz` and `data/shortnumber_metadata.xml.gz` have been
  regenerated with the current builder: the main metadata now marks absent types
  with `possibleLength = [-1]` (0 `"NA"` national patterns remain, 1132 `[-1]`
  lengths across 245 regions), and the short metadata now carries `<smsServices>`
  for all 241 regions. The now-dead `"NA"` sentinel special-cases were removed from
  `descHasData` / `descHasPossibleNumberData` (both now match upstream verbatim),
  and the synthetic fixture in `metadatasource_test.go` was updated to mark absent
  types with `possibleLength = [-1]` like the real builder. The remaining `"NA"`
  literals (the builder's `intlFormat="NA"` handling and AsYouTypeFormatter's
  placeholder international prefix) are faithful upstream behavior, not the sentinel.
- **Parsing edge case:** `normalizeDigits` may not map every non-ASCII unicode
  digit script (e.g. Mongolian) to ASCII; the Arabic-Indic, Eastern-Arabic and
  fullwidth digits exercised by `TestNormaliseOtherDigits` do convert correctly.
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
