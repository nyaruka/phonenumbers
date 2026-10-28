# phonenumbers

[![Build Status](https://github.com/Accompany-Health/phonenumbers/workflows/CI/badge.svg)](https://github.com/Accompany-Health/phonenumbers/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/Accompany-Health/phonenumbers/branch/main/graph/badge.svg)](https://codecov.io/gh/Accompany-Health/phonenumbers)
[![Go Reference](https://pkg.go.dev/badge/github.com/Accompany-Health/phonenumbers.svg)](https://pkg.go.dev/github.com/Accompany-Health/phonenumbers)

## What this repo is

Go port of Google's [libphonenumber](https://github.com/google/libphonenumber), forked from [nyaruka/phonenumbers](https://github.com/nyaruka/phonenumbers). Used across Accompany Health backend services to parse, format, and validate phone numbers for all world regions. This library supersedes the older `libphonenumber` repo.

## What this owns

- Phone number parsing from free-text and structured input
- Formatting in E.164, INTERNATIONAL, NATIONAL, and RFC3966 formats
- Validation (globally and per-region)
- Number type detection (FIXED_LINE, MOBILE, TOLL_FREE, VOIP, etc.)
- Short number lookup (emergency services, info lines)

No runtime infrastructure — this is a pure Go library with no database, network, or service dependencies.

## Related systems

**Upstream (code):** [nyaruka/phonenumbers](https://github.com/nyaruka/phonenumbers) — the Go fork this repo tracks for implementation changes.

**Upstream (metadata):** [google/libphonenumber](https://github.com/google/libphonenumber) — source of territory formatting/validation XML, refreshed via `cmd/buildmetadata`.

**Downstream:** Any Accompany Health service that imports `github.com/Accompany-Health/phonenumbers`.

## Stack

- **Language:** Go 1.23
- **Serialization:** Protocol Buffers (`google.golang.org/protobuf`)
- **Metadata embedding:** Go `//go:embed` — gzipped XML/text files in `data/`

## Local Development

See [docs/local-development.md](docs/local-development.md) for setup, testing, linting, metadata refresh, and upstream sync instructions.

## Environment variables

None. This is a pure library with no runtime configuration.

## Where key things live

| Path | Purpose |
|---|---|
| `phonenumbers.go` | Public API — Parse, Format, IsValidNumber, GetNumberType, etc. |
| `shortnumber_info.go` | Short number (emergency/info service) validation |
| `data/` | Embedded gzipped metadata files (territory rules, carrier/geo/timezone maps) |
| `metadata.go` | `go:embed` directives wiring `data/` into the binary |
| `cmd/buildmetadata/` | CLI tool to refresh `data/` from upstream Google libphonenumber |

## Deployment

This is a versioned Go module. Consuming services import it via `go.mod`:

```bash
go get github.com/Accompany-Health/phonenumbers@v1.x.y
```

Releases are tag-triggered — pushing a `v1.x.y` tag runs `.goreleaser.yml` via CI and creates a GitHub release. Consuming services should pin a specific version tag.

## Troubleshooting

See [docs/troubleshooting.md](docs/troubleshooting.md) for common problems including unexpected parsing results, stale metadata, and build failures.

## Ownership

- **Team:** ninjas
- **Slack:** #eng-backend <!-- TODO: confirm actual Slack channel for ninjas -->
- **CODEOWNERS:** `@Accompany-Health/ninjas`

---

> [!IMPORTANT]
> The aim of this project is strictly to be a port and match as closely as possible the functionality in libphonenumber. Please don't submit feature requests for functionality that doesn't exist in libphonenumber.

> [!IMPORTANT]
> We use the metadata from libphonenumber so if you encounter unexpected parsing results, please first verify if the problem affects libphonenumber and report there if so. You can use their [online demo](https://libphonenumber.appspot.com) to quickly check parsing results.

---

## Version Numbers

As we don't want to bump our major semantic version number in step with the upstream library, we use independent version numbers from the Google libphonenumber repo. Release notes will mention what version of the metadata a release was built against.

---

## Usage

### Installation

```bash
go get github.com/Accompany-Health/phonenumbers
```

```go
import "github.com/Accompany-Health/phonenumbers"
```

The library is safe for concurrent use. Metadata is loaded lazily on first use and cached; there is no mutable shared state after initialization.

### Parsing and formatting

```go
// Parse a phone number (region hint used when no country code is present)
num, err := phonenumbers.Parse("6502530000", "US")

// Format in various styles
phonenumbers.Format(num, phonenumbers.E164)          // "+16502530000"
phonenumbers.Format(num, phonenumbers.INTERNATIONAL) // "+1 650-253-0000"
phonenumbers.Format(num, phonenumbers.NATIONAL)      // "(650) 253-0000"
phonenumbers.Format(num, phonenumbers.RFC3966)       // "tel:+16502530000"
```

### Validation

```go
phonenumbers.IsValidNumber(num)                    // true/false — globally valid
phonenumbers.IsValidNumberForRegion(num, "US")     // true/false — valid for region
phonenumbers.IsPossibleNumber(num)                 // true/false — plausible length
```

### Number type and region

```go
phonenumbers.GetNumberType(num)           // FIXED_LINE, MOBILE, TOLL_FREE, VOIP, ...
phonenumbers.GetRegionCodeForNumber(num)  // "US", "GB", "IN", ...
```

### Short numbers (emergency / info services)

```go
sni := phonenumbers.NewShortNumberInfo()
sni.IsEmergencyNumber("911", "US")        // true
sni.IsValidShortNumberForRegion(num, "US")
```

---

## Updating Metadata

The `buildmetadata` command fetches the latest files from the official Google repo and rebuilds the gzipped metadata in `data/`:

```bash
go install github.com/Accompany-Health/phonenumbers/cmd/buildmetadata
$GOPATH/bin/buildmetadata
```

It will update the following files:

- `data/metadata.xml.gz` — territory formatting and validation rules
- `data/shortnumber_metadata.xml.gz` — short number metadata
- `data/countrycode_to_region.xml.gz` — country code → region mappings
- `data/prefix_to_timezone.xml.gz` — number prefix → timezone
- `data/prefix_to_carriers/` — number prefix → carrier (per locale)
- `data/prefix_to_geocodings/` — number prefix → city/region (per locale)

Commit the updated files and note the upstream metadata version in `CHANGELOG.md`.

---

## Architecture

See [docs/architecture.md](docs/architecture.md) for a detailed breakdown of package structure, the public API surface, and how metadata is embedded and loaded at runtime.
