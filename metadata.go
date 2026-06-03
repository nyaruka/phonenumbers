package phonenumbers

import _ "embed"

//go:embed data/shortnumber_metadata.xml.gz
var shortNumberData []byte

//go:embed data/alternateformats_metadata.xml.gz
var alternateFormatsData []byte
