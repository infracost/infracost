package aws

import (
    "testing"

    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
)

func TestCloudwatchMetricResolution(t *testing.T) {
    tests := []struct{
        inputValue decimal.Decimal
        expected bool
    }{
        {decimal.NewFromInt(60), true},
        {decimal.NewFromInt(120), true},
        {decimal.NewFromInt(30), false},
        {decimal.NewFromInt(10), false},
    }

    for _, test := range tests {
        actual := calcMetricResolution(test.inputValue)
        assert.Equal(t, test.expected, actual)
    }
}