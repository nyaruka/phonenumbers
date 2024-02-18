v1.3.2 (2024-02-18)
-------------------------
 * Update metadata

v1.3.1 (2024-01-26)
-------------------------
 * Update metadata

v1.3.0 (2023-12-14)
-------------------------
 * Refactor buildmetadata, no longer requires SVN
 * Update metadata

v1.2.3 (2023-12-12)
-------------------------
 * Update metadata
 * Update dependencies and go version to 1.19

v1.2.2 (2023-11-24)
-------------------------
 * Update metadata
 * Update support for phone-context

v1.2.1 (2023-11-20)
-------------------------
 * Update metadata
 * Replace github.com/golang/protobuf with google.golang.org/protobuf

v1.2.0 (2023-11-17)
-------------------------
 * Update metadata
 * Fix regex matching in GetLengthOfNationalDestinationCode
 * Implement carrier GetSafeDisplayName

v1.1.9 (2023-11-08)
-------------------------
 * Update metadata

v1.1.8 (2023-08-09)
-------------------------
 * Update metadata

v1.1.7 (2023-05-10)
-------------------------
 * Merge pull request #137 from nyaruka/updates
 * Add isValid display to phoneparser
 * Update metadata
 * Merge pull request #133 from nyaruka/dependabot/go_modules/golang.org/x/text-0.3.8
 * Merge pull request #134 from nyaruka/dependabot/go_modules/cmd/phoneparser/golang.org/x/text-0.3.8
 * Merge pull request #135 from nyaruka/dependabot/go_modules/cmd/phoneserver/golang.org/x/text-0.3.8
 * Bump golang.org/x/text from 0.3.7 to 0.3.8 in /cmd/phoneserver
 * Bump golang.org/x/text from 0.3.7 to 0.3.8 in /cmd/phoneparser
 * Bump golang.org/x/text from 0.3.7 to 0.3.8

v1.1.6 (2023-02-13)
-------------------------
 * Update metadata

v1.1.5 (2023-01-27)
-------------------------
 * Update metadata

v1.1.4 (2022-11-28)
-------------------------
 * Bump required go version to 1.18

v1.1.3 (2022-11-28)
-------------------------
 * Update metadata

v1.1.2
----------
 * Update metadata
 * Fix slice out of bounds in GetTimezonesForPrefix

v1.1.1
----------
 * Update metadata

v1.1.0
----------
 * Update to latest metadata
 * Port initial short number support

v1.0.75
----------
 * Cleanup some of the unit tests using testify library
 * Update metadata and add test for new 0326 PK numbers

v1.0.74
----------
 * Update to latest metadata

v1.0.73
----------
 * Added fallback to region for GetGeocodingForNumber

v1.0.72
----------
 * Update metadata to v8.12.33

v1.0.71
----------
 * Update metadata to v8.12.31

v1.0.70
----------
 * Update metadata to v8.12.24

v1.0.69
----------
 * update metadata to 8.12.22
 * update test case for AR formatting

v1.0.68
----------
 * Add GetCarrierWithPrefixForNumber (thanks @RaMin0)

v1.0.67
----------
 * Update metadata (tracking 8.12.19 upstream)

v1.0.66
----------
 * Updated metadata

v1.0.65
----------
 * Add exported IsNumberMatchWithNumbers and IsNumberMatchWithOneNumber (thanks @akurth)

v1.0.64
----------
 * test goreleaser config

v1.0.63
----------
 * test goreleaser

v1.0.62
----------
 * Fix country code parsing
 * Update metadata

v1.0.61
----------
 * Update metadata
 * Add MaybeSeparatePhoneFromExtension helper function (thanks @richard-rance)

v1.0.60
----------
 * update metadata
 * better error logging in buildmetadata
 * update CI worflow (thanks @cristaloleg)
 * fix maybeExtractCountryCode regexp func (thanks @cristaloleg)

v1.0.59
----------
 * update to latest metadata

