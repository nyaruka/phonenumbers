package phonenumbers

import (
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Metadata loads once per process, so each scenario re-runs this test binary to
// start from a process in which nothing has loaded.
func runInFreshProcess(t *testing.T, scenario func(t *testing.T)) {
	const env = "PHONENUMBERS_FRESH_PROCESS"
	if os.Getenv(env) == "1" {
		scenario(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
	cmd.Env = append(os.Environ(), env+"=1")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestShortNumberMetadataCollectionBeforeFirstLookup(t *testing.T) {
	runInFreshProcess(t, func(t *testing.T) {
		require.Empty(t, shortNumberRegionToMetadataMap)

		coll, err := ShortNumberMetadataCollection()
		require.NoError(t, err)
		require.NotEmpty(t, coll.GetMetadata())

		assert.True(t, IsEmergencyNumber("911", "US"))

		again, err := ShortNumberMetadataCollection()
		require.NoError(t, err)
		assert.Same(t, coll, again)
	})
}

func TestConcurrentFirstUse(t *testing.T) {
	runInFreshProcess(t, func(t *testing.T) {
		var wg sync.WaitGroup
		for range 16 {
			wg.Go(func() {
				num, err := Parse("+12065550100", "")
				assert.NoError(t, err)
				assert.True(t, IsValidNumber(num))
				assert.True(t, IsEmergencyNumber("911", "US"))

				_, err = MetadataCollection()
				assert.NoError(t, err)
				_, err = ShortNumberMetadataCollection()
				assert.NoError(t, err)
			})
		}
		wg.Wait()
	})
}
