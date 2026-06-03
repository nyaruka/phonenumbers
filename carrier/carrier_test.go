package carrier

import (
	"sync"
	"testing"

	"github.com/nyaruka/phonenumbers/v2"
)

func TestNameForNumber(t *testing.T) {
	tests := []struct {
		num      string
		lang     string
		expected string
	}{
		{num: "+8613702032331", lang: "en", expected: "China Mobile"},
		{num: "+8613702032331", lang: "zh", expected: "中国移动"},
		{num: "+6281377468527", lang: "en", expected: "Telkomsel"},
		{num: "+8613323241342", lang: "en", expected: "China Telecom"},
		{num: "+61491570156", lang: "en", expected: "Telstra"},
		{num: "+917999999543", lang: "en", expected: "Reliance Jio"},
		{num: "+593992218722", lang: "en", expected: "Claro"},
		// no carrier data for the number
		{num: "+201987654321", lang: "en", expected: ""},
		// no data for the requested language (and no English fallback)
		{num: "+201987654321", lang: "notFound", expected: ""},
	}
	for _, test := range tests {
		number, err := phonenumbers.Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, err := NameForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to get carrier for the number %s: %s", test.num, err)
		}
		if test.expected != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, carrier, test.num)
		}
	}
}

func TestNameForNumberConcurrency(t *testing.T) {
	number, _ := phonenumbers.Parse("+8613702032331", "ZZ")

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := NameForNumber(number, "en")
			if err != nil {
				t.Errorf("Failed to get carrier for the number %s: %s", "+8613702032331", err)
			}
		}()
	}

	wg.Wait()
}

func TestSafeDisplayName(t *testing.T) {
	tests := []struct {
		num      string
		lang     string
		expected string
	}{
		{num: "+447387654321", lang: "en", expected: ""},
		{num: "+244917654321", lang: "en", expected: "Movicel"},
	}
	for _, test := range tests {
		number, err := phonenumbers.Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, err := SafeDisplayName(number, test.lang)
		if err != nil {
			t.Errorf("Failed to SafeDisplayName for the number %s: %s", test.num, err)
		}
		if test.expected != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, carrier, test.num)
		}
	}
}

func BenchmarkNameForNumber(b *testing.B) {
	number, _ := phonenumbers.Parse("+8613702032331", "ZZ")

	for n := 0; n < b.N; n++ {
		_, err := NameForNumber(number, "en")
		if err != nil {
			b.Errorf("Failed to get carrier for the number %s: %s", "+8613702032331", err)
		}
	}
}