v1.0.58
----------
 * Update metadata to version v8.12.11

v1.0.57
----------
 * fix panic in IsNumberMatch() 

v1.0.56
----------
 * Update to metadata v8.12.5
 * Update test for Sydney tz (validated against source data)

v1.0.55
----------
 * Update metadata to v8.12.1 for upstream project

v1.0.54
----------
 * update metadata for v8.11.0

v1.0.53
----------
 * Metadata update for upstream v8.10.23

v1.0.52
----------
 * Reset italian leading zero when false, fixed bug when phonenumber struct is reused

v1.0.51
----------
 * Update metadata to upstream 8.10.21

v1.0.50
----------
 * Fix formatting of country code in out-of-country format (thanks @janh)
 * Fix FormatInOriginalFormat for numbers with national prefix (thanks @janh)
 * Fix panic due to calling proto.Merge on nil destination (thanks @janh)

v1.0.49
----------
 * fix Makefile for phoneserver

v1.0.48
----------
 * another test travis rev, ignore

v1.0.47
----------
 * test tag for travis deploy

v1.0.46
----------
 * update metadata for v8.10.19
 * remove aws-lambda-go as dependency (thanks @shaxbee)

v1.0.45
----------
 * Update metadata to fix Mexican formatting (thanks @bvisness)
 * Add tests specifically for Mexico (thanks @bvisness)

v1.0.44
----------
 * update metadata for v8.10.16
 * upgrade to the latest release of protobuf

v1.0.43
----------
 * Update metadata for v8.10.14

v1.0.42
----------
 * Update for metadata changes in v8.10.13
 * fix yoda expressions
 * fix slice operations
 * fix regex escaping
 * fix make calls
 * fix error strings

v1.0.41
----------
 * update metadata for v8.10.12

v1.0.40
----------
 * add unit test for valid/possible US/CA number, include commit in netlify version, lastest metadata
 * update readme to add svn dependency

v1.0.39
----------
 * add dist to gitignore
 * tweak goreleaser

v1.0.38
----------
 * update travis env to always enable modules

v1.0.37
----------
 * plug in goreleaser and add it to travis

v1.0.36
----------
 * Update for upstream metadata v8.10.7

v1.0.35
----------
 * update metadata for v8.10.4 release
 * update AR test number to valid AR fixed line

v1.0.34
----------
 * update travis file

v1.0.33
----------
 * remove goreleaser since we no longer use docker for test deploys
 * latest google metadata

v1.0.32
----------
 * add /functions to gitignore
 * update to latest google metadata

v1.0.31
----------
 * update to latest metadata v8.10.1, test case changes validated against google lib
 * add link in readme to test function

v1.0.30
----------
 * fix FormatByPattern with user defined pattern. Fixes: #16

v1.0.29
----------
 * update metadata v8.9.16 (test diff validated against python lib)

v1.0.28
----------
 * update metadata to v8.9.14, fix go.mod dependency

v1.0.27
----------
 * update to metadata v8.9.13, remove must dependency

v1.0.26
----------
 * Fix cache strict look up bug and unify cache management, thanks @eugene-gurevich

v1.0.25
----------
 * save possible lengths to metadata, change implementation to use, add IS_POSSIBLE_LOCAL_ONLY and INVALID_LENGTH as possible return values to IsPossibleNumberWithReason
 * update metadata to version v8.9.12

v1.0.24
----------
 * update to metadata for v8.9.10

v1.0.23
----------
 * add GetSupportedCallingCodes
 * return sets as map[int]bool instead of map[int]struct{}

v1.0.22
----------
* add GetCarrierForNumber and GetGeocodingForNumber

v1.0.21
----------
 * Update for libphonenumber v8.9.8

v1.0.20
----------
 * updated metadata for v8.9.7

v1.0.19
----------
 * update metadata for v8.9.6

v1.0.18
----------
 * update metadata for v8.9.5

v1.0.17
----------
 * Fix maybe strip extension, thanks @vlastv

