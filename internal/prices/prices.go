package prices

import (
	"runtime"
	"sync"

	"github.com/infracost/infracost/internal/events"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func PopulatePrices(resources []*schema.Resource, concurrency int) error {
	q := NewGraphQLQueryRunner()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		summary := output.BuildResourceSummary(resources, output.ResourceSummaryOptions{
			IncludeUnsupportedProviders: true,
		})

		events.SendReport("resourceSummary", summary)
	}()

	err := GetPricesConcurrent(resources, q, concurrency)
	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}

// GetPricesConcurrent gets the prices of all resources concurrently. Concurrency level can be
// configured with --concurrency flag. It defaults to max(4, number of CPUs * 4).
func GetPricesConcurrent(resources []*schema.Resource, q QueryRunner, concurrency int) error {
	// Set the number of workers
	numWorkers := concurrency
	if numWorkers == 0 {
		// User did not specify the level of concurrency. Using default.
		numWorkers = 4
		numCPU := runtime.NumCPU()
		if numCPU*4 > numWorkers {
			numWorkers = numCPU * 4
		}
	}
	numJobs := len(resources)
	jobs := make(chan *schema.Resource, numJobs)
	resultErrors := make(chan error, numJobs)

	// Fire up the workers
	for i := 0; i < numWorkers; i++ {
		go func(jobs <-chan *schema.Resource, resultErrors chan<- error) {
			for r := range jobs {
				err := GetPrices(r, q)
				resultErrors <- err
			}
		}(jobs, resultErrors)
	}

	// Feed the workers the jobs of getting prices
	for _, r := range resources {
		jobs <- r
	}

	// Get the result errors of jobs
	for i := 0; i < numJobs; i++ {
		err := <-resultErrors
		if err != nil {
			return err
		}
	}
	return nil
}

func GetPrices(r *schema.Resource, q QueryRunner) error {
	if r.IsSkipped {
		return nil
	}
	results, err := q.RunQueries(r)
	if err != nil {
		return err
	}

	for _, r := range results {
		setCostComponentPrice(r.Resource, r.CostComponent, r.Result)
	}

	return nil
}

func setCostComponentPrice(r *schema.Resource, c *schema.CostComponent, res gjson.Result) {
	var p decimal.Decimal

	products := res.Get("data.products").Array()
	if len(products) == 0 {
		if c.IgnoreIfMissingPrice {
			log.Debugf("No products found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		log.Warnf("No products found for %s %s, using 0.00", r.Name, c.Name)
		c.SetPrice(decimal.Zero)
		return
	}
	if len(products) > 1 {
		log.Warnf("Multiple products found for %s %s, using the first product", r.Name, c.Name)
	}

	prices := products[0].Get("prices").Array()
	if len(prices) == 0 {
		if c.IgnoreIfMissingPrice {
			log.Debugf("No prices found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		log.Warnf("No prices found for %s %s, using 0.00", r.Name, c.Name)
		c.SetPrice(decimal.Zero)
		return
	}
	if len(prices) > 1 {
		log.Warnf("Multiple prices found for %s %s, using the first price", r.Name, c.Name)
	}

	var err error
	p, err = decimal.NewFromString(prices[0].Get("USD").String())
	if err != nil {
		log.Warnf("Error converting price (using 0.00) '%v': %s", prices[0].Get("USD").String(), err.Error())
		c.SetPrice(decimal.Zero)
		return
	}

	c.SetPrice(p)
	c.SetPriceHash(prices[0].Get("priceHash").String())
}
