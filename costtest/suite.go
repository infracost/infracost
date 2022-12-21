package costtest

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/output"
)

var (
	defaultSuite *Suite

	saveCoverage = flag.Bool("costtest.coverfile", false, "save costtest coverage report to file")
	showCoverage = flag.Bool("costtest.cover", false, "show costtest coverage in a simple table output")
	porcelain    = flag.Bool("costtest.porcelain", false, "print machine understandable output")
)

type Out struct {
	Currency string `json:"currency"`
	Projects []struct {
		Name      string `json:"name"`
		Breakdown struct {
			Resources []*Resource
		} `json:"breakdown"`
		Summary Summary `json:"summary"`
	} `json:"projects"`
}

type Summary struct {
	TotalDetectedResources    int            `json:"totalDetectedResources"`
	TotalSupportedResources   int            `json:"totalSupportedResources"`
	TotalUnsupportedResources int            `json:"totalUnsupportedResources"`
	TotalUsageBasedResources  int            `json:"totalUsageBasedResources"`
	TotalNoPriceResources     int            `json:"totalNoPriceResources"`
	UnsupportedResourceCounts map[string]int `json:"unsupportedResourceCounts"`
	NoPriceResourceCounts     map[string]int `json:"noPriceResourceCounts"`
}

type Resource struct {
	Metadata struct {
		Filename string `json:"filename"`
	} `json:"metadata"`

	Name           string         `json:"name"`
	HourlyCost     float64        `json:"hourlyCost,string"`
	MonthlyCost    float64        `json:"monthlyCost,string"`
	CostComponents CostComponents `json:"costComponents"`
	SubResources   Resources      `json:"subresources"`

	covered bool
	parent  *Resource
}

func (r *Resource) setCovered() {
	if r.parent != nil {
		r.parent.setCovered()
	}

	r.covered = true
}

type Resources []Resource

func (r Resources) Get(pattern string) Resource {
	for _, resource := range r {
		if resource.Name == pattern {
			return resource
		}
	}

	return Resource{}
}

type CostComponents []CostComponent

func (c CostComponents) Get(pattern string) CostComponent {
	for _, component := range c {
		if component.Name == pattern {
			return component
		}
	}

	return CostComponent{}
}

type CostComponent struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	HourlyQuantity  float64 `json:"hourlyQuantity,string"`
	MonthlyQuantity float64 `json:"monthlyQuantity,string"`
	Price           float64 `json:"price,string"`
	HourlyCost      float64 `json:"hourlyCost,string"`
	MonthlyCost     float64 `json:"monthlyCost,string"`
}

type InitOptions struct {
	Path string

	fromInit bool
}

type Suite struct {
	Resources []*Resource
	Summary   Summary

	parent   *Suite
	dir      string
	currency string
}

func NewSuite(ops InitOptions) *Suite {
	_, file, _, _ := runtime.Caller(1)
	if ops.fromInit {
		_, file, _, _ = runtime.Caller(2)
	}
	projectDir := filepath.Join(filepath.Dir(file), ops.Path)

	cmd := exec.Command("infracost", "breakdown", "--path", projectDir, "--format", "json")
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("error initialising Suite: failed to run infracost in %q dir stderr: %s", projectDir, stderr.String())
	}

	var o Out
	err = json.Unmarshal(out, &o)
	if err != nil {
		log.Fatalf("error initialising: could not marshal infracost output to internal struct %s", err)
	}

	return &Suite{
		dir:       projectDir,
		currency:  o.Currency,
		Resources: o.Projects[0].Breakdown.Resources,
		Summary:   o.Projects[0].Summary,
	}
}

func (s *Suite) Cleanup() {
	if saveCoverage != nil && *saveCoverage {
		s.saveCoverage()
	}

	if showCoverage != nil && *showCoverage {
		s.printCoverage()
	}
}

func (s *Suite) Currency() string {
	if s.parent != nil {
		return s.parent.Currency()
	}

	return s.currency
}

type coverage struct {
	Uncovered []Resource `json:"uncovered"`
	Covered   []Resource `json:"covered"`
}

func (s *Suite) saveCoverage() {
	cov := s.buildCoverage()

	b, err := json.Marshal(cov)
	if err != nil {
		log.Printf("ERROR: could not marshal coverage report %s\n", err)
	}

	err = os.WriteFile("costtest.cov", b, os.ModePerm)
	if err != nil {
		log.Printf("ERROR: could not write coverage report %s\n", err)
	}
}

type filecov struct {
	name           string
	total          float64
	covered        float64
	uncoveredCosts float64
}

type filecovs []filecov

func (f filecovs) Len() int {
	return len(f)
}

func (f filecovs) Less(i, j int) bool {
	return f[i].name < f[j].name
}

