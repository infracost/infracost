package terraform

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/infracost/infracost/pkg/schema"
	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"
)

type terraformProvider struct {
	jsonFile string
	planFile string
	dir      string
}

// New returns new Terraform Provider
func New() schema.Provider {
	return &terraformProvider{}
}

func (p *terraformProvider) ProcessArgs(c *cli.Context) error {
	p.jsonFile = c.String("tfjson")
	p.planFile = c.String("tfplan")
	p.dir = c.String("tfdir")

	if p.jsonFile != "" && p.planFile != "" {
		return errors.New("Please provide either a Terraform Plan JSON file (tfjson) or a Terraform Plan file (tfplan)")
	}

	if p.planFile != "" && p.dir == "" {
		return errors.New("Please provide a path to the Terraform project (tfdir) if providing a Terraform Plan file (tfplan)")
	}

	return nil
}

func (p *terraformProvider) LoadResources() ([]*schema.Resource, error) {
	var plan []byte
	var err error

	if p.jsonFile != "" {
		plan, err = loadPlanJSON(p.jsonFile)
	} else {
		plan, err = generatePlanJSON(p.dir, p.planFile)
	}

	if err != nil {
		return []*schema.Resource{}, errors.Wrap(err, "error loading resources")
	}

	return parsePlanJSON(plan), nil
}

func loadPlanJSON(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "error opening file '%v'", path)
	}
	defer f.Close()

	out, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "error reading file '%v'", path)
	}

	return out, nil
}

func generatePlanJSON(dir string, path string) ([]byte, error) {
	opts := &cmdOptions{
		TerraformDir: dir,
	}

	if path == "" {
		_, err := terraformCmd(opts, "init", "-no-color")
		if err != nil {
			return []byte{}, errors.Wrap(err, "error initializing terraform working directory")
		}

		f, err := ioutil.TempFile(os.TempDir(), "tfplan")
		if err != nil {
			return []byte{}, errors.Wrap(err, "error creating temporary file 'tfplan'")
		}
		defer os.Remove(f.Name())

		_, err = terraformCmd(opts, "plan", "-input=false", "-lock=false", "-no-color", fmt.Sprintf("-out=%s", f.Name()))
		if err != nil {
			return []byte{}, errors.Wrap(err, "error generating terraform execution plan")
		}

		path = f.Name()
	}

	out, err := terraformCmd(opts, "show", "-json", path)
	if err != nil {
		return []byte{}, errors.Wrap(err, "error inspecting terraform plan")
	}

	return out, nil
}

func CountSkippedResources(resources []*schema.Resource) (map[string]int, int, int) {
	skippedCount := 0
	unsupportedCount := 0
	unsupportedTypeCount := make(map[string]int)
	for _, r := range resources {
		if r.IsSkipped {
			skippedCount++
			unsupportedCount++
			if _, ok := unsupportedTypeCount[r.ResourceType]; !ok {
				unsupportedTypeCount[r.ResourceType] = 0
			}
			unsupportedTypeCount[r.ResourceType]++
		}
	}
	return unsupportedTypeCount, skippedCount, unsupportedCount
}

func SkippedResourcesMessage(resources []*schema.Resource, showDetails bool) string {
	unsupportedTypeCount, _, unsupportedCount := CountSkippedResources(resources)
	if unsupportedCount == 0 {
		return ""
	}
	message := fmt.Sprintf("%d out of %d resources couldn't be estimated as Infracost doesn't support them yet (https://www.infracost.io/docs/supported_resources)", unsupportedCount, len(resources))
	if showDetails {
		message += ".\n"
	} else {
		message += ", re-run with --show-skipped to see the list.\n"
	}
	message += "We're continually adding new resources, please create an issue if you'd like us to prioritize your list.\n"
	if showDetails {
		for rType, count := range unsupportedTypeCount {
			message += fmt.Sprintf("%d x %s\n", count, rType)
		}
	}
	return message
}
