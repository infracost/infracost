package azure

import (
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestMapDiskName(t *testing.T) {
	tests := []struct {
		diskType      string
		requestedSize int
		expected      string
	}{
		{"Standard_LRS", 32, "S4"},
		{"Standard_LRS", 4000, "S50"},
		{"Standard_LRS", 8192, "S60"},
		{"Standard_LRS", 400000, ""},
		{"StandardSSD_LRS", 32, "E4"},
		{"StandardSSD_LRS", 4000, "E50"},
		{"StandardSSD_LRS", 8192, "E60"},
		{"StandardSSD_LRS", 400000, ""},
		{"Premium_LRS", 32, "P4"},
		{"Premium_LRS", 4000, "P50"},
		{"Premium_LRS", 8192, "P60"},
		{"Premium_LRS", 400000, ""},
	}

	for _, test := range tests {
		actual := mapDiskName(test.diskType, test.requestedSize)
		assert.Equal(t, test.expected, actual)
	}
}

func TestMapUltraDiskSize(t *testing.T) {
	tests := []struct {
		requestedSize int
		expected      int
	}{
		{0, 4},
		{3, 4},
		{4, 4},
		{5, 8},
		{1023, 1024},
		{1024, 1024},
		{1025, 2048},
		{65536, 65536},
		{65537, 65536},
	}

	for _, test := range tests {
		actual := mapUltraDiskSize(test.requestedSize)
		assert.Equal(t, test.expected, actual)
	}
}
