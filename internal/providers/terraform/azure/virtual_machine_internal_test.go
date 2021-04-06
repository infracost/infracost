package azure

import (
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestMapDiskName(t *testing.T) {
	tests := []struct {
		diskType   string
		diskSizeGB int
		expected   string
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
		actual := mapDiskName(test.diskType, test.diskSizeGB)
		assert.Equal(t, test.expected, actual)
	}
}
