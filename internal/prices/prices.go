package prices

import (
	"runtime"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func PopulatePrices(ctx *config.ProjectContext, project *schema.Project) error {
	resources := project.AllResources()

	c := apiclient.NewPricingAPIClient(ctx.Config)

	err := GetPricesConcurrent(ctx, c, resources)
	if err != nil {
		return err
	}
	return nil
}

// GetPricesConcurrent gets the prices of all resources concurrently.
// Concurrency level is calculated using the following formula:
// max(min(4, numCPU * 4), 16)
func GetPricesConcurrent(ctx *config.ProjectContext, c *apiclient.PricingAPIClient, resources []*schema.Resource) error {
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

func GetPrices(ctx *config.ProjectContext, c *apiclient.PricingAPIClient, r *schema.Resource) error {
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

func setCostComponentPrice(ctx *config.ProjectContext, currency string, r *schema.Resource, c *schema.CostComponent, res gjson.Result) {
	var p decimal.Decimal
	
	logger := ctx.Logger.With().Str("resource", r.Name).Str("cost_component", c.Name).Logger()

	products := res.Get("data.products").Array()
	if len(products) == 0 {
		if c.IgnoreIfMissingPrice {
			logger.Debug().Msg("No products found, ignoring since IgnoreIfMissingPrice is set.")
			r.RemoveCostComponent(c)
			return
		}

		logger.Warn().Msg("No products found, using 0.00")
		c.SetPrice(decimal.Zero)
		return
	}
	if len(products) > 1 {
		logger.Warn().Msg("Multiple products found, using first product")
	}

	prices := products[0].Get("prices").Array()
	if len(prices) == 0 {
		if c.IgnoreIfMissingPrice {
			logger.Debug().Msg("No prices found, ignoring since IgnoreIfMissingPrice is set.")
			r.RemoveCostComponent(c)
			return
		}

		logger.Warn().Msg("No prices found, using 0.00")
		c.SetPrice(decimal.Zero)
		return
	}
	if len(prices) > 1 {
		logger.Warn().Msg("Multiple prices found, using the first price")
	}

	var err error
	p, err = decimal.NewFromString(prices[0].Get(currency).String())
	if err != nil {
		logger.Warn().Err(err).Msgf("Error converting price from %v to %v (using 0.00)", prices[0].Get(currency).String(), currency)
		c.SetPrice(decimal.Zero)
		return
	}

	c.SetPrice(p)
	c.SetPriceHash(prices[0].Get("priceHash").String())
}
