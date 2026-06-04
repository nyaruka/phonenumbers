# Upgrading to v2.x

`v2.0.0` is the first release after a substantial refactor whose goal was to make this a
**strict port** of Google's [libphonenumber](https://github.com/google/libphonenumber) Java
reference: the public API now mirrors the Java surface, lookup features are split into focused
subpackages, and a large amount of internal machinery that was previously exported has been
hidden.

This document lists every breaking change between **`v1.8.0`** (the last `1.x` release) and
**`v2.0.0`**, plus what's new and a migration checklist.

> The embedded metadata and parsing/validation **behaviour** are unchanged by this upgrade —
> the same numbers parse, validate, and format the same way. The breaking changes are all about
> the **shape of the API**, not results.

---

## 1. Import path and Go version

The module moved to the `/v2` path, as required for a Go major version:

```go
// before
import "github.com/nyaruka/phonenumbers"

// after
import "github.com/nyaruka/phonenumbers/v2"
```

```sh
go get github.com/nyaruka/phonenumbers/v2@v2.0.0
```

The package name is still `phonenumbers`, so unqualified references (`phonenumbers.Parse`, …)
don't change — only the import path does. The minimum Go version is now **1.24** (was 1.23),
largely because the matcher API uses range-over-func iterators (`iter.Seq`).

---

## 2. Carrier, geocoding and timezone moved to subpackages

These three lookups used to be top-level functions in the root package. They now live in
dedicated subpackages (`carrier`, `geocoding`, `timezone`) and their names match the upstream
Java mapper methods. All of them take a parsed `*phonenumbers.PhoneNumber`.

### Carrier — `github.com/nyaruka/phonenumbers/v2/carrier`

| v1.8.0 | v2.x |
|---|---|
| `phonenumbers.GetCarrierForNumber(num, lang)` | `carrier.GetNameForNumber(num, lang)` |
| `phonenumbers.GetSafeCarrierDisplayNameForNumber(num, lang)` | `carrier.GetSafeDisplayName(num, lang)` |
| `phonenumbers.GetCarrierWithPrefixForNumber(num, lang)` | **removed** — no upstream equivalent; use `carrier.GetNameForNumber` for the name |
| — | `carrier.GetNameForValidNumber(num, lang)` *(new — skips the internal validity check)* |

### Geocoding — `github.com/nyaruka/phonenumbers/v2/geocoding`

| v1.8.0 | v2.x |
|---|---|
| `phonenumbers.GetGeocodingForNumber(num, lang)` | `geocoding.GetDescriptionForNumber(num, lang)` |
| — | `geocoding.GetDescriptionForValidNumber(num, lang)` *(new)* |
| — | `geocoding.GetDescriptionForNumberForUserRegion(num, lang, userRegion)` *(new)* |
| — | `geocoding.GetDescriptionForValidNumberForUserRegion(num, lang, userRegion)` *(new)* |

> Note: country names are now rendered via `golang.org/x/text` rather than a `java.util.Locale`
> equivalent. Pass a language tag string (e.g. `"en"`, `"de"`) as `lang`.

### Timezone — `github.com/nyaruka/phonenumbers/v2/timezone`

| v1.8.0 | v2.x |
|---|---|
| `phonenumbers.GetTimezonesForNumber(num)` | `timezone.GetTimeZonesForNumber(num)` *(note the `TimeZones` casing)* |
| `phonenumbers.GetTimezonesForPrefix(numberString)` | `timezone.GetTimeZonesForGeographicalNumber(num)` — **signature changed**: now takes a parsed `*PhoneNumber`, not a string. `Parse` first. |
| const `phonenumbers.UNKNOWN_TIMEZONE` | `timezone.Unknown` |

**Example:**

```go
import (
    "github.com/nyaruka/phonenumbers/v2"
    "github.com/nyaruka/phonenumbers/v2/carrier"
    "github.com/nyaruka/phonenumbers/v2/timezone"
)

num, _ := phonenumbers.Parse("+12125550000", "US")
name, _ := carrier.GetNameForNumber(num, "en")
zones, _ := timezone.GetTimeZonesForNumber(num)
```

