phonenumbers  [![Build Status](https://travis-ci.org/nyaruka/phonenumbers.svg?branch=master)](https://travis-ci.org/nyaruka/phonenumbers) 
[![GoDoc](https://godoc.org/github.com/nyaruka/phonenumbers?status.svg)](https://godoc.org/github.com/nyaruka/phonenumbers)
==============

golang port of Google's libphonenumber, forked from [libphonenumber from ttacon](https://github.com/ttacon/libphonenumber) which in turn is a port of the original [Java library](https://github.com/googlei18n/libphonenumber/tree/master/java/libphonenumber/src/com/google/i18n/phonenumbers).

You can play with a recent version of this library and how it parses numbers with the [phoneserver](http://phoneserver.nyaruka.com).

This fork cleans up a few things, fixes some bugs and adds the `buildmetadata` cmd to allow for rebuilding the metadata protocol buffers, country code to region maps and timezone prefix maps. We also keep this library up to date with the upstream Google repo as metadata changes take place.

This library is used daily in production for parsing and validation of numbers across the world, so is well maintained. Please open an issue if you encounter any problems.

Version Numbers
=======

As we don't want to bump our major semantic version number in step with the upstream library (especially in light of vgo), we use independent version numbers than the Google libphonenumber repo.

Usage
========

```go
// parse our phone number
num, err := phonenumbers.Parse("6502530000", "US")

// format it using national format
formattedNum := phonenumbers.Format(num, phonenumbers.NATIONAL)
```

Rebuilding Metadata and Maps
===============================

The `buildmetadata` command will fetch the latest XML file from the official Google repo and rebuild the go source files containing all
the territory metadata, timezone and region maps. It will rebuild the following files:

`metadata_bin.go` - contains the protocol buffer definitions for all the various formats across countries etc..

`countrycode_to_region.go` - contains a map built from the metadata to ease looking up possible regions for a country code

`prefix_to_timezone.go` - contains a map built from the timezone file within the Google repo mapping number prefixes to possible timezones

```bash
% go get github.com/nyaruka/phonenumbers
% go install github.com/nyaruka/phonenumbers/cmd/buildmetadata
% $GOPATH/bin/buildmetadata
```
