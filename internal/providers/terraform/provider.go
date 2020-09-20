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
		_, err := terraformCmd(opts, "init")
		if err != nil {
			return []byte{}, errors.Wrap(err, "error initializing terraform working directory")
		}

		f, err := ioutil.TempFile(os.TempDir(), "tfplan")
		if err != nil {
			return []byte{}, errors.Wrap(err, "error creating temporary file 'tfplan'")
		}
		defer os.Remove(f.Name())

		_, err = terraformCmd(opts, "plan", "-input=false", "-lock=false", fmt.Sprintf("-out=%s", f.Name()))
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

func ShowSkippedResources(resources []*schema.Resource, showDetails bool) {
	skippedCount := 0
	unSupportedCount := 0
	freeCount := 0
	unSupportedTypeCount := make(map[string]int)
	for _, r := range resources {
		if r.IsSkipped {
			skippedCount++
			if r.IsFree() {
				freeCount++
			} else {
				unSupportedCount++
				if _, ok := unSupportedTypeCount[r.ResourceType]; !ok {
					unSupportedTypeCount[r.ResourceType] = 0
				}
				unSupportedTypeCount[r.ResourceType]++
			}
		}
	}
	if unSupportedCount == 0 {
		return
	}
	message := "%d out of %d resources couldn't be estimated as Infracost doesn't support them yet (https://www.infracost.io/docs/supported_resources)"
	if showDetails {
		message += ".\n"
	} else {
		message += ", re-run with --show-skipped to see the list.\n"
	}
	fmt.Printf(message, unSupportedCount, len(resources))
	if showDetails {
		for rType, count := range unSupportedTypeCount {
			fmt.Printf("%d x %s \n", count, rType)
		}
	}
	fmt.Println("We're continually adding new resources, please create an issue if you'd like us to prioritize your list.")
}
