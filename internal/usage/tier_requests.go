package usage

import (
	"github.com/shopspring/decimal"
)

// Use this method to calculate resource's tiers
// In tierLimits send only [next] tiers (first-tier used as next-tier)
// 'Over' tier calculated as the remainder of (requests - sum of 'next' tiers)

func CalculateTierBuckets(requestCount decimal.Decimal, tierLimits []int) []decimal.Decimal {
	overTier := false
	tiers := make([]decimal.Decimal, 0)

	for limit := range tierLimits {
		tier := decimal.NewFromInt(int64(tierLimits[limit]))

		if requestCount.GreaterThanOrEqual(tier) {
			tiers = append(tiers, tier)
			requestCount = requestCount.Sub(tier)
			overTier = true
		} else if requestCount.LessThan(tier) {
			tiers = append(tiers, requestCount)
			overTier = false
		}
	}

	if overTier {
		tiers = append(tiers, requestCount)
	} else {
		tiers = append(tiers, decimal.NewFromInt(0))
	}
	return tiers
}
