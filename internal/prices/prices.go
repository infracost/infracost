package prices

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/ghodss/yaml"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var defaultK8sConfig = k8sPrices{defaults: gjson.Parse(`
{
	"cpu": 0.031611,
	"ram": 0.004237
}
`)}

type Pricer interface {
	GetPrices(r *schema.Resource) error
	GetPricesConcurrent(resources []*schema.Resource) error
}

func NewPricerForProject(ctx *config.ProjectContext, project *schema.Project) Pricer {
	apiFetcher := NewAPIPricer(ctx.RunContext)

	if project.Metadata == nil {
		return apiFetcher
	}

	if project.Metadata.Type == "k8s" {
		return NewStaticFilePricer(ctx.ProjectConfig.K8sFile)
	}

	return apiFetcher
}

func PopulatePrices(ctx *config.ProjectContext, project *schema.Project) error {
	pricer := NewPricerForProject(ctx, project)

	err := pricer.GetPricesConcurrent(project.AllResources())
	if err != nil {
		return err
	}

	return nil
}

type k8sPrices struct {
	defaults gjson.Result
	matches  []priceMatcher
}

func (k *k8sPrices) UnmarshalJSON(bytes []byte) error {
	var out k8sConfig
	err := json.Unmarshal(bytes, &out)
	if err != nil {
		return err
	}

	if len(out.Defaults) == 0 {
		return errors.New("invalid k8s config file, required field 'defaults' missing")
	}
	matches := make([]priceMatcher, len(out.Matchers))
	for i, m := range out.Matchers {
		labels := make(map[string]*regexp.Regexp, len(m.Match.Labels))
		for k, v := range m.Match.Labels {
			compile, err := regexp.Compile(v)
			if err != nil {
				return err
			}

			labels[k] = compile
		}

		matcher := priceMatcher{
			labels:   labels,
			contents: gjson.ParseBytes(m.Prices),
		}

		if m.Match.Name != nil {
			compile, err := regexp.Compile(*m.Match.Name)
			if err != nil {
				return err
			}

			matcher.name = compile
		}

		if m.Match.Namespace != nil {
			compile, err := regexp.Compile(*m.Match.Namespace)
			if err != nil {
				return err
			}

			matcher.namespace = compile
		}

		matches[i] = matcher
	}

	k.defaults = gjson.ParseBytes(out.Defaults)
	k.matches = matches

	return nil
}

type k8sConfig struct {
	Defaults json.RawMessage `json:"defaults"`
	Matchers []match         `json:"matchers"`
}

type match struct {
	Match  pattern         `json:"match"`
	Prices json.RawMessage `json:"prices"`
}

type pattern struct {
	Labels    map[string]string `json:"labels"`
	Name      *string           `json:"name"`
	Namespace *string           `json:"namespace"`
}

type priceMatcher struct {
	labels    map[string]*regexp.Regexp
	name      *regexp.Regexp
	namespace *regexp.Regexp
	contents  gjson.Result
}

func (p priceMatcher) match(r *schema.Resource) gjson.Result {
	for k, s := range p.labels {
		v, ok := r.Tags[fmt.Sprintf("labels.%s", k)]
		if !ok {
			return gjson.Result{}
		}

		if !s.MatchString(v) {
			return gjson.Result{}
		}
	}

	if p.name != nil {
		v, ok := r.Tags["name"]
		if !ok {
			return gjson.Result{}
		}

		if !p.name.MatchString(v) {
			return gjson.Result{}
		}
	}

	if p.namespace != nil {
		v, ok := r.Tags["namespace"]
		if !ok {
			return gjson.Result{}
		}

		if !p.namespace.MatchString(v) {
			return gjson.Result{}
		}
	}

	return p.contents
}

type StaticFilePricer struct {
	prices k8sPrices
}

func NewStaticFilePricer(path string) StaticFilePricer {
	if path == "" {
		return StaticFilePricer{prices: defaultK8sConfig}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		logging.Logger.WithError(err).Errorf("could not read static pricing file %s using static defaults", path)
		return StaticFilePricer{prices: defaultK8sConfig}
	}

	var prices k8sPrices
	err = yaml.Unmarshal(b, &prices)
	if err != nil {
		logging.Logger.WithError(err).Errorf("could not read k8s prices config %s using static defaults", path)
		return StaticFilePricer{prices: defaultK8sConfig}
	}

	return StaticFilePricer{prices: prices}
}

func (s StaticFilePricer) GetPrices(r *schema.Resource) error {
	if r.IsSkipped {
		return nil
	}

	prices := s.prices.defaults
	for _, m := range s.prices.matches {
		if res := m.match(r); res.Exists() {
			prices = res
			break
		}
	}

	return s.getPrices(r, prices)
}

func (s StaticFilePricer) getPrices(r *schema.Resource, prices gjson.Result) error {
	for _, sub := range r.SubResources {
		err := s.getPrices(sub, prices)
		if err != nil {
			//TODO
		}
	}

	for _, component := range r.CostComponents {
		if component.ProductFilter == nil || component.ProductFilter.Sku == nil {
			continue
		}

		price := prices.Get(*component.ProductFilter.Sku)
		if !price.Exists() {
			continue
		}

		component.SetPrice(decimal.NewFromFloat(price.Float()))
	}

	return nil
}

func (s StaticFilePricer) GetPricesConcurrent(resources []*schema.Resource) error {
	return getPricesConcurrent(resources, s.GetPrices)
}

type APIPricer struct {
	Client  *apiclient.PricingAPIClient
	Context *config.RunContext
}

func NewAPIPricer(ctx *config.RunContext) APIPricer {
	return APIPricer{
		Client:  apiclient.NewPricingAPIClient(ctx),
		Context: ctx,
	}
}

func (a APIPricer) GetPrices(r *schema.Resource) error {
	if r.IsSkipped {
		return nil
	}

	results, err := a.Client.RunQueries(r)
	if err != nil {
		return err
	}

	for _, r := range results {
		a.setCostComponentPrice(r.Resource, r.CostComponent, r.Result)
	}

	return nil
}

func (a APIPricer) GetPricesConcurrent(resources []*schema.Resource) error {
	return getPricesConcurrent(resources, a.GetPrices)
}

func (a APIPricer) setCostComponentPrice(r *schema.Resource, c *schema.CostComponent, res gjson.Result) {
	currency := "USD"
	if a.Context.Config.Currency != "" {
		currency = a.Context.Config.Currency
	}

	var p decimal.Decimal

	if c.CustomPrice() != nil {
		logging.Logger.Debugf("Using user-defined custom price %v for %s %s.", *c.CustomPrice(), r.Name, c.Name)
		c.SetPrice(*c.CustomPrice())
		return
	}

	products := res.Get("data.products").Array()
	if len(products) == 0 {
		if c.IgnoreIfMissingPrice {
			logging.Logger.Debugf("No products found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		logging.Logger.Warnf("No products found for %s %s, using 0.00", r.Name, c.Name)
		a.setResourceWarningEvent(r, "No products found")
		c.SetPrice(decimal.Zero)
		return
	}

	if len(products) > 1 {
		logging.Logger.Debugf("Multiple products found for %s %s, filtering those with prices", r.Name, c.Name)
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
			logging.Logger.Debugf("No prices found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		logging.Logger.Warnf("No prices found for %s %s, using 0.00", r.Name, c.Name)
		a.setResourceWarningEvent(r, "No prices found")
		c.SetPrice(decimal.Zero)
		return
	}

	if len(productsWithPrices) > 1 {
		logging.Logger.Warnf("Multiple products with prices found for %s %s, using the first product", r.Name, c.Name)
		a.setResourceWarningEvent(r, "Multiple products found")
	}

	prices := productsWithPrices[0].Get("prices").Array()
	if len(prices) > 1 {
		logging.Logger.Warnf("Multiple prices found for %s %s, using the first price", r.Name, c.Name)
		a.setResourceWarningEvent(r, "Multiple prices found")
	}

	var err error
	p, err = decimal.NewFromString(prices[0].Get(currency).String())
	if err != nil {
		logging.Logger.Warnf("Error converting price to '%v' (using 0.00)  '%v': %s", currency, prices[0].Get(currency).String(), err.Error())
		a.setResourceWarningEvent(r, "Error converting price")
		c.SetPrice(decimal.Zero)
		return
	}

	c.SetPrice(p)
	c.SetPriceHash(prices[0].Get("priceHash").String())
}

func (a APIPricer) setResourceWarningEvent(r *schema.Resource, msg string) {
	warnings := a.Context.GetResourceWarnings()
	if warnings == nil {
		warnings = make(map[string]map[string]int)
		a.Context.SetResourceWarnings(warnings)
	}

	resourceWarnings := warnings[r.ResourceType]
	if resourceWarnings == nil {
		resourceWarnings = make(map[string]int)
		warnings[r.ResourceType] = resourceWarnings
	}

	resourceWarnings[msg] += 1
}

// getPricesConcurrent gets the prices of all resources concurrently. Concurrency
// level is calculated using the following formula: max(min(4, numCPU * 4), 16)
func getPricesConcurrent(resources []*schema.Resource, f func(r *schema.Resource) error) error {
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
				err := f(r)
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
