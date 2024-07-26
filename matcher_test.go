package phonenumbers

import (
	"testing"
)

func TestNewPhoneNumberMatcherForRegion(t *testing.T) {
	tests := []struct {
		name   string
		seq    string
		region string
		want   map[uint64]int // expected phone numbers with their expected counts
	}{
		{
			name:   "Valid US numbers",
			seq:    "Call me at 202-555-0130 or 415-555-0198 for more info.",
			region: "US",
			want:   map[uint64]int{2025550130: 1, 4155550198: 1},
		},
		{
			name:   "Invalid patterns mixed with valid numbers",
			seq:    "Try 12345 and then call 503-555-0110 or reach out at 999-000",
			region: "US",
			want:   map[uint64]int{5035550110: 1},
		},
		{
			name:   "Valid Tunisian numbers",
			seq:    "Call me at 71 123 456 or 71 123 457 for more info.",
			region: "TN",
			want:   map[uint64]int{71123456: 1, 71123457: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPhoneNumberMatcherForRegion(tt.seq, tt.region)
			got := make(map[uint64]int)
			for matcher != nil {
				got[*matcher.PhoneNumber.NationalNumber]++
				matcher = matcher.Next
			}

			for num, count := range tt.want {
				if got[num] != count {
					t.Errorf("Test %s failed: expected %d occurrences of %d, got %d", tt.name, count, num, got[num])
				}
			}
		})
	}
}
