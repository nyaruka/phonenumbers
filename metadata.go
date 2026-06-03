package phonenumbers

import "embed"

//go:embed data/prefix_to_carriers/*.gz
var carrierData embed.FS
var carrierDataPath = "data/prefix_to_carriers"

//go:embed data/prefix_to_geocodings/*.gz
var geocodingData embed.FS
var geocodingDataPath = "data/prefix_to_geocodings"

//go:embed data/prefix_to_timezone.xml.gz
var timezoneData []byte

//go:embed data/shortnumber_metadata.xml.gz
var shortNumberData []byte

//go:embed data/alternateformats_metadata.xml.gz
var alternateFormatsData []byte
