# ☎️ phonenumbers 
[![Build Status](https://github.com/nyaruka/phonenumbers/workflows/CI/badge.svg)](https://github.com/nyaruka/phonenumbers/actions?query=workflow%3ACI) 
[![Go Reference](https://pkg.go.dev/badge/github.com/nyaruka/phonenumbers/v2.svg)](https://pkg.go.dev/github.com/nyaruka/phonenumbers/v2)

Go port of Google's [libphonenumber](https://github.com/google/libphonenumber). Specifically it tracks the **Java** implementation — that library's reference implementation, of which the C++ and JavaScript versions are themselves ports. This library is used daily in production for parsing and validation of numbers worldwide and is well maintained.

> [!IMPORTANT]
> This project is a strict port of libphonenumber. Please do not submit feature requests for functionality that doesn't exist there. It also uses the metadata from libphonenumber, so if you encounter unexpected parsing results, please first verify if the problem affects libphonenumber and report there if so. You can use the [online demo](https://libphonenumber.appspot.com) to quickly check parsing results.

> [!TIP]
> See [`UPGRADING.md`](UPGRADING.md) for details on upgrading from 1.x to 2.x.

## Usage

```go
import "github.com/nyaruka/phonenumbers/v2"

// parse our phone number
num, err := phonenumbers.Parse("8886418722", "US")

// format it using national format
formatted := phonenumbers.Format(num, phonenumbers.NATIONAL)
```

## Updating

### Metadata

The `buildmetadata` command regenerates the embedded metadata from an upstream release. It
clones the release and rebuilds the gzipped data blobs embedded across the packages (core
territory metadata, short numbers, alternate formats, country-code→region, carrier,
geocoding, and timezone), recording the release it built from in the generated
[`metadata/version.go`](metadata/version.go).

```bash
# resolve and build from the latest upstream release
% go run ./cmd/buildmetadata

# or pin a specific release tag
% go run ./cmd/buildmetadata v9.0.31
```

This part is mechanical and fully automated.

### Code

Regenerating the metadata is only half of a sync — the ported Go code also has to be
reconciled against the new release's Java *logic* changes, which is judgment work rather than
a mechanical rebuild.

If you use [Claude Code](https://claude.com/claude-code), the **`sync-upstream`** skill
([`.claude/skills/sync-upstream/SKILL.md`](.claude/skills/sync-upstream/SKILL.md)) walks
through the whole process — metadata regen plus the per-file code reconciliation. Invoke it
with `/sync-upstream` or by asking it to sync with upstream. [SYNC.md](SYNC.md) records the
upstream version the code is reconciled against and the deliberate divergences from upstream.
