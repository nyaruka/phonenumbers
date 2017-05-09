gophone  [![Build Status](https://travis-ci.org/nyaruka/gophone.svg?branch=master)](https://travis-ci.org/nyaruka/gophone)
==============

golang port of Google's libphonenumber, forked from [libphonenumber from ttacon](https://github.com/ttacon/libphonenumber)

This will ultimately be a rewrite of ttacon's library, but for now mostly cleans up a few things and adds the `buildmetadata` cmd to 
allow for rebuilding the metadata protocol buffers, country code to region maps and timezone prefix maps.

API may change in the immediate future as we get this all buttoned up, but we depend on this heavily so we aim to have a fully
working port of the main library that is performant before too long.

Examples
========

```go
// parse our phone number
num, err := gophone.Parse("6502530000", "US")

// format it using national format
formattedNum := gophone.Format(num, gophone.NATIONAL)
```

Rebuilding Metadata and Maps
===============================

The `buildmetadata` command will fetch the latest XML file from the official Google repo and rebuild the go source files containing all
the territory metadata, timezone and region maps. It will rebuild the following files:

`metadata_bin.go` - contains the protocol buffer definitions for all the various formats across countries etc..

`countrycode_to_region.go` - contains a map built from the metadata to ease looking up possible regions for a country code

`prefix_to_timezone.go` - contains a map built from the timezone file within the Google repo mapping number prefixes to possible timezones

```bash
% go get github.com/nyaruka/gophone
% go install github.com/nyaruka/gophone/cmd/buildmetadata
% $GOPATH/bin/buildmetadata
```
