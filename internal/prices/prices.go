package prices

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var (
	batchSize          = 5
	warningMu          = &sync.Mutex{}
	NotFoundComponents = &notFound{
		data: make(map[string]*notFoundData),
		mux:  &sync.RWMutex{},
	}
)

// notFoundData represents a single price not found entry
type notFoundData struct {
	ResourceType  string
	ResourceNames []string
	Count         int
}

// notFound provides a thread-safe way to aggregate 'price not found'
// data. This is used to provide a summary of missing prices at the end of a run.
// It should be used as a singleton which is shared across the application.
type notFound struct {
	data map[string]*notFoundData
	mux  *sync.RWMutex
}

// Add adds an instance of a missing price to the aggregator.
func (p *notFound) Add(r *schema.Resource) {
	p.mux.Lock()
	defer p.mux.Unlock()

	key := r.BaseResourceType()
	name := r.BaseResourceName()

	if entry, exists := p.data[key]; exists {
		entry.Count++

		var found bool
		for _, resourceName := range entry.ResourceNames {
			if resourceName == name {
				found = true
				break
			}
		}

		if !found {
			entry.ResourceNames = append(entry.ResourceNames, name)
		}
	} else {
		p.data[key] = &notFoundData{
			ResourceType:  key,
			ResourceNames: []string{name},
			Count:         1,
		}
	}
}

// Len returns the number of missing prices.
func (p *notFound) Len() int {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return len(p.data)
}

// Log writes the notFound prices to the application log. If the log level is
// above the debug level we also include resource names the log output.
func (p *notFound) Log(ctx *config.RunContext) {
	p.mux.RLock()
	defer p.mux.RUnlock()
	if len(p.data) == 0 {
		return
	}

	var keys []string
	for k := range p.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	level, _ := zerolog.ParseLevel(ctx.Config.LogLevel)
	includeResourceNames := level <= zerolog.DebugLevel

	s := strings.Builder{}
	warningPad := strings.Repeat(" ", 5)
	resourcePad := strings.Repeat(" ", 3)
	for i, k := range keys {
		v := p.data[k]
		priceDesc := "price"
		if v.Count > 1 {
			priceDesc = "prices"
		}

		resourceDesc := "resource"
		if len(v.ResourceNames) > 1 {
			resourceDesc = "resources"
		}

		formattedResourceMsg := ui.FormatIfNotCI(ctx, ui.WarningString, v.ResourceType)
		msg := fmt.Sprintf("%d %s %s missing across %d %s\n", v.Count, formattedResourceMsg, priceDesc, len(v.ResourceNames), resourceDesc)

		// pad the next warning line so that it appears inline with the last warning.
		if i > 0 {
			msg = fmt.Sprintf("%s%s", warningPad, msg)
		}
		s.WriteString(msg)

		if includeResourceNames {
			for _, resourceName := range v.ResourceNames {
				name := ui.FormatIfNotCI(ctx, ui.UnderlineString, resourceName)
				s.WriteString(fmt.Sprintf("%s%s- %s \n", warningPad, resourcePad, name))
			}
		}
	}

	logging.Logger.Warn().Msg(s.String())
}

func PopulatePrices(ctx *config.RunContext, project *schema.Project) error {
	resources := project.AllResources()

	c := apiclient.GetPricingAPIClient(ctx)

	err := GetPricesConcurrent(ctx, c, resources)
	if err != nil {
		return err
	}
	return nil
}

// GetPricesConcurrent gets the prices of all resources concurrently.
// Concurrency level is calculated using the following formula:
// max(min(4, numCPU * 4), 16)
func GetPricesConcurrent(ctx *config.RunContext, c *apiclient.PricingAPIClient, resources []*schema.Resource) error {
	// Set the number of workers
	numWorkers := 4
	numCPU := runtime.NumCPU()
	if numCPU*4 > numWorkers {
		numWorkers = numCPU * 4
	}
	if numWorkers > 16 {
		numWorkers = 16
	}

	reqs := c.BatchRequests(resources, batchSize)

	numJobs := len(reqs)
	jobs := make(chan apiclient.BatchRequest, numJobs)
	resultErrors := make(chan error, numJobs)

	// Fire up the workers
	for i := 0; i < numWorkers; i++ {
		go func(jobs <-chan apiclient.BatchRequest, resultErrors chan<- error) {
			for req := range jobs {
				err := GetPrices(ctx, c, req)
				resultErrors <- err
			}
		}(jobs, resultErrors)
	}

	// Feed the workers the jobs of getting prices
	for _, r := range reqs {
		jobs <- r
	}

	// Get the result of the jobs
	for i := 0; i < numJobs; i++ {
		err := <-resultErrors
		if err != nil {
			return err
		}
	}
	return nil
}

