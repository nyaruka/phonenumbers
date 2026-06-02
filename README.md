# ☎️ phonenumbers 
[![Build Status](https://github.com/nyaruka/phonenumbers/workflows/CI/badge.svg)](https://github.com/nyaruka/phonenumbers/actions?query=workflow%3ACI) 
[![codecov](https://codecov.io/gh/nyaruka/phonenumbers/branch/main/graph/badge.svg)](https://codecov.io/gh/nyaruka/phonenumbers)
[![Go Reference](https://pkg.go.dev/badge/github.com/nyaruka/phonenumbers.svg)](https://pkg.go.dev/github.com/nyaruka/phonenumbers)

Go port of Google's [libphonenumber](https://github.com/google/libphonenumber), forked from [ttacon/libphonenumber](https://github.com/ttacon/libphonenumber). Specifically it tracks the **Java** implementation — libphonenumber's reference implementation, of which the C++ and JavaScript versions are themselves ports — so type names, method names, and tests mirror Java's `PhoneNumberUtil` and `PhoneNumberUtilTest`. This library is used daily in production for parsing and validation of numbers worldwide and is well maintained. Please open an issue if you encounter any problems, and we'll do our best to address them.

> [!IMPORTANT]
> The aim of this project is strictly to be a port and match as closely as possible the functionality in libphonenumber. Please don't submit feature requests for functionality that doesn't exist in libphonenumber.

> [!IMPORTANT]
> We use the metadata from libphonenumber so if you encounter unexpected parsing results, please first verify if the problem affects libphonenumber and report there if so. You can use their [online demo](https://libphonenumber.appspot.com) to quickly check parsing results.

## Version Numbers

As we don't want to bump our major semantic version number in step with the upstream library, we use independent version numbers than the Google libphonenumber repo. The release notes will mention what version of the metadata a release was built against.

## Planned for v2

A future v2 (published under a `/v2` module path) will tidy up some legacy API surface. These are breaking changes, so they won't land on v1 — current consumers are unaffected until they choose to upgrade:

- Remove the deprecated `FormatWithBuf`. Use `Format` instead; the buffer reuse it offered is negligible in practice.
- Make the `StringBuilder` type private once nothing in the public API depends on it (after `FormatWithBuf` is gone). It only exists to mirror Java's mutable-`StringBuilder` formatting helper and isn't meant for external use.

## Usage

```go
// parse our phone number
num, err := phonenumbers.Parse("6502530000", "US")

// format it using national format
formattedNum := phonenumbers.Format(num, phonenumbers.NATIONAL)
```

## Updating Metadata

The `buildmetadata` command clones a pinned upstream libphonenumber release (set by the
`upstreamVersion` constant in `cmd/buildmetadata/main.go`) and rebuilds the embedded
metadata into the gzipped files under `data/`:

 * `data/metadata.xml.gz` - core territory metadata (number formats, validation rules, etc.)
 * `data/shortnumber_metadata.xml.gz` - short-number metadata
 * `data/alternateformats_metadata.xml.gz` - alternate format patterns used when matching
 * `data/countrycode_to_region.xml.gz` - maps a country code to its region(s)
 * `data/prefix_to_carriers/*.gz` - maps a phone number prefix to a carrier
 * `data/prefix_to_geocodings/*.gz` - maps a phone number prefix to a geographic area
 * `data/prefix_to_timezone.xml.gz` - maps a phone number prefix to a timezone

```bash
% go run ./cmd/buildmetadata
```

To sync to a newer upstream release, bump `upstreamVersion`, re-run the command, run the
tests, then update [SYNC.md](SYNC.md) — which records the upstream version each part of
the port is currently reconciled against.
