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
	batchSize = 5
)

// notFoundData represents a single price not found entry
type notFoundData struct {
	ResourceType  string
	ResourceNames []string
	Count         int
}

// PriceFetcher provides a thread-safe way to aggregate 'price not found'
// data. This is used to provide a summary of missing prices at the end of a run.
// It should be used as a singleton which is shared across the application.
type PriceFetcher struct {
	resources  map[string]*notFoundData
	components map[string]int
	mux        *sync.RWMutex
	client     *apiclient.PricingAPIClient
	runCtx     *config.RunContext
}

func NewPriceFetcher(ctx *config.RunContext) *PriceFetcher {
	return &PriceFetcher{
		resources:  make(map[string]*notFoundData),
		components: make(map[string]int),
		mux:        &sync.RWMutex{},
		runCtx:     ctx,
		client:     apiclient.NewPricingAPIClient(ctx),
	}
}

// addNotFoundResult adds an instance of a missing price to the aggregator.
func (p *PriceFetcher) addNotFoundResult(result apiclient.PriceQueryResult) {
	p.mux.Lock()
	defer p.mux.Unlock()

	variables := result.Query.Variables
	b, _ := json.MarshalIndent(variables, "     ", " ")

	logging.Logger.Debug().Msgf("No products found for %s %s\n     %s", result.Resource.Name, result.CostComponent.Name, string(b))

	resource := result.Resource

	key := resource.BaseResourceType()
	name := resource.BaseResourceName()

	if entry, exists := p.resources[key]; exists {
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
		p.resources[key] = &notFoundData{
			ResourceType:  key,
			ResourceNames: []string{name},
			Count:         1,
		}
	}

	// build a key for the component, this is used to aggregate the number of
	// missing prices by cost component and resource type. The key is in the
	// format: resource_type.cost_component_name.
	componentName := strings.ToLower(result.CostComponent.Name)
	pieces := strings.Split(componentName, "(")
	if len(pieces) > 1 {
		componentName = strings.TrimSpace(pieces[0])
	}
	componentKey := fmt.Sprintf("%s.%s", key, strings.ReplaceAll(componentName, " ", "_"))

	if entry, exists := p.components[componentKey]; exists {
		entry++
		p.components[componentKey] = entry
	} else {
		p.components[componentKey] = 1

	}

	result.CostComponent.SetPriceNotFound()
}

// MissingPricesComponents returns a map of missing prices by component name, component
// names are in the format: resource_type.cost_component_name.
func (p *PriceFetcher) MissingPricesComponents() []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	var result []string
	for key, count := range p.components {
		for i := 0; i < count; i++ {
			result = append(result, key)
		}
	}

	return result
}

// MissingPricesLen returns the number of missing prices.
func (p *PriceFetcher) MissingPricesLen() int {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return len(p.resources)
}

// LogWarnings writes the PriceFetcher prices to the application log. If the log level is
// above the debug level we also include resource names the log output.
func (p *PriceFetcher) LogWarnings() {
	p.mux.RLock()
	defer p.mux.RUnlock()
	if len(p.resources) == 0 {
		return
	}

	var data []*notFoundData
	for _, v := range p.resources {
		data = append(data, v)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Count > data[j].Count
	})

	level, _ := zerolog.ParseLevel(p.runCtx.Config.LogLevel)
	includeResourceNames := level <= zerolog.DebugLevel

	s := strings.Builder{}
	warningPad := strings.Repeat(" ", 5)
	resourcePad := strings.Repeat(" ", 3)
	for i, v := range data {
		priceDesc := "price"
		if v.Count > 1 {
			priceDesc = "prices"
		}

		resourceDesc := "resource"
		if len(v.ResourceNames) > 1 {
			resourceDesc = "resources"
		}

		formattedResourceMsg := ui.FormatIfNotCI(p.runCtx, ui.WarningString, v.ResourceType)
		msg := fmt.Sprintf("%d %s %s missing across %d %s\n", v.Count, formattedResourceMsg, priceDesc, len(v.ResourceNames), resourceDesc)

		// pad the next warning line so that it appears inline with the last warning.
		if i > 0 {
			msg = fmt.Sprintf("%s%s", warningPad, msg)
		}
		s.WriteString(msg)

		if includeResourceNames {
			for _, resourceName := range v.ResourceNames {
				name := ui.FormatIfNotCI(p.runCtx, ui.UnderlineString, resourceName)
				s.WriteString(fmt.Sprintf("%s%s- %s \n", warningPad, resourcePad, name))
			}
		}
	}

	logging.Logger.Warn().Msg(s.String())
}

func (p *PriceFetcher) PopulatePrices(project *schema.Project) error {
	resources := project.AllResources()

	err := p.getPricesConcurrent(resources)
	if err != nil {
		return err
	}
	return nil
}

// getPricesConcurrent gets the prices of all resources concurrently.
// Concurrency level is calculated using the following formula:
// max(min(4, numCPU * 4), 16)
func (p *PriceFetcher) getPricesConcurrent(resources []*schema.Resource) error {
	// Set the number of workers
	numWorkers := 4
	numCPU := runtime.NumCPU()
	if numCPU*4 > numWorkers {
		numWorkers = numCPU * 4
	}
	if numWorkers > 16 {
		numWorkers = 16
	}

	reqs := p.client.BatchRequests(resources, batchSize)

	numJobs := len(reqs)
	jobs := make(chan apiclient.BatchRequest, numJobs)
	resultErrors := make(chan error, numJobs)

	// Fire up the workers
	for i := 0; i < numWorkers; i++ {
		go func(jobs <-chan apiclient.BatchRequest, resultErrors chan<- error) {
			for req := range jobs {
				err := p.getPrices(req)
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

func (p *PriceFetcher) getPrices(req apiclient.BatchRequest) error {
	results, err := p.client.PerformRequest(req)
	if err != nil {
		return err
	}

	for _, r := range results {
		p.setCostComponentPrice(r)
	}

	return nil
}

func (p *PriceFetcher) setCostComponentPrice(result apiclient.PriceQueryResult) {
	currency := p.client.Currency

	var pp decimal.Decimal
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

		p.addNotFoundResult(result)
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

		p.addNotFoundResult(result)
		return
	}

	if len(productsWithPrices) > 1 {
		logging.Logger.Debug().Msgf("Multiple products with prices found for %s %s, using the first product", result.Resource.Name, result.CostComponent.Name)
	}

	prices := productsWithPrices[0].Get("prices").Array()
	if len(prices) > 1 {
		logging.Logger.Debug().Msgf("Multiple prices found for %s %s, using the first price", result.Resource.Name, result.CostComponent.Name)
	}

	var err error
	pp, err = decimal.NewFromString(prices[0].Get(currency).String())
	if err != nil {
		logging.Logger.Warn().Msgf("Error converting price to '%v' (using 0.00)  '%v': %s", currency, prices[0].Get(currency).String(), err.Error())
		result.CostComponent.SetPrice(decimal.Zero)
		return
	}

	result.CostComponent.SetPrice(pp)
	result.CostComponent.SetPriceHash(prices[0].Get("priceHash").String())
}
