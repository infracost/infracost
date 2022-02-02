package aws_test

import (
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCloudwatchMetricResolution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		inputValue decimal.Decimal
		expected   bool
	}{
		{decimal.NewFromInt(60), true},
		{decimal.NewFromInt(120), true},
		{decimal.NewFromInt(30), false},
		{decimal.NewFromInt(10), false},
	}

	resource := resources.CloudwatchMetricAlarm{}

	for _, test := range tests {
		actual := resource.CalcMetricResolution(test.inputValue)
		assert.Equal(t, test.expected, actual)
	}
}
