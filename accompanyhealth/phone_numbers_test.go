package accompanyhealth

import (
	"errors"
	"testing"

	ahcontext "github.com/Accompany-Health/ah-context/context"
	stats "github.com/Accompany-Health/ah-stats"
)

func init() {
	stats.InitStats("test")
}

func TestFormatPhoneNumberForDatabase(t *testing.T) {
	result := "+12345678901"
	ctx := ahcontext.CreateSystemContext("test")
	tests := []struct {
		name        string
		phoneNumber string
		expected    string
		expectedErr error
	}{
		{
			name:        "valid phone number",
			phoneNumber: "2345678901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with spaces",
			phoneNumber: "234 567 8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with dashes",
			phoneNumber: "234-567-8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with dots",
			phoneNumber: "234.567.8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with parentheses",
			phoneNumber: "(234) 567-8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with parentheses and spaces",
			phoneNumber: "(234) 567 8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with parentheses and dashes",
			phoneNumber: "(234) 567-8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with parentheses and dots",
			phoneNumber: "(234) 567.8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "valid phone number with country code",
			phoneNumber: "+1 234 567 8901",
			expected:    result,
		},
		{
			name:        "valid phone number with all",
			phoneNumber: "+1 (234) 567-8901",
			expected:    result,
			expectedErr: nil,
		},
		{
			name:        "failure due to parsing error",
			phoneNumber: "+1800TEST",
			expected:    "",
			expectedErr: errors.New("The phone number supplied was empty."),
		},
		{
			name:        "failure due to parsing error",
			phoneNumber: "TEST",
			expected:    "",
			expectedErr: errors.New("The phone number supplied was empty."),
		},
		{
			name:        "parsing error due to empty phone number",
			phoneNumber: "",
			expected:    "",
			expectedErr: errors.New("The phone number supplied was empty."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := FormatPhoneNumberForDatabase(tt.phoneNumber); got != tt.expected && err != tt.expectedErr {
				t.Errorf("FormatPhoneNumberForDatabase() = %v, want %v", got, tt.expected)
			}
			if got := IsValidPhoneNumber(ctx, tt.phoneNumber); got != (tt.expected != "") {
				t.Errorf("IsValidPhoneNumber() = %v, want %v", got, (tt.expected != ""))
			}
		})
	}
}

func TestPhoneNumberIsEqual(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		expected    bool
		target      string
	}{
		{
			name:        "phone number",
			phoneNumber: "2345678901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #1",
			phoneNumber: " 234 567 8901 ",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #2",
			phoneNumber: "234-567-8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #3",
			phoneNumber: "234.567.8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #4",
			phoneNumber: "(234) 567-8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #5",
			phoneNumber: "(234) 567 8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #6",
			phoneNumber: "(234) 567-8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with with different formats #7",
			phoneNumber: "(234) 567.8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone number with different formats",
			phoneNumber: "+1 (234) 567-8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers with same formats",
			phoneNumber: "+1 234 567 8901",
			expected:    true,
			target:      "+1 234-567-8901",
		},
		{
			name:        "phone numbers not equal",
			phoneNumber: "+1 334-567-8901",
			expected:    false,
			target:      "+1 234-567-8901",
		},
		{
			name:        "invalid phone numbers but equal values",
			phoneNumber: "TEST",
			expected:    true,
			target:      "TEST",
		},
		{
			name:        "invalid phone number and valid number",
			phoneNumber: "+1234567",
			expected:    false,
			target:      "+1 234-567-8901",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEqual(tt.phoneNumber, tt.target); got != tt.expected {
				t.Errorf("PatientService.IsEqual() = %v, want %v", got, tt.expected)
			}
		})
	}
}
