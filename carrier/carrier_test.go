package carrier

import (
	"sync"
	"testing"

	"github.com/nyaruka/phonenumbers/v2"
)

func TestForNumber(t *testing.T) {
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
	}
	for _, test := range tests {
		number, err := phonenumbers.Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, err := ForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to getCarrier for the number %s: %s", test.num, err)
		}
		if test.expected != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, carrier, test.num)
		}
	}
}

func TestWithPrefixForNumber(t *testing.T) {
	tests := []struct {
		num             string
		lang            string
		expectedCarrier string
		expectedPrefix  int
	}{
		{num: "+8613702032331", lang: "en", expectedCarrier: "China Mobile", expectedPrefix: 86137},
		{num: "+8613702032331", lang: "zh", expectedCarrier: "中国移动", expectedPrefix: 86137},
		{num: "+6281377468527", lang: "en", expectedCarrier: "Telkomsel", expectedPrefix: 62813},
		{num: "+8613323241342", lang: "en", expectedCarrier: "China Telecom", expectedPrefix: 86133},
		{num: "+61491570156", lang: "en", expectedCarrier: "Telstra", expectedPrefix: 6149},
		{num: "+917999999543", lang: "en", expectedCarrier: "Reliance Jio", expectedPrefix: 917999},
		{num: "+593992218722", lang: "en", expectedCarrier: "Claro", expectedPrefix: 5939922},
		{num: "+201987654321", lang: "en", expectedCarrier: "", expectedPrefix: 0},
		{num: "+201987654321", lang: "notFound", expectedCarrier: "", expectedPrefix: 0},
	}
	for _, test := range tests {
		number, err := phonenumbers.Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, prefix, err := WithPrefixForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to getCarrierWithPrefix for the number %s: %s", test.num, err)
		}
		if test.expectedCarrier != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expectedCarrier, carrier, test.num)
		}
		if test.expectedPrefix != prefix {
			t.Errorf("Expected '%d', got '%d' for '%s'", test.expectedPrefix, prefix, test.num)
		}
	}
}

func TestWithPrefixForNumberConcurrency(t *testing.T) {
	number, _ := phonenumbers.Parse("+8613702032331", "ZZ")

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := WithPrefixForNumber(number, "en")
			if err != nil {
				t.Errorf("Failed to getCarrierWithPrefix for the number %s: %s", "+8613702032331", err)
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

func BenchmarkWithPrefixForNumber(b *testing.B) {
	number, _ := phonenumbers.Parse("+8613702032331", "ZZ")

	for n := 0; n < b.N; n++ {
		_, _, err := WithPrefixForNumber(number, "en")
		if err != nil {
			b.Errorf("Failed to getCarrierWithPrefix for the number %s: %s", "+8613702032331", err)
		}
	}
}