func GetPrices(ctx *config.RunContext, c *apiclient.PricingAPIClient, req apiclient.BatchRequest) error {
	results, err := c.PerformRequest(req)
	if err != nil {
		return err
	}

	for _, r := range results {
		setCostComponentPrice(ctx, c.Currency, r)
	}

	return nil
}

func setCostComponentPrice(ctx *config.RunContext, currency string, result apiclient.PriceQueryResult) {
	var p decimal.Decimal
	if result.CostComponent.CustomPrice() != nil {
		logging.Logger.Debug().Msgf("Using user-defined custom price %v for %s %s.", *result.CostComponent.CustomPrice(), result.Resource.Name, result.CostComponent.Name)
		result.CostComponent.SetPrice(*result.CostComponent.CustomPrice())
		return
	}

	products := result.Result.Get("data.products").Array()
	if len(products) == 0 {
		if result.CostComponent.IgnoreIfMissingPrice {
			logging.Logger.Debug().Msgf("No products found for %s %s, ignoring since IgnoreIfMissingPrice is set.", result.Resource.Name, result.CostComponent.Name)
			result.Resource.RemoveCostComponent(result.CostComponent)
			return
		}

		debugNoProductsFound(result)
		NotFoundComponents.Add(result.Resource)
		setResourceWarningEvent(ctx, result.Resource, "No products found")
		result.CostComponent.SetPriceNotFound()
		return
	}

	if len(products) > 1 {
		logging.Logger.Debug().Msgf("Multiple products found for %s %s, filtering those with prices", result.Resource.Name, result.CostComponent.Name)
	}

	// Some resources may have identical records in CPAPI for the same product
	// filters, several products are always returned and they can only be
	// distinguished by their prices. However if we pick the first product it may not
	// have the price due to price filter and the lookup fails. Filtering the
	// products with prices helps to solve that.
	var productsWithPrices []gjson.Result
	for _, product := range products {
		if len(product.Get("prices").Array()) > 0 {
			productsWithPrices = append(productsWithPrices, product)
		}
	}

	if len(productsWithPrices) == 0 {
		if result.CostComponent.IgnoreIfMissingPrice {
			logging.Logger.Debug().Msgf("No prices found for %s %s, ignoring since IgnoreIfMissingPrice is set.", result.Resource.Name, result.CostComponent.Name)
			result.Resource.RemoveCostComponent(result.CostComponent)
			return
		}

		debugNoProductsFound(result)
		NotFoundComponents.Add(result.Resource)
		setResourceWarningEvent(ctx, result.Resource, "No prices found")
		result.CostComponent.SetPriceNotFound()
		return
	}

	if len(productsWithPrices) > 1 {
		logging.Logger.Debug().Msgf("Multiple products with prices found for %s %s, using the first product", result.Resource.Name, result.CostComponent.Name)
		setResourceWarningEvent(ctx, result.Resource, "Multiple products found")
	}

	prices := productsWithPrices[0].Get("prices").Array()
	if len(prices) > 1 {
		logging.Logger.Warn().Msgf("Multiple prices found for %s %s, using the first price", result.Resource.Name, result.CostComponent.Name)
		setResourceWarningEvent(ctx, result.Resource, "Multiple prices found")
	}

	var err error
	p, err = decimal.NewFromString(prices[0].Get(currency).String())
	if err != nil {
		logging.Logger.Warn().Msgf("Error converting price to '%v' (using 0.00)  '%v': %s", currency, prices[0].Get(currency).String(), err.Error())
		setResourceWarningEvent(ctx, result.Resource, "Error converting price")
		result.CostComponent.SetPrice(decimal.Zero)
		return
	}

	result.CostComponent.SetPrice(p)
	result.CostComponent.SetPriceHash(prices[0].Get("priceHash").String())
}

func debugNoProductsFound(result apiclient.PriceQueryResult) {
	variables := result.Query.Variables
	b, _ := json.MarshalIndent(variables, "     ", " ")

	logging.Logger.Debug().Msgf("No products found for %s %s\n     %s", result.Resource.Name, result.CostComponent.Name, string(b))
}

func setResourceWarningEvent(ctx *config.RunContext, r *schema.Resource, msg string) {
	warningMu.Lock()
	defer warningMu.Unlock()

	warnings := ctx.GetResourceWarnings()
	if warnings == nil {
		warnings = make(map[string]map[string]int)
		ctx.SetResourceWarnings(warnings)
	}

	resourceWarnings := warnings[r.ResourceType]
	if resourceWarnings == nil {
		resourceWarnings = make(map[string]int)
		warnings[r.ResourceType] = resourceWarnings
	}

	resourceWarnings[msg] += 1
}
