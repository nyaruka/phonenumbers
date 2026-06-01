package phonenumbers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuilderProcessesSmsServices verifies the short-number metadata builder now
// reads the <smsServices> element (it was previously dropped, leaving
// IsSmsServiceForRegion non-functional). Exercised through the builder directly
// since the committed embedded short metadata predates this fix and carries no
// smsServices data.
func TestBuilderProcessesSmsServices(t *testing.T) {
	const shortXML = `<?xml version="1.0" encoding="UTF-8"?>
<phoneNumberMetadata>
  <territories>
    <territory id="US" countryCode="1">
      <generalDesc>
        <nationalNumberPattern>[1-9]\d{2,5}</nationalNumberPattern>
      </generalDesc>
      <shortCode>
        <possibleLengths national="[3-6]"/>
        <exampleNumber>112</exampleNumber>
        <nationalNumberPattern>[2-9]\d{2,5}</nationalNumberPattern>
      </shortCode>
      <smsServices>
        <possibleLengths national="5,6"/>
        <exampleNumber>20000</exampleNumber>
        <nationalNumberPattern>[2-9]\d{4,5}</nationalNumberPattern>
      </smsServices>
    </territory>
  </territories>
</phoneNumberMetadata>`

	coll, err := BuildPhoneMetadataCollection([]byte(shortXML), false, false, true)
	require.NoError(t, err)
	require.Len(t, coll.GetMetadata(), 1)

	us := coll.GetMetadata()[0]
	sms := us.GetSmsServices()
	require.NotNil(t, sms)
	assert.Equal(t, `[2-9]\d{4,5}`, sms.GetNationalNumberPattern())
	assert.Equal(t, []int32{5, 6}, sms.GetPossibleLength())

	// And the matching helper used by IsSmsServiceForRegion now finds it.
	assert.True(t, matchesPossibleNumberAndNationalNumber("21234", sms))
	assert.False(t, matchesPossibleNumberAndNationalNumber("112", sms))
}
