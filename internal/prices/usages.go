package prices

import (
	"runtime"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
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
	for range numJobs {
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
		RepoURL:              ctx.VCSRepositoryURL(),
		ProjectWithWorkspace: project.NameWithWorkspace(),
		Address:              r.Name,
		Currency:             c.Currency,
	}
	actualCostResults, err := c.ListActualCosts(vars)
	if actualCostResults == nil || err != nil {
		return err
	}

	for _, actualCost := range actualCostResults {
		actualCosts := &schema.ActualCosts{
			ResourceID:     actualCost.ResourceID,
			StartTimestamp: actualCost.StartTimestamp.UTC(),
			EndTimestamp:   actualCost.EndTimestamp.UTC(),
			CostComponents: make([]*schema.CostComponent, 0, len(actualCost.CostComponents)),
		}

		for _, actual := range actualCost.CostComponents {
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

			actualCosts.CostComponents = append(actualCosts.CostComponents, cc)
		}

		if len(actualCosts.CostComponents) > 0 {
			r.ActualCosts = append(r.ActualCosts, actualCosts)
		}
	}

	return nil
}

// chunkPartialResourcesWithUsage collects all partiral resources with a core resource usage schema
// into groups of the specified chunkSize
func chunkPartialResourcesWithUsage(resources []*schema.PartialResource, chunkSize int) [][]*schema.PartialResource {
	var usageResourceChunks [][]*schema.PartialResource
	var currentChunk []*schema.PartialResource
	for _, rb := range resources {
		if rb.CoreResource != nil && len(rb.CoreResource.UsageSchema()) > 0 {
			if len(currentChunk) >= chunkSize {
				usageResourceChunks = append(usageResourceChunks, currentChunk)
				currentChunk = nil
			}
			currentChunk = append(currentChunk, rb)
		}
	}
	if len(currentChunk) > 0 {
		usageResourceChunks = append(usageResourceChunks, currentChunk)
	}

	return usageResourceChunks
}

// FetchUsageData fetches usage estimates derived from cloud provider reported usage
// from the Infracost Cloud Usage API for each supported resource in the project
func FetchUsageData(ctx *config.RunContext, project *schema.Project) (schema.UsageMap, error) {
	c := apiclient.NewUsageAPIClient(ctx)

	// Set the number of workers
	numWorkers := 4
	numCPU := runtime.NumCPU()
	if numCPU*4 > numWorkers {
		numWorkers = numCPU * 4
	}
	if numWorkers > 16 {
		numWorkers = 16
	}

	// gather all the CoreResource into chunks
	usageResourceChunks := chunkPartialResourcesWithUsage(project.AllPartialResources(), 10)

	usageMap := make(map[string]*schema.UsageData, len(usageResourceChunks)*10)

	numJobs := len(usageResourceChunks)
	jobs := make(chan []*schema.PartialResource, numJobs)
	responses := make(chan batchResponse, numJobs)

	// Fire up the workers
	for i := 0; i < numWorkers; i++ {
		go func(jobs <-chan []*schema.PartialResource, responses chan<- batchResponse) {
			for req := range jobs {
				res := fetchUsageDataBatch(ctx, c, project, req)
				responses <- res
			}
		}(jobs, responses)
	}

	// Feed the workers the jobs
	for _, r := range usageResourceChunks {
		jobs <- r
	}

	// Get the result of the jobs
	for range numJobs {
		res := <-responses
		if res.Error != nil {
			return schema.NewUsageMap(usageMap), res.Error
		}

		for _, ud := range res.UsageData {
			usageMap[ud.Address] = ud
		}
	}

	return schema.NewUsageMap(usageMap), nil
}

type batchResponse struct {
	UsageData []*schema.UsageData
	Error     error
}

func fetchUsageDataBatch(ctx *config.RunContext, c *apiclient.UsageAPIClient, project *schema.Project, resources []*schema.PartialResource) batchResponse {
	chunkedQueryVars := make([]*apiclient.UsageQuantitiesQueryVariables, 0, len(resources))
	for _, partialResource := range resources {
		var usageParams []schema.UsageParam
		if crWithUsageParams, ok := partialResource.CoreResource.(schema.CoreResourceWithUsageParams); ok {
			usageParams = crWithUsageParams.UsageEstimationParams()
		}

		queryVars := apiclient.UsageQuantitiesQueryVariables{
			RepoURL:              ctx.VCSRepositoryURL(),
			ProjectWithWorkspace: project.NameWithWorkspace(),
			ResourceType:         partialResource.CoreResource.CoreType(),
			Address:              partialResource.Address,
			UsageKeys:            flattenUsageKeys(partialResource.CoreResource.UsageSchema()),
			UsageParams:          usageParams,
		}
		chunkedQueryVars = append(chunkedQueryVars, &queryVars)
	}

	ud, err := c.ListUsageQuantities(chunkedQueryVars)

	return batchResponse{
		UsageData: ud,
		Error:     err,
	}
}

// UploadCloudResourceIDs sends the project scoped cloud resource ids to the Usage API, so they can be used
// to provide usage estimates.
func UploadCloudResourceIDs(ctx *config.RunContext, project *schema.Project) error {
	c := apiclient.NewUsageAPIClient(ctx)

	var resourceIDs []apiclient.ResourceIDAddress
	for _, partial := range project.AllPartialResources() {
		for _, resourceID := range partial.CloudResourceIDs {
			resourceIDs = append(resourceIDs, apiclient.ResourceIDAddress{
				Address:    partial.Address,
				ResourceID: resourceID},
			)
		}
	}

	vars := apiclient.CloudResourceIDVariables{
		RepoURL:              ctx.VCSRepositoryURL(),
		ProjectWithWorkspace: project.NameWithWorkspace(),
		ResourceIDAddresses:  resourceIDs,
	}

	err := c.UploadCloudResourceIDs(vars)
	if err != nil {
		return err
	}

	return nil
}

func flattenUsageKeys(usageSchema []*schema.UsageItem) []string {
	usageKeys := make([]string, 0, len(usageSchema))
	for _, usageItem := range usageSchema {
		if usageItem.ValueType == schema.SubResourceUsage {
			ru := usageItem.DefaultValue.(*usage.ResourceUsage)
			// recursively flatten any nested keys, then add them to the current list
			for _, nestedKey := range flattenUsageKeys(ru.Items) {
				usageKeys = append(usageKeys, usageItem.Key+"."+nestedKey)
			}
		} else {
			usageKeys = append(usageKeys, usageItem.Key)
		}
	}

	return usageKeys
}