func (f filecovs) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (s *Suite) printCoverage() {
	cov := s.buildCoverage()
	filemap := make(map[string]filecov)

	for _, r := range cov.Uncovered {
		name, err := filepath.Rel(s.dir, r.Metadata.Filename)
		if err != nil {
			continue
		}

		if v, ok := filemap[name]; ok {
			filemap[name] = filecov{name: v.name, total: v.total + 1, uncoveredCosts: v.uncoveredCosts + r.MonthlyCost}
		} else {
			filemap[name] = filecov{name: name, total: 1, uncoveredCosts: r.MonthlyCost}
		}
	}

	for _, r := range cov.Covered {
		name, err := filepath.Rel(s.dir, r.Metadata.Filename)
		if err != nil {
			continue
		}

		if v, ok := filemap[name]; ok {
			filemap[name] = filecov{name: v.name, total: v.total + 1, covered: v.covered + 1, uncoveredCosts: v.uncoveredCosts}
		} else {
			filemap[name] = filecov{name: name, total: 1, covered: 1}
		}
	}

	var files filecovs
	for _, f := range filemap {
		files = append(files, f)
	}

	sort.Sort(files)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"File", "Resources", "Uncovered costs"})

	for _, file := range files {
		percent := (file.covered / file.total) * 100
		colour := tablewriter.FgRedColor
		if percent >= 50 {
			colour = tablewriter.FgYellowColor
		}

		if percent >= 85 {
			colour = tablewriter.FgGreenColor
		}

		c := decimal.NewFromFloat(file.uncoveredCosts)
		formatted := output.FormatCost2DP(s.Currency(), &c)
		table.Rich(
			[]string{file.name, fmt.Sprintf("%.2f%% (%.0f/%.0f)", percent, file.covered, file.total), formatted},
			[]tablewriter.Colors{{colour}, {colour}, {colour}},
		)
	}

	if porcelain != nil && *porcelain {
		fmt.Println("--- COSTTEST:")
	}
	table.Render()
	if porcelain != nil && *porcelain {
		fmt.Println("COSTTEST")
	}
}

func (s *Suite) buildCoverage() coverage {
	var cov coverage

	for _, r := range s.Resources {
		if r.covered {
			cov.Covered = append(cov.Covered, *r)
			continue
		}

		cov.Uncovered = append(cov.Uncovered, *r)
	}
	return cov
}

type TestFunc func(t *T, r *Resource)

func (s *Suite) Run(t *testing.T, selector string, f TestFunc) {
	t.Helper()

	for i, r := range s.Resources {
		if r.Name == selector {
			s.Resources[i].setCovered()

			t.Run(selector, func(t2 *testing.T) {
				f(&T{
					T:        t2,
					selector: selector,
				}, r)
			})

			return
		}
	}

	t.Fatalf("unable to find resource with matching selector %q", selector)
}

func (s *Suite) Group(t *testing.T, prefix string) *Suite {
	t.Helper()
	sub := Suite{
		dir:    s.dir,
		parent: s,
	}

	if !strings.HasSuffix(prefix, ".") {
		prefix = prefix + "."
	}

	for _, r := range s.Resources {
		if strings.HasPrefix(r.Name, prefix) {
			sub.Resources = append(sub.Resources, &Resource{
				Metadata:       r.Metadata,
				Name:           strings.TrimPrefix(r.Name, prefix),
				HourlyCost:     r.HourlyCost,
				MonthlyCost:    r.MonthlyCost,
				CostComponents: r.CostComponents,
				SubResources:   r.SubResources,
				parent:         r,
			})
		}
	}

	if len(sub.Resources) == 0 {
		t.Fatalf("failed to find any resources matching prefix %q", prefix)
	}

	return &sub
}

func Init(ops InitOptions) {
	ops.fromInit = true
	defaultSuite = NewSuite(ops)
}

func Cleanup() {
	defaultSuite.Cleanup()
}

func Run(t *testing.T, selector string, f TestFunc) {
	t.Helper()

	if defaultSuite == nil {
		t.Fatalf("called costtest.Run before initalising a Suite call costtest.Init or costtest.NewSuite in testing.Main function")
	}

	defaultSuite.Run(t, selector, f)
}

func Group(t *testing.T, pattern string) *Suite {
	t.Helper()

	if defaultSuite == nil {
		t.Fatalf("called costtest.Run before initalising a Suite call costtest.Init or costtest.NewSuite in testing.Main function")
	}

	return defaultSuite.Group(t, pattern)
}

type T struct {
	selector string

	T *testing.T
}

func (t *T) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.T.Helper()

	return assert.Equal(t.T, expected, actual, msgAndArgs...)
}

func (t *T) InDelta(expected, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	t.T.Helper()

	return assert.InDelta(t.T, expected, actual, delta, msgAndArgs...)
}

func (t *T) GreaterThan(a, b interface{}, msgAndArgs ...interface{}) bool {
	t.T.Helper()

	return assert.Greater(t.T, a, b, msgAndArgs...)
}

func (t *T) GreaterThanOrEqual(a, b interface{}, msgAndArgs ...interface{}) bool {
	t.T.Helper()

	return assert.GreaterOrEqual(t.T, a, b, msgAndArgs...)
}

func (t *T) LessThan(a, b interface{}, msgAndArgs ...interface{}) bool {
	t.T.Helper()

	return assert.Less(t.T, a, b, msgAndArgs...)
}

func (t *T) LessThanOrEqual(a, b interface{}, msgAndArgs ...interface{}) bool {
	t.T.Helper()

	return assert.Less(t.T, a, b, msgAndArgs...)
}
