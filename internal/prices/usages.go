package prices

import (
	"github.com/shopspring/decimal"
	"runtime"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

// PopulateActualCosts fetches cloud provider reported costs from the Infracost
// Cloud Usage API and adds corresponding cost components to the project's resources
func PopulateActualCosts(ctx *config.RunContext, project *schema.Project) error {
	resources := project.AllResources()

	c := apiclient.NewUsageAPIClient(ctx)

	err := popResourcesConcurrent(ctx, c, project, resources)
	if err != nil {
		return err
	}
	return nil
}

// popResourcesConcurrent gets the actual usage of all resources concurrently.
// Concurrency level is calculated using the following formula:
// max(min(4, numCPU * 4), 16)
func popResourcesConcurrent(ctx *config.RunContext, c *apiclient.UsageAPIClient, project *schema.Project, resources []*schema.Resource) error {
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
				err := popResourceActualCosts(ctx, c, project, r)
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

func popResourceActualCosts(ctx *config.RunContext, c *apiclient.UsageAPIClient, project *schema.Project, r *schema.Resource) error {
	if r.IsSkipped {
		return nil
	}

	vars := apiclient.ActualCostsQueryVariables{
		RepoURL:  ctx.VCSRepositoryURL(),
		Project:  project.Name,
		Address:  r.Name,
		Currency: c.Currency,
	}
	results, err := c.ListActualCosts(vars)

	if err != nil {
		return err
	}

	for _, actual := range results {
		monthlyCost, err := decimal.NewFromString(actual.MonthlyCost)
		if err != nil {
			break
		}

		monthlyQuantity, err := decimal.NewFromString(actual.MonthlyQuantity)
		if err != nil {
			break
		}
		price, err := decimal.NewFromString(actual.Price)
		if err != nil {
			break
		}

		cc := &schema.CostComponent{
			Name:            actual.Description,
			Unit:            actual.Unit,
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyCost:     &monthlyCost,
			MonthlyQuantity: &monthlyQuantity,
		}
		cc.SetPrice(price)

		r.CostComponents = append(r.CostComponents, cc)
	}
	return nil
}

// FetchUsageData fetches usage estimates derived from cloud provider reported usage
// from the Infracost Cloud Usage API for each supported resource in the project
func FetchUsageData(ctx *config.RunContext, project *schema.Project) (map[string]*schema.UsageData, error) {
	c := apiclient.NewUsageAPIClient(ctx)

	// gather all the CoreResource
	coreResources := make(map[string]schema.CoreResource)
	for _, rb := range project.PartialResources {
		if rb.CoreResource != nil {
			coreResources[rb.ResourceData.Address] = rb.CoreResource
		}
	}
	for _, rb := range project.PartialPastResources {
		if rb.CoreResource != nil {
			coreResources[rb.ResourceData.Address] = rb.CoreResource
		}
	}

	usageMap := make(map[string]*schema.UsageData, len(coreResources))
	// look up the usage for each core resource.
	for address, cr := range coreResources {
		// TODO: add concurrency
		usageKeys := make([]string, len(cr.UsageSchema()))
		for i, usageItem := range cr.UsageSchema() {
			usageKeys[i] = usageItem.Key
		}

		if len(usageKeys) > 0 {
			vars := apiclient.UsageQuantitiesQueryVariables{
				RepoURL:      ctx.VCSRepositoryURL(),
				Project:      project.Name,
				ResourceType: cr.CoreType(),
				Address:      address,
				UsageKeys:    usageKeys,
			}

			attributes, err := c.ListUsageQuantities(vars)
			if err != nil {
				return nil, err
			}

			usageMap[address] = &schema.UsageData{
				Address:    address,
				Attributes: attributes,
			}
		}
	}

	return usageMap, nil
}