---

## 3. Number matcher reworked into iterators

The `1.x` `PhoneNumberMatcher` type / `NewPhoneNumberMatcher` constructor (which had no usable
exported methods) have been **removed**. Finding numbers in free text is now done with
range-over-func iterators that mirror upstream's `findNumbers`:

```go
// find valid numbers, no limit
for m := range phonenumbers.FindNumbers(text, "US") {
    fmt.Println(m.RawString(), m.Number())
}

// with explicit leniency and a cap on attempts
for m := range phonenumbers.FindNumbersWithLeniency(text, "US", phonenumbers.VALID, math.MaxInt) {
    _ = m
}
```

- `FindNumbers(text, defaultRegion) iter.Seq[*PhoneNumberMatch]`
- `FindNumbersWithLeniency(text, defaultRegion, leniency, maxTries) iter.Seq[*PhoneNumberMatch]`
- New `PhoneNumberMatch` type with `Number()`, `Start()`, `End()`, `RawString()`, `String()`.

> **`Start()` / `End()` are byte offsets** into `text` (so `text[m.Start():m.End()] == m.RawString()`),
> not UTF-16 offsets as in Java.

---

## 4. Removed / changed top-level functions

### `GetExampleNumberForType` signature changed

```go
// before — (regionCode, type)
phonenumbers.GetExampleNumberForType("US", phonenumbers.MOBILE)

// after — region-scoped overload was renamed; the bare name is now "any region"
phonenumbers.GetExampleNumberForTypeInRegion("US", phonenumbers.MOBILE)
phonenumbers.GetExampleNumberForType(phonenumbers.MOBILE) // new: first valid region
```

### `Builder` / `FormatWithBuf` removed

The exported `Builder` byte-buffer type and its constructors (`NewBuilder`, `NewBuilderString`)
and the `FormatWithBuf(num, format, *Builder)` formatting overload are gone (the buffer is now
`internal/stringbuilder`). Use the value-returning `Format`:

```go
s := phonenumbers.Format(num, phonenumbers.NATIONAL)
```

### Internal helpers no longer exported

