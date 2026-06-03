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

## Usage

```go
// parse our phone number
num, err := phonenumbers.Parse("6502530000", "US")

// format it using national format
formattedNum := phonenumbers.Format(num, phonenumbers.NATIONAL)
```

## Updating Metadata

The `buildmetadata` command clones an upstream libphonenumber release and rebuilds the
embedded metadata into gzipped files, each embedded by the package that uses it:

 * `metadata/data/metadata.xml.gz` - core territory metadata (number formats, validation rules, etc.)
 * `data/shortnumber_metadata.xml.gz` - short-number metadata
 * `data/alternateformats_metadata.xml.gz` - alternate format patterns used when matching
 * `metadata/data/countrycode_to_region.xml.gz` - maps a country code to its region(s)
 * `carrier/data/*.gz` - maps a phone number prefix to a carrier
 * `geocoding/data/*.gz` - maps a phone number prefix to a geographic area
 * `timezone/data/prefix_to_timezone.xml.gz` - maps a phone number prefix to a timezone

By default it resolves the **latest** upstream release tag, rebuilds `data/`, and records
the release it used in the generated `metadata/version.go` (the `metadata.Version`
constant):

```bash
% go run ./cmd/buildmetadata
```

To rebuild from a specific release instead, pass the tag:

```bash
% go run ./cmd/buildmetadata v9.0.31
```

After syncing, run the tests and update [SYNC.md](SYNC.md) — which records the upstream
version the port is reconciled against and the deliberate divergences from upstream.

Regenerating the metadata is only half of a sync; the ported Go code also has to be
reconciled against the new release's Java changes. If you use [Claude Code](https://claude.com/claude-code),
the **`sync-upstream`** skill ([`.claude/skills/sync-upstream/SKILL.md`](.claude/skills/sync-upstream/SKILL.md))
walks through the whole process — metadata regen plus the per-file code reconciliation — and
updates [SYNC.md](SYNC.md) at the end. Invoke it with `/sync-upstream` or by asking it to
sync with upstream.
