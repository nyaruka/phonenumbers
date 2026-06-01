package phonenumbers

// Harness for running the library against upstream libphonenumber's synthetic
// test metadata, the Go analogue of upstream's TestMetadataTestCase + RegionCode.
//
// Upstream's PhoneNumberUtilTest et al. run against resources/PhoneNumberMetadata
// ForTesting.xml — a frozen, hand-authored set of regions with made-up number
// ranges — so test expectations never change when real-world metadata updates.
// We compile that same XML (committed under testdata/, reconciled against
// upstream v9.0.32) through our builder and swap it in via the metadata seam.
//
// Mirrors TestMetadataTestCase: useTestMetadata == setUp's setInstance(test),
// and the registered t.Cleanup == tearDown's setInstance(null). As upstream
// notes, such tests must NOT run in parallel (the active metadata is global).

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// regionCode mirrors upstream's RegionCode constants so ported tests can read
// regionCode.US the way the Java tests read RegionCode.US.
var regionCode = struct {
	UN001, AD, AE, AM, AO, AR, AU, BB, BR, BS, BY, CA, CC, CN, CO, CX, DE, FR,
	GB, GG, IT, JP, KR, MX, NZ, PL, RE, RU, SE, SG, TA, US, UZ, YT, ZZ string
}{
	UN001: "001", AD: "AD", AE: "AE", AM: "AM", AO: "AO", AR: "AR", AU: "AU",
	BB: "BB", BR: "BR", BS: "BS", BY: "BY", CA: "CA", CC: "CC", CN: "CN",
	CO: "CO", CX: "CX", DE: "DE", FR: "FR", GB: "GB", GG: "GG", IT: "IT",
	JP: "JP", KR: "KR", MX: "MX", NZ: "NZ", PL: "PL", RE: "RE", RU: "RU",
	SE: "SE", SG: "SG", TA: "TA", US: "US", UZ: "UZ", YT: "YT", ZZ: "ZZ",
}

var (
	testMetadataOnce      sync.Once
	testMetadataContainer *metadataContainer
	testMetadataErr       error
)

// loadTestMetadataContainer compiles the synthetic test XML once and caches the
// resulting container, following the exact build path cmd/buildmetadata uses for
// the real metadata.
func loadTestMetadataContainer() (*metadataContainer, error) {
	testMetadataOnce.Do(func() {
		xml, err := os.ReadFile("testdata/PhoneNumberMetadataForTesting.xml")
		if err != nil {
			testMetadataErr = err
			return
		}
		coll, err := BuildPhoneMetadataCollection(xml, false, false, false)
		if err != nil {
			testMetadataErr = err
			return
		}
		testMetadataContainer, testMetadataErr = newMetadataContainer(coll, BuildCountryCodeToRegionMap(coll))
	})
	return testMetadataContainer, testMetadataErr
}

// useTestMetadata activates the synthetic test metadata for the duration of the
// test and restores the embedded metadata afterwards (via t.Cleanup, so it runs
// even on a panic/FailNow). Do not call from a test that uses t.Parallel().
func useTestMetadata(t *testing.T) {
	mc, err := loadTestMetadataContainer()
	require.NoError(t, err)
	t.Cleanup(useMetadata(mc))
}