These were never part of the intended public API (they're matcher/leniency/metadata internals)
and are now unexported or live under `internal/`. There is no public replacement:

`AllNumberGroupsAreExactlyPresent`, `AllNumberGroupsRemainGrouped`, `CheckNumberGroupingIsValid`,
`ContainsMoreThanOneSlashInNationalNumber`, `ContainsOnlyValidXChars`,
`IsNationalPrefixPresentIfRequired`, `MatchNationalNumber`, `MaybeSeparateExtensionFromPhone`,
`BuildCountryCodeToRegionMap`, `BuildPhoneMetadataCollection`.

---

## 5. Internal constants are no longer exported

`1.x` exported roughly 60 implementation-detail constants — regex patterns, character-class
strings, character/token mapping tables, length limits, and the `RFC3966_*` grammar fragments.
libphonenumber keeps all of these `private`; `v2.x` follows suit and unexports them. Examples:
`VALID_PUNCTUATION`, `DIGITS`, `VALID_ALPHA`, `PLUS_SIGN`, `STAR_SIGN`, `ALPHA_MAPPINGS`,
`ALL_PLUS_NUMBER_GROUPING_SYMBOLS`, `MOBILE_TOKEN_MAPPINGS`, `MIN_LENGTH_FOR_NSN`,
`MAX_LENGTH_FOR_NSN`, `MAX_INPUT_STRING_LENGTH`, `EXTN_PATTERN`, `VALID_PHONE_NUMBER_PATTERN`,
all `RFC3966_*`, `GEO_MOBILE_COUNTRIES`, `UNKNOWN_REGION`, and so on.

The only constant that remains exported is **`REGION_CODE_FOR_NON_GEO_ENTITY`** (`"001"`), which is
the sole `public` constant upstream.

If you depended on one of these (e.g. you reused `VALID_PUNCTUATION` to build your own regex),
you'll need to vendor your own copy of the value — they are no longer part of the API contract
and may change between releases to track upstream.

The exported **enum** constants are unaffected: `PhoneNumberFormat` (`E164`, `INTERNATIONAL`,
`NATIONAL`, `RFC3966`), `PhoneNumberType` (`FIXED_LINE`, `MOBILE`, …), `ValidationResult`,
`MatchType`, and `Leniency` are all still exported.

---

## 6. Metadata types and builder

The protobuf metadata value types now live in the `metadata` subpackage and are re-exported
from the root package as **type aliases**, so most code is unaffected:

```go
phonenumbers.PhoneMetadata           // = metadata.PhoneMetadata
phonenumbers.PhoneMetadataCollection // = metadata.PhoneMetadataCollection
phonenumbers.PhoneNumberDesc         // = metadata.PhoneNumberDesc
phonenumbers.NumberFormat            // = metadata.NumberFormat
```

`*phonenumbers.PhoneNumber` and all its `Get*` accessors are unchanged.

The XML-binding builder structs (`NumberFormatE`, `PhoneNumberDescE`, `PhoneNumberMetadataE`,
`PossibleLengthE`, `TerritoryE`) and the `Default_*` proto default constants are **no longer
exported** — they moved into `internal/metadatabuilder`. Regenerating embedded metadata is done
via `go run ./cmd/buildmetadata`.

`MetadataCollection()` and `ShortNumberMetadataCollection()` remain available in the root package.

---

## 7. Removed error variables

The `Builder`/buffer-related error values are gone along with the `Builder` type:
`ErrFailedToGrow`, `ErrInvalidIndex`, `ErrTooLarge`.

The parsing/validation errors are unchanged: `ErrEmptyMetadata`, `ErrInvalidCountryCode`,
`ErrNotANumber`, `ErrNumTooLong`, `ErrTooShortAfterIDD`, `ErrTooShortNSN`.

---

## What's new in 2.x

Beyond the restructuring, `v2.0.0` adds public API that was missing in `1.x`, closing gaps with
upstream:

- **As-you-type formatter** — `phonenumbers.GetAsYouTypeFormatter(region)` returning an
  `*AsYouTypeFormatter` with `InputDigit`, `InputDigitAndRememberPosition`, `GetRememberedPosition`,
  `Clear`.
- **Short-number cost** — `GetExpectedCost`, `GetExpectedCostForRegion`, and the `ShortNumberCost`
  type (`TOLL_FREE_COST`, `STANDARD_RATE_COST`, `PREMIUM_RATE_COST`, `UNKNOWN_COST`).
- **More short-number / validation helpers** — `IsCarrierSpecific`, `IsCarrierSpecificForRegion`,
  `IsSmsServiceForRegion`, `IsPossibleNumberForType`, `IsPossibleNumberForTypeWithReason`,
  `IsPossibleNumberFromRegion`, `IsNumberGeographical`, `IsNumberGeographicalForType`,
  `CanBeInternationallyDialled`, `GetSupportedTypesForRegion`, `GetSupportedTypesForNonGeoEntity`,
  `GetInvalidExampleNumber`, `NormalizeDiallableCharsOnly`.
- **Metadata injection** — the `metadata` subpackage exposes `Use`/`Container` for swapping the
  metadata source (useful in tests).

---

## Migration checklist

1. Update the import path to `.../v2` and run `go get github.com/nyaruka/phonenumbers/v2@v2.0.0`.
2. Ensure your toolchain is Go 1.24+.
3. Move carrier/geocoding/timezone calls to the new subpackages and rename them (§2).
4. Replace any `NewPhoneNumberMatcher` usage with the `FindNumbers` iterators (§3).
5. Replace `GetExampleNumberForType(region, type)` with `GetExampleNumberForTypeInRegion` (§4).
6. Replace `FormatWithBuf` / `Builder` usage with `Format` (§4).
7. Inline any internal constants you relied on (§5).
8. Drop references to the removed `Builder` error values (§7).
9. `go build ./...` — the compiler will flag every remaining call site.
