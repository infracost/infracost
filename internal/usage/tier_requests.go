package usage

import (
	"github.com/shopspring/decimal"
	"strconv"
)

func CalculateTierRequests(requestCount decimal.Decimal, tierLimits []int) map[string]decimal.Decimal {
	requestTierMap := map[string]decimal.Decimal{}

	for limit := range tierLimits {
		tier := decimal.NewFromInt(int64(tierLimits[limit]))

		if requestCount.GreaterThan(tier) && limit+1 < len(tierLimits) {
			requestTierMap[strconv.Itoa(limit+1)] = tier
		} else {
			requestTierMap[strconv.Itoa(limit+1)] = requestCount
		}

		if requestCount.GreaterThanOrEqual(tier) {
			requestCount = requestCount.Sub(tier)
		} else {
			requestCount = requestCount.Sub(requestCount)
		}
	}

	return requestTierMap
}
