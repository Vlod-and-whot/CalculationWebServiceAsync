package test

import (
	"testing"

	"CalcWebServiceAsync/pkg/calculation"
)

func TestComputeOperations(t *testing.T) {
	delay := 1

	tests := []struct {
		a, b      float64
		op        string
		expected  float64
		expectErr bool
	}{
		{2, 3, "+", 5, false},
		{5, 3, "-", 2, false},
		{4, 3, "*", 12, false},
		{10, 2, "/", 5, false},
		{10, 0, "/", 0, true},
	}

	for _, tc := range tests {
		result, err := calculation.Compute(tc.a, tc.b, tc.op, delay)
		if tc.expectErr {
			if err == nil {
				t.Errorf("Expected error for operation %s with %v and %v, but got none", tc.op, tc.a, tc.b)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for operation %s: %v", tc.op, err)
			}
			if result != tc.expected {
				t.Errorf("Expected result %v for operation %s, but got %v", tc.expected, tc.op, result)
			}
		}
	}
}
