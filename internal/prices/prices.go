package prices

import (
	"runtime"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func PopulatePrices(ctx *config.RunContext, project *schema.Project) error {
	resources := project.AllResources()

	c := apiclient.NewPricingAPIClient(ctx)

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
	numJobs := len(resources)
	jobs := make(chan *schema.Resource, numJobs)
	resultErrors := make(chan error, numJobs)

	// Fire up the workers
	for i := 0; i < numWorkers; i++ {
		go func(jobs <-chan *schema.Resource, resultErrors chan<- error) {
			for r := range jobs {
				err := GetPrices(ctx, c, r)
				resultErrors <- err
			}
		}(jobs, resultErrors)
	}

	// Feed the workers the jobs of getting prices
	for _, r := range resources {
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

func GetPrices(ctx *config.RunContext, c *apiclient.PricingAPIClient, r *schema.Resource) error {
	if r.IsSkipped {
		return nil
	}

	results, err := c.RunQueries(r)
	if err != nil {
		return err
	}

	for _, r := range results {
		setCostComponentPrice(ctx, c.Currency, r.Resource, r.CostComponent, r.Result)
	}

	return nil
}

func setCostComponentPrice(ctx *config.RunContext, currency string, r *schema.Resource, c *schema.CostComponent, res gjson.Result) {
	var p decimal.Decimal

	if c.CustomPrice() != nil {
		log.Debugf("Using user-defined custom price %v for %s %s.", *c.CustomPrice(), r.Name, c.Name)
		c.SetPrice(*c.CustomPrice())
		return
	}

	products := res.Get("data.products").Array()
	if len(products) == 0 {
		if c.IgnoreIfMissingPrice {
			log.Debugf("No products found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		log.Warnf("No products found for %s %s, using 0.00", r.Name, c.Name)
		setResourceWarningEvent(ctx, r, "No products found")
		c.SetPrice(decimal.Zero)
		return
	}

	if len(products) > 1 {
		log.Debugf("Multiple products found for %s %s, filtering those with prices", r.Name, c.Name)
	}

	// Some resources may have identical records in CPAPI for the same product
	// filters, several products are always returned and they can only be
	// distinguished by their prices. However if we pick the first product it may not
	// have the price due to price filter and the lookup fails. Filtering the
	// products with prices helps to solve that.
	productsWithPrices := []gjson.Result{}
	for _, product := range products {
		if len(product.Get("prices").Array()) > 0 {
			productsWithPrices = append(productsWithPrices, product)
		}
	}

	if len(productsWithPrices) == 0 {
		if c.IgnoreIfMissingPrice {
			log.Debugf("No prices found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		log.Warnf("No prices found for %s %s, using 0.00", r.Name, c.Name)
		setResourceWarningEvent(ctx, r, "No prices found")
		c.SetPrice(decimal.Zero)
		return
	}

	if len(productsWithPrices) > 1 {
		log.Warnf("Multiple products with prices found for %s %s, using the first product", r.Name, c.Name)
		setResourceWarningEvent(ctx, r, "Multiple products found")
	}

	prices := productsWithPrices[0].Get("prices").Array()
	if len(prices) > 1 {
		log.Warnf("Multiple prices found for %s %s, using the first price", r.Name, c.Name)
		setResourceWarningEvent(ctx, r, "Multiple prices found")
	}

	var err error
	p, err = decimal.NewFromString(prices[0].Get(currency).String())
	if err != nil {
		log.Warnf("Error converting price to '%v' (using 0.00)  '%v': %s", currency, prices[0].Get(currency).String(), err.Error())
		setResourceWarningEvent(ctx, r, "Error converting price")
		c.SetPrice(decimal.Zero)
		return
	}

	c.SetPrice(p)
	c.SetPriceHash(prices[0].Get("priceHash").String())
}

func setResourceWarningEvent(ctx *config.RunContext, r *schema.Resource, msg string) {
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
