package aws

import "testing"

func TestElasticacheReservationResolverPriceFilterMalformedNodeType(t *testing.T) {
	for _, nodeType := range []string{"", "invalid"} {
		r := elasticacheReservationResolver{
			term:          "1_year",
			paymentOption: "no_upfront",
			cacheNodeType: nodeType,
		}
		_, err := r.PriceFilter()
		if err == nil {
			t.Errorf("expected an error for malformed cacheNodeType %q, got nil", nodeType)
		}
	}
}
