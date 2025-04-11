package apiclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	json "github.com/json-iterator/go"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/vcs"
)

func TestPricingAPIClient_PerformRequest(t *testing.T) {
	client := retryablehttp.NewClient()
	conf := config.DefaultConfig()
	i := 0
	requestMap := map[int]string{}
	responseMap := map[int][]byte{
		0: []byte(`
[
	{"data":{"products":[{"prices":[{"priceHash":"99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1250000000"}]}]}},
	{"data":{"products":[{"prices":[{"priceHash":"d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78","USD":"0.0650000000"}]}]}}
]
`),
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		requestMap[i] = string(bodyBytes)

		if len(responseMap)-1 < i {
			t.Errorf("received unexpected request to pricing API at index: %d body: %s", i, bodyBytes)
			_, err := w.Write([]byte{})
			assert.NoError(t, err)
			return
		}

		_, err := w.Write(responseMap[i])
		assert.NoError(t, err)
		i++
	}))

	conf.PricingAPIEndpoint = ts.URL
	ctx := &config.RunContext{
		Config:        conf,
		State:         nil,
		VCSMetadata:   vcs.Metadata{},
		CMD:           "",
		ContextValues: nil,
		ModuleMutex:   nil,
		StartTime:     0,
		OutWriter:     nil,
		ErrWriter:     nil,
		Exit:          nil,
	}

	c := &PricingAPIClient{
		APIClient: APIClient{
			httpClient: client.StandardClient(),
			endpoint:   ctx.Config.PricingAPIEndpoint,
			apiKey:     ctx.Config.APIKey,
			uuid:       ctx.UUID(),
		},
	}
	initCache(ctx, c)

	cachedProduct := &schema.ProductFilter{
		VendorName: strPtr("aws"),
		Service:    strPtr("some service"),
	}

	resources := []*schema.Resource{
		{
			Name: "test",
			CostComponents: []*schema.CostComponent{
				{
					ProductFilter: cachedProduct,
					PriceFilter: &schema.PriceFilter{
						PurchaseOption: strPtr("something else"),
					},
				},
				{
					ProductFilter: &schema.ProductFilter{
						VendorName: strPtr("aws"),
						Service:    strPtr("some service 2"),
					},
				},
			},
		},
		{
			Name: "test2",
			CostComponents: []*schema.CostComponent{
				{
					ProductFilter: cachedProduct,
				},
				{
					ProductFilter: &schema.ProductFilter{
						VendorName: strPtr("aws"),
						Service:    strPtr("some service 2"),
					},
				},
			},
		},
	}
	q := c.buildQuery(cachedProduct, nil, "USD")
	k, err := hashstructure.Hash(q, hashstructure.FormatV2, nil)
	assert.NoError(t, err)
	c.cache.Add(k, cacheValue{Result: gjson.Parse(`{"data":{"products":[{"prices":[{"priceHash":"cached-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1250000000"}]}]}`), ExpiresAt: time.Now().Add(time.Hour)})

	batches := c.BatchRequests(resources, 100, "USD")
	result, err := c.PerformRequest(batches[0])

	assert.Len(t, requestMap, 1, "invalid number of requests made to pricing API")
	assert.JSONEq(
		t,
		`
[
  {
    "query": "query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {products(filter: $productFilter) {prices(filter: $priceFilter) {priceHashUSD}}}",
    "variables": {
      "priceFilter": {
        "purchaseOption": "something else"
      },
      "productFilter": {
        "vendorName": "aws",
        "service": "some service"
      }
    }
  },
  {
    "query": "query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {products(filter: $productFilter) {prices(filter: $priceFilter) {priceHashUSD}}}",
    "variables": {
      "priceFilter": null,
      "productFilter": {
        "vendorName": "aws",
        "service": "some service 2"
      }
    }
  }
]
`,
		strings.ReplaceAll(strings.ReplaceAll(requestMap[0], `\t`, ""), `\n`, ""),
	)
	assert.NoError(t, err)
	assertResults(t, result)

	// try the request again to ensure we cache everything now
	result, err = c.PerformRequest(batches[0])
	assert.NoError(t, err)
	assertResults(t, result)
}

func assertResults(t *testing.T, result []PriceQueryResult) {
	require.Len(t, result, 4)
	assertResultEqual(
		t,
		result[0],
		"test",
		`{"vendorName":"aws","service":"some service"}`,
		`{"purchaseOption":"something else"}`,
		`{"data":{"products":[{"prices":[{"priceHash":"99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1250000000"}]}]}}`,
	)
	assertResultEqual(
		t,
		result[1],
		"test",
		`{"vendorName":"aws","service":"some service 2"}`,
		`null`,
		`{"data":{"products":[{"prices":[{"priceHash":"d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78","USD":"0.0650000000"}]}]}}`,
	)
	assertResultEqual(
		t,
		result[2],
		"test2",
		`{"vendorName":"aws","service":"some service"}`,
		`null`,
		`{"data":{"products":[{"prices":[{"priceHash":"cached-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1250000000"}]}]}`,
	)
	assertResultEqual(
		t,
		result[3],
		"test2",
		`{"vendorName":"aws","service":"some service 2"}`,
		`null`,
		`{"data":{"products":[{"prices":[{"priceHash":"d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78","USD":"0.0650000000"}]}]}}`,
	)
}

func assertResultEqual(t *testing.T, queryResult PriceQueryResult, name string, expProductFilter string, expPriceFilter string, result string) {
	t.Helper()

	assert.Equal(t, name, queryResult.Resource.Name)
	productFilter := queryResult.CostComponent.ProductFilter
	b, _ := json.Marshal(productFilter)
	assert.JSONEq(t, expProductFilter, string(b))

	priceFilter := queryResult.CostComponent.PriceFilter
	b, _ = json.Marshal(priceFilter)
	assert.JSONEq(t, expPriceFilter, string(b))

	assert.Equal(t, result, queryResult.Result.Raw)
}

func strPtr(str string) *string {
	return &str
}
