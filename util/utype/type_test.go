package utype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStrAnyMap(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]any
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: map[string]any{},
		},
		{
			name:     "valid string map",
			input:    map[string]string{"key1": "value1", "key2": "value2"},
			expected: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "invalid type",
			input:    123,
			expected: map[string]any{},
		},
		{
			name:     "valid non-string map",
			input:    map[string]int{"key1": 1.0, "key2": 2.0}, // Don't use int, it's type will be converted to json float64, the test result never match
			expected: map[string]any{"key1": 1.0, "key2": 2.0},
		},
		{
			name: "valid struct",
			input: struct {
				Name string
			}{Name: "name"},
			expected: map[string]any{"Name": "name"},
		},
		{
			name: "contains valid empty slice",
			input: struct {
				S []string
			}{S: []string{}},
			expected: map[string]any{},
		},
		// Add more test cases if needed
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ToStrAnyMap(test.input) // Call the function under test
			//fmt.Printf("%+v %+v %+v\n", 111, test.expected, result)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGeometry_Scan(t *testing.T) {
	v := Geometry{}
	var dataSet = []struct {
		input string
	}{
		{"POINT(12 2)"},
		{"POINT(12.1 2)"},
		{"POINT(123 2.345)"},
		{"POINT(112.1 2.3455)"},
		// 带负号
		{"POINT(-1.1 -2.3455)"},
		{"POINT(-112.1 -2)"},
		{"POINT(-234.23 2.342)"},
	}
	for _, data := range dataSet {
		v.Geometry = nil
		err := v.Scan(data.input)
		if err != nil {
			t.Fatal(err)
		}
		if v.Geometry == nil {
			t.Error("Geometry is nil")
			continue
		}
		t.Logf("%+v", v)
	}
}
