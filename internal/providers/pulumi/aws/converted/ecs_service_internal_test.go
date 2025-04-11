package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVCPUMemoryString(t *testing.T) {
	t.Parallel()
	half := float64(0.5)
	one := float64(1)
	two := float64(2)

	tests := []struct {
		inputStr string
		expected float64
	}{
		{"1GB", one},
		{"1gb", one},
		{" 1 Gb ", one}, // mixed case and pre/middle/post whitespace
		{"0.5 GB", half},
		{".5 GB", half},
		{"1VCPU", one},
		{"1vcpu", one},
		{" 1 vCPU ", one}, // mixed case and pre/middle/post whitespace
		{"1024", one},
		{" 1024 ", one},
		{"512", half},
		{"2048", two},
	}

	for _, test := range tests {
		actual := parseVCPUMemoryString(test.inputStr)
		assert.Equal(t, test.expected, actual)
	}
}
