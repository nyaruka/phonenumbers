---
name: sync-upstream
description: Sync this Go port with a new upstream google/libphonenumber release — regenerate the embedded metadata and reconcile the ported Java logic. Use when the user says "sync with upstream", "do an upstream sync", "reconcile against libphonenumber", "pull the latest libphonenumber", bump to a new libphonenumber version, or invokes /sync-upstream. SYNC.md holds the state (baseline version + deliberate divergences); this skill is the procedure.
---

# Sync with upstream libphonenumber

This package is a Go port of [google/libphonenumber](https://github.com/google/libphonenumber),
tracking its **Java** reference implementation. A sync is **two independent operations** —
do them in order, but understand they're different in kind:

1. **Metadata regen** — *mechanical.* Rebuild the embedded `data/` blobs from a new upstream
   release. Fully automated by `cmd/buildmetadata`.
2. **Code reconciliation** — *judgment.* Port new Java *logic* changes into the Go files. Can't
   be automated; this is the real reason [SYNC.md](../../../SYNC.md) exists.

State lives in [SYNC.md](../../../SYNC.md): the **code baseline** (`Code reconciled against vX.Y.Z`)
and the **deliberate divergences** (what we intentionally do *not* port). Read it first. The
embedded-metadata version is tracked separately in the generated `metadata/version.go`
(`metadata.Version`) — never hand-edit that.

Run everything from the repo root (`/Users/rowan/nyaruka/phonenumbers`).

---

## Phase 1 — Metadata regen (mechanical)

```sh
go run ./cmd/buildmetadata            # latest upstream release
# go run ./cmd/buildmetadata v9.0.32  # or pin a specific tag
```

This resolves the target tag, wipes and re-clones upstream into `_build/` (a shallow
`--depth=1` checkout **at the target tag** — gitignored), rebuilds every embedded blob
(`data/`, `metadata/data/`, `carrier/data/`, `geocoding/data/`, `timezone/data/`), and
rewrites the generated `metadata/version.go` with the tag it built from.

Then:

```sh
go test ./...
```

Review the diff in `data/`, `metadata/data/`, the per-package `*/data/` blobs, and
`metadata/version.go`. A metadata-only release (no Java logic change) can stop here: Phase 1
has already bumped `metadata/version.go`, and the **`Code reconciled against`** baseline in
SYNC.md stays where it is (the two versions are tracked separately for exactly this case).

> The target tag you built becomes the new code baseline once Phase 2 reconciles against it.

---

## Phase 2 — Code reconciliation (judgment)

Goal: apply any upstream Java *logic* changes between the **old baseline** (SYNC.md) and the
**target** (the tag Phase 1 built) to each Go port.

### Get both tags side by side for diffing

Phase 1 left `_build/` as a shallow clone at the **target** tag. Fetch the **baseline** tag
into it so you can diff tree-to-tree (no merge base needed — `git diff A B` compares trees):

```sh
BASE=v9.0.31      # from SYNC.md "Code reconciled against"
TARGET=v9.0.32    # the tag Phase 1 built (see metadata/version.go)
git -C _build fetch --depth=1 origin tag "$BASE"
```

### Diff each ported source, baseline→target

The per-file headers are the source of truth for the Go→Java mapping — every reconcilable
file carries a `// Port of <upstream/path>.` header. Enumerate them:

```sh
grep -rn "Port of" --include='*.go' . | grep -v _test.go | grep -v _build/
```

For each upstream path `P`, diff it across the window and reconcile anything non-trivial:

```sh
git -C _build diff "$BASE" "$TARGET" -- "$P"
```

The mapping (kept current; **if files move, the headers win** and this table is refreshed):

| Go | upstream Java path |
|---|---|
| `phonenumberutil.go`, `enums.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberUtil.java` |
| `asyoutypeformatter.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/AsYouTypeFormatter.java` |
| `shortnumberinfo.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/ShortNumberInfo.java` |
| `phonenumbermatcher.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberMatcher.java` |
| `phonenumbermatch.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberMatch.java` |
| `errors.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/NumberParseException.java` |
| `internal/regexcache/` | `java/libphonenumber/src/com/google/i18n/phonenumbers/internal/RegexCache.java` |
| `internal/regexbasedmatcher/` | `java/libphonenumber/src/com/google/i18n/phonenumbers/internal/RegexBasedMatcher.java`, `.../internal/MatcherApi.java` |
| `metadata/source.go`, `alternateformats.go` | `java/libphonenumber/src/com/google/i18n/phonenumbers/metadata/source/*`, `.../MetadataLoader.java` |
| `carrier/` | `java/carrier/src/com/google/i18n/phonenumbers/PhoneNumberToCarrierMapper.java` |
| `geocoding/` | `java/geocoder/src/com/google/i18n/phonenumbers/geocoding/PhoneNumberOfflineGeocoder.java` |
| `timezone/` | `java/geocoder/src/com/google/i18n/phonenumbers/PhoneNumberToTimeZonesMapper.java` |
| `internal/prefixmapper/` | `java/internal/prefixmapper/src/com/google/i18n/phonenumbers/prefixmapper/*` |
| `internal/metadatabuilder/` | `tools/java/common/src/com/google/i18n/phonenumbers/BuildMetadataFromXml.java` |

`internal/serialize` and `internal/stringbuilder` are Go-specific scaffolding (gzip/binary
decode; a `java.lang.StringBuilder` stand-in) with no upstream logic to reconcile — skip them.

The corresponding **upstream tests** move in lockstep. The Go tests mirror the Java test
classes (e.g. `PhoneNumberUtilTest` → `phonenumberutil_test.go` and friends); diff those too
and port new cases:

```sh
git -C _build diff "$BASE" "$TARGET" -- "$(echo "$P" | sed 's#/src/#/test/#')"
```

### Reconcile

For each non-empty diff, translate the relevant change into the Go port. While doing so:

- **Preserve upstream source order.** The core ports keep their functions/methods in the same
  order as the upstream Java so the *next* sync's diff lines up — maintain that as you reconcile.
  A few files are deliberately reorganized for Go idiom and don't follow upstream order
  (`metadata/source.go`, `internal/regexcache`, `internal/regexbasedmatcher`,
  `internal/prefixmapper`); that's expected.
- **Java-shaped over idiomatic — in ports.** Inside a `// Port of …` file, keep constructs close
  to the Java so the next diff lines up, even when a linter (`gopls modernize`, `staticcheck`)
  suggests a more idiomatic Go form — e.g. don't rewrite an index `for` loop as `for i := range n`,
  or a manual scan as `slices.Contains`. Let those modernizations ride in *with* upstream when the
  Java itself changes. Non-ported code (`cmd/`, CI, build glue, and Go-specific scaffolding like
  `internal/serialize` / `internal/stringbuilder`) has no Java counterpart to track, so make it
  fully idiomatic.
- **Honor deliberate divergences.** Check SYNC.md before porting something back; some upstream
  code is intentionally absent (e.g. Mockito-style metadata-source injection tests, Java
  iterator semantics with no Go equivalent). Don't re-introduce it. If you make a *new*
  intentional divergence, record it.
- **Keep it a port, not an enhancement.** Match libphonenumber's behaviour; don't add API.
- **Mirror Java's visibility.** A symbol that is `public` in Java is exported in Go; one that is
  `private`/package-private/`@VisibleForTesting` stays unexported (lower-case camelCase). Constants
  especially: `REGION_CODE_FOR_NON_GEO_ENTITY` is the *only* `public static final` upstream, so it
  is the lone exported constant — every other ported regex/char-class/length-limit/mapping is
  unexported. The deliberate exceptions are the Go-idiomatic `Err*` vars (standing in for Java's
  checked exceptions) and the public enum types. Don't export a new internal just because it's a
  package-level `const`/`var`.
- **Java overloads → distinct Go names.** Go can't overload, so a set of Java methods sharing a
  name maps to several Go funcs: give one the bare PascalCase name and suffix the rest with the
  parameter(s) that distinguish them. Follow the established pattern — `Parse`/`ParseToNumber`,
  `IsNumberMatch`/`IsNumberMatchWithNumbers`/`IsNumberMatchWithOneNumber`,
  `GetExampleNumberForType`/`GetExampleNumberForTypeInRegion`,
  `IsPossibleNumber`/`IsPossibleNumberFromRegion`, `IsNumberGeographical`/`IsNumberGeographicalForType`.
  Porting an overload that *already exists* upstream is faithful porting, not the "don't add API"
  case above — fill the gap rather than skipping it. Which overload keeps the bare name (and the
  exact suffix) is a judgment call; flag new public names in the PR.
- **Don't hand-edit generated files** (`metadata/version.go`, `*.pb.go`) or the `data/` blobs —
  those come from Phase 1 / protoc.

Then make it green:

```sh
go test ./...
```

---

## Finalize

Update [SYNC.md](../../../SYNC.md) (state only):

1. Bump the **`Code reconciled against`** version to the target tag.
2. If anything new was intentionally left unported, add it under **Deliberate divergences**.

Record what was reconciled (and the version) in the commit / PR message — git history is the
sync log, so SYNC.md doesn't keep one. Don't touch `CHANGELOG.md` either; it's maintained by
the release process.

Leave public-facing text (commit messages, comments) describing the software in general terms.
