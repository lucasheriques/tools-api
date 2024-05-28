package validator

import (
	"regexp"
	"testing"

	"tools.lucasfaria.dev/internal/assert"
)

func TestNewValidator(t *testing.T) {
	v := New()
	assert.Equal(t, v != nil, true)
	assert.Equal(t, len(v.Errors), 0)
}

func TestValidator_Valid(t *testing.T) {
	v := New()
	assert.Equal(t, v.Valid(), true)

	v.AddError("key", "error message")
	assert.Equal(t, v.Valid(), false)
}

func TestValidator_AddError(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		message  string
		expected map[string]string
	}{
		{"Add single error", "key1", "error message 1", map[string]string{"key1": "error message 1"}},
		{"Add another error", "key2", "error message 2", map[string]string{"key1": "error message 1", "key2": "error message 2"}},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v.AddError(tt.key, tt.message)
			assert.Equal(t, v.Errors[tt.key], tt.message)
		})
	}
}

func TestValidator_Check(t *testing.T) {
	tests := []struct {
		name      string
		ok        bool
		key       string
		message   string
		numErrors int
	}{
		{"Check false", false, "key1", "error message 1", 1},
		{"Check true", true, "key2", "error message 2", 0},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v.Check(tt.ok, tt.key, tt.message)
			if tt.ok {
				assert.Equal(t, tt.numErrors, 0)
			} else {
				assert.Equal(t, tt.numErrors, 1)
			}
		})
	}
}

func TestPermittedValue(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		permitted []int
		expected  bool
	}{
		{"Permitted value", 5, []int{1, 2, 3, 4, 5}, true},
		{"Not permitted value", 6, []int{1, 2, 3, 4, 5}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := PermittedValue(tt.value, tt.permitted...)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestPermittedValues(t *testing.T) {
	tests := []struct {
		name            string
		values          []int
		permittedValues []int
		expected        bool
	}{
		{"All permitted values", []int{1, 2, 3}, []int{1, 2, 3, 4, 5}, true},
		{"Some not permitted values", []int{1, 2, 6}, []int{1, 2, 3, 4, 5}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := PermittedValues(tt.values, tt.permittedValues)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestMatches(t *testing.T) {
	rx := regexp.MustCompile("^[a-zA-Z0-9]+$")

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Matches regex", "abc123", true},
		{"Does not match regex", "abc-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Matches(tt.value, rx)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		expected bool
	}{
		{"Unique values", []int{1, 2, 3}, true},
		{"Non-unique values", []int{1, 2, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Unique(tt.values)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected bool
	}{
		{"Valid date", "2024-05-28", true},
		{"Invalid date (Feb 30)", "2023-02-30", false},
		{"Invalid month", "2022-13-01", false},
		{"Valid date", "2022-12-01", true},
		{"Invalid leap year date", "2022-02-29", false},
		{"Invalid day", "2021-11-31", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ValidateDate(tt.dateStr)
			assert.Equal(t, actual, tt.expected)
		})
	}
}
