package aws

import (
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestCalcApiV2Requests(t *testing.T) {
    oneMillionApi := decimal.NewFromInt(1000000)
    oneBillionApi := decimal.NewFromInt(1000000000)

    apiTierOneLimit := decimal.NewFromInt(300000000)
    apiTierTwoLimit := decimal.NewFromInt(300000001)

    var apiTierRequests = map[string]decimal.Decimal{
        "apiRequestTierOne":   decimal.Zero,
        "apiRequestTierTwo":   decimal.Zero,
    }

    var apiTierOneRequests = map[string]decimal.Decimal{
        "apiRequestTierOne":   decimal.NewFromInt(1000000),
        "apiRequestTierTwo":   decimal.Zero,
    }

    var apiTierTwoRequests = map[string]decimal.Decimal{
        "apiRequestTierOne":   decimal.NewFromInt(300000000),
        "apiRequestTierTwo":   decimal.NewFromInt(1),
    }

    tests := []struct {
        requests          decimal.Decimal
        inputTierRequests map[string]decimal.Decimal
        expected          map[string]decimal.Decimal
    }{
        {requests: oneMillionApi, inputTierRequests: apiTierRequests, expected: apiTierOneRequests},
        {requests: oneBillionApi, inputTierRequests: apiTierRequests, expected: apiTierTwoRequests},
    }

    for _, test := range tests {
        actual := calculateApiV2Requests(test.requests, test.inputTierRequests, apiTierOneLimit, apiTierTwoLimit)

        if test.requests == oneMillionApi {
            assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
        }

        if test.requests == oneBillionApi {
            assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
            assert.Equal(t, test.expected["tierTwo"], actual["tierTwo"])
        }
    }
}