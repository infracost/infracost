package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	excludedEnv = map[string]struct{}{
		"repoMetadata": {},
	}
	pricingClient *PricingAPIClient
	pricingMu     = &sync.Mutex{}
)

type PricingAPIClient struct {
	APIClient
	Currency       string
	EventsDisabled bool

	mu          *sync.RWMutex
	cache       *map[uint64]gjson.Result
	objectLimit int
	cacheFile   string
}

type PriceQueryKey struct {
	Resource      *schema.Resource
	CostComponent *schema.CostComponent
}

type PriceQueryResult struct {
	PriceQueryKey
	Result gjson.Result

	filled bool
}

type BatchRequest struct {
	keys    []PriceQueryKey
	queries []GraphQLQuery
}

// GetPricingAPIClient initializes and returns an instance of PricingAPIClient
// using the given RunContext configuration. If an instance of PricingAPIClient
// has already been created, it will return the existing instance. This is done
// to ensure that the client cache is global across the application.
func GetPricingAPIClient(ctx *config.RunContext) *PricingAPIClient {
	pricingMu.Lock()
	defer pricingMu.Unlock()

	if pricingClient != nil {
		return pricingClient
	}

	currency := ctx.Config.Currency
	if currency == "" {
		currency = "USD"
	}

	tlsConfig := tls.Config{} // nolint: gosec

	if ctx.Config.TLSCACertFile != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		caCerts, err := os.ReadFile(ctx.Config.TLSCACertFile)
		if err != nil {
			log.Errorf("Error reading CA cert file %s: %v", ctx.Config.TLSCACertFile, err)
		} else {
			ok := rootCAs.AppendCertsFromPEM(caCerts)

			if !ok {
				log.Warningf("No CA certs appended, only using system certs")
			} else {
				log.Debugf("Loaded CA certs from %s", ctx.Config.TLSCACertFile)
			}
		}

		tlsConfig.RootCAs = rootCAs
	}

	if ctx.Config.TLSInsecureSkipVerify != nil {
		tlsConfig.InsecureSkipVerify = *ctx.Config.TLSInsecureSkipVerify // nolint: gosec
	}

	client := retryablehttp.NewClient()
	client.Logger = &LeveledLogger{Logger: logging.Logger.WithField("library", "retryablehttp")}
	client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = &tlsConfig

	c := &PricingAPIClient{
		APIClient: APIClient{
			httpClient: client.StandardClient(),
			endpoint:   ctx.Config.PricingAPIEndpoint,
			apiKey:     ctx.Config.APIKey,
			uuid:       ctx.UUID(),
		},
		Currency:       currency,
		EventsDisabled: ctx.Config.EventsDisabled,
	}

	initCache(ctx, c)
	pricingClient = c
	return c
}

func initCache(ctx *config.RunContext, c *PricingAPIClient) {
	if ctx.Config.PricingCacheDisabled {
		return
	}

	cacheFile := filepath.Join(ctx.Config.CachePath(), config.InfracostDir, "pricing.gob")
	c.cacheFile = cacheFile
	cache := loadCacheFromFile(cacheFile)
	c.cache = &cache

	c.mu = &sync.RWMutex{}
	c.objectLimit = 200
	if ctx.Config.PricingCacheObjectSize > 0 {
		c.objectLimit = ctx.Config.PricingCacheObjectSize
	}
}

func loadCacheFromFile(cacheFile string) map[uint64]gjson.Result {
	cache := make(map[uint64]gjson.Result)
	info, err := os.Stat(cacheFile)
	if err != nil {
		return cache
	}

	// if the cache is older than a day don't use it
	if info.ModTime().Before(time.Now().AddDate(0, 0, -1)) {
		return cache
	}

	f, err := os.Open(cacheFile)
	if err != nil {
		logging.Logger.WithError(err).Debugf("could not load cache file %s", cacheFile)
		return cache
	}

	var storedCache map[uint64]string
	err = gob.NewDecoder(f).Decode(&storedCache)
	if err != nil {
		logging.Logger.WithError(err).Debugf("failed to decode cache file %s", cacheFile)
		return cache
	}

	for k, raw := range storedCache {
		cache[k] = gjson.Parse(raw)
	}

	return cache
}

// FlushCache writes the in memory cache to the filesystem. This allows the cache
// to be persisted between runs. FlushCache should only be called once, at
// program termination.
func (c *PricingAPIClient) FlushCache() error {
	if c.cache == nil {
		return nil
	}

	logging.Logger.Debugf("writing %d objects to filesystem cache", len(c.cache))

	// we store the cache as a string instead of the gjson.Result so the size is
	// smaller on the filesystem and gob encoding doesn't have to work as hard.
	storedCache := map[uint64]string{}
	for k, result := range *c.cache {
		storedCache[k] = result.Raw
	}

	f, err := os.OpenFile(c.cacheFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	return gob.NewEncoder(f).Encode(storedCache)
}

func (c *PricingAPIClient) AddEvent(name string, env map[string]interface{}) error {
	if c.EventsDisabled {
		return nil
	}

	filtered := make(map[string]interface{})
	for k, v := range env {
		if _, ok := excludedEnv[k]; ok {
			continue
		}

		filtered[k] = v
	}

	d := map[string]interface{}{
		"event": name,
		"env":   filtered,
	}

	_, err := c.doRequest("POST", "/event", d)
	return err
}

func (c *PricingAPIClient) buildQuery(product *schema.ProductFilter, price *schema.PriceFilter) GraphQLQuery {
	v := map[string]interface{}{}
	v["productFilter"] = product
	v["priceFilter"] = price

	query := fmt.Sprintf(`
		query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {
			products(filter: $productFilter) {
				prices(filter: $priceFilter) {
					priceHash
					%s
				}
			}
		}
	`, c.Currency)

	return GraphQLQuery{query, v}
}

// BatchRequests batches all the queries for these resources so we can use less GraphQL requests
// Use PriceQueryKeys to keep track of which query maps to which sub-resource and price component.
func (c *PricingAPIClient) BatchRequests(resources []*schema.Resource, batchSize int) []BatchRequest {
	reqs := make([]BatchRequest, 0)

	keys := make([]PriceQueryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, r := range resources {
		for _, component := range r.CostComponents {
			keys = append(keys, PriceQueryKey{r, component})
			queries = append(queries, c.buildQuery(component.ProductFilter, component.PriceFilter))
		}

		for _, subresource := range r.FlattenedSubResources() {
			for _, component := range subresource.CostComponents {
				keys = append(keys, PriceQueryKey{subresource, component})
				queries = append(queries, c.buildQuery(component.ProductFilter, component.PriceFilter))
			}
		}
	}

	for i := 0; i < len(queries); i += batchSize {
		keysEnd := int64(math.Min(float64(i+batchSize), float64(len(keys))))
		queriesEnd := int64(math.Min(float64(i+batchSize), float64(len(queries))))

		reqs = append(reqs, BatchRequest{keys[i:keysEnd], queries[i:queriesEnd]})
	}

	return reqs
}

type pricingQuery struct {
	hash  uint64
	query GraphQLQuery

	result gjson.Result
}

// PerformRequest sends a batch request to the Pricing API endpoint to fetch
// pricing details for the provided queries. It optimizes the API call by
// checking a local cache for previous results. If the results of a given query
// are cached, they are used directly; otherwise, a request to the API is made.
func (c *PricingAPIClient) PerformRequest(req BatchRequest) ([]PriceQueryResult, error) {
	log.Debugf("Getting pricing details for %d cost components from %s", len(req.queries), c.endpoint)
	res := make([]PriceQueryResult, len(req.keys))
	for i, key := range req.keys {
		res[i].PriceQueryKey = key
	}

	queries := make([]pricingQuery, len(req.queries))
	for i, query := range req.queries {
		key, err := hashstructure.Hash(query, hashstructure.FormatV2, nil)
		if err != nil {
			logging.Logger.WithError(err).Debugf("failed to hash query %s will use nil hash", query)
		}

		queries[i] = pricingQuery{
			hash:  key,
			query: query,
		}
	}

	// first filter any queries that have been stored in the cache. We don't need to
	// send requests for these as we already have the results in memory.
	var serverQueries []pricingQuery
	if c.cache == nil {
		serverQueries = queries
	} else {
		var hit int
		for i, query := range queries {
			c.mu.RLock()
			v, ok := (*c.cache)[query.hash]
			c.mu.RUnlock()
			if ok {
				logging.Logger.Debugf("cache hit for query hash: %d", query.hash)
				hit++
				res[i] = PriceQueryResult{
					PriceQueryKey: req.keys[i],
					Result:        v,
					filled:        true,
				}
			} else {
				serverQueries = append(serverQueries, query)
			}
		}

		logging.Logger.Debugf("%d/%d queries were built from cache", hit, len(queries))
	}

	// now we deduplicate the queries, ensuring that a request for a price only happens once.
	var deduplicatedServerQueries []pricingQuery
	seenQueries := map[uint64]bool{}
	for _, query := range serverQueries {
		if seenQueries[query.hash] {
			continue
		}

		deduplicatedServerQueries = append(deduplicatedServerQueries, query)
		seenQueries[query.hash] = true
	}

	// send the deduplicated queries to the pricing API to fetch live prices.
	rawQueries := make([]GraphQLQuery, len(deduplicatedServerQueries))
	for i, query := range deduplicatedServerQueries {
		rawQueries[i] = query.query
	}
	resultsFromServer, err := c.doQueries(rawQueries)
	if err != nil {
		return []PriceQueryResult{}, err
	}

	// if the cache is enabled lets store each pricing result returned in the cache.
	if c.cache != nil {
		for i, query := range deduplicatedServerQueries {
			c.mu.RLock()
			keyLength := len(*c.cache)
			c.mu.RUnlock()
			if keyLength >= c.objectLimit {
				logging.Logger.Debugf("cache is at object limit of %d, will not add any additional keys to the cache", keyLength)
				continue
			}

			if len(resultsFromServer)-1 >= i {
				c.mu.Lock()
				(*c.cache)[query.hash] = resultsFromServer[i]
				c.mu.Unlock()
			}
		}
	}

	// now lets match the results from the server to their initial deduplicated queries.
	for i, result := range resultsFromServer {
		deduplicatedServerQueries[i].result = result
	}

	// Then we match deduplicated server queries to the initial list using the unique
	// query hash to tie a query to it's deduped query.
	resultMap := make(map[uint64]gjson.Result, len(deduplicatedServerQueries))
	for _, query := range deduplicatedServerQueries {
		resultMap[query.hash] = query.result
	}

	for i, query := range serverQueries {
		serverQueries[i].result = resultMap[query.hash]
	}

	// finally let's use the server queries to fill any results that haven't been
	// already populated from the cache.
	var x int
	for i, re := range res {
		if !re.filled {
			res[i].Result = serverQueries[x].result
			x++
		}
	}

	return res, nil
}
