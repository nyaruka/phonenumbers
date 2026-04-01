package phonenumbers

import "embed"

//go:embed data/countrycode_to_region.xml.gz
var regionData []byte

//go:embed data/metadata.xml.gz
var numberData []byte

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
