gophone
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
