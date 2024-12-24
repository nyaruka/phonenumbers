# ☎️ phonenumbers 
[![Build Status](https://github.com/nyaruka/phonenumbers/workflows/CI/badge.svg)](https://github.com/nyaruka/phonenumbers/actions?query=workflow%3ACI) 
[![codecov](https://codecov.io/gh/nyaruka/phonenumbers/branch/main/graph/badge.svg)](https://codecov.io/gh/nyaruka/phonenumbers)
[![GoDoc](https://godoc.org/github.com/nyaruka/phonenumbers?status.svg)](https://godoc.org/github.com/nyaruka/phonenumbers)

golang port of Google's [libphonenumber](https://github.com/googlei18n/libphonenumber) forked from libphonenumber from [ttacon/libphonenumber](https://github.com/ttacon/libphonenumber). This library is used daily in production for parsing and validation of numbers across the world, so is well maintained. Please open an issue if you encounter any problems, we'll do our best to address them.

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

The `buildmetadata` command will fetch the latest XML file from the official Google repo and rebuild the go source files 
containing all the territory metadata, timezone and region maps.

It will rebuild the following files:

 * `gen/metadata_bin.go` - protocol buffer definitions for all the various formats across countries etc..
 * `gen/shortnumber_metadata_bin.go` - protocol buffer definitions for ShortNumberMetadata.xml
 * `gen/countrycode_to_region_bin.go` - information needed to map a country code to a region
 * `gen/prefix_to_carrier_bin.go` - information needed to map a phone number prefix to a carrier
 * `gen/prefix_to_geocoding_bin.go` - information needed to map a phone number prefix to a city or region
 * `gen/prefix_to_timezone_bin.go` - information needed to map a phone number prefix to a city or region

```bash
% go install github.com/nyaruka/phonenumbers/cmd/buildmetadata
% $GOPATH/bin/buildmetadata
```
