package terraform

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"plancosts/internal/terraform/aws"
	"plancosts/pkg/base"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func createResource(resourceType string, address string, rawValues map[string]interface{}, providerConfig gjson.Result) base.Resource {
	awsRegion := providerConfig.Get("aws.expressions.region.constant_value").String()

	switch resourceType {
	case "aws_instance":
		return aws.NewEc2Instance(address, awsRegion, rawValues)
	case "aws_ebs_volume":
		return aws.NewEbsVolume(address, awsRegion, rawValues)
	case "aws_ebs_snapshot":
		return aws.NewEbsSnapshot(address, awsRegion, rawValues)
	case "aws_ebs_snapshot_copy":
		return aws.NewEbsSnapshotCopy(address, awsRegion, rawValues)
	case "aws_launch_configuration":
		return aws.NewEc2LaunchConfiguration(address, awsRegion, rawValues)
	case "aws_launch_template":
		return aws.NewEc2LaunchTemplate(address, awsRegion, rawValues)
	case "aws_autoscaling_group":
		return aws.NewEc2AutoscalingGroup(address, awsRegion, rawValues)
	case "aws_elb":
		return aws.NewElb(address, awsRegion, rawValues, true) // is classic
	case "aws_lb":
		return aws.NewElb(address, awsRegion, rawValues, false)
	case "aws_alb": // alias for aws_lb
		return aws.NewElb(address, awsRegion, rawValues, false)
	}
	return nil
}

type TerraformOptions struct {
	TerraformDir string
}

func terraformCommand(options *TerraformOptions, args ...string) ([]byte, error) {
	cmd := exec.Command("terraform", args...)
	log.Info(color.HiGreenString("Running command: %s", cmd.String()))
	cmd.Dir = options.TerraformDir

	var outbuf bytes.Buffer
	mw := io.MultiWriter(log.StandardLogger().WriterLevel(log.DebugLevel), &outbuf)
	cmd.Stdout = mw
	cmd.Stderr = log.StandardLogger().WriterLevel(log.ErrorLevel)
	err := cmd.Run()
	return outbuf.Bytes(), err
}

func LoadPlanJSON(path string) ([]byte, error) {
	planFile, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer planFile.Close()
	out, err := ioutil.ReadAll(planFile)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

func GeneratePlanJSON(tfdir string, planPath string) ([]byte, error) {
	var err error

	opts := &TerraformOptions{
		TerraformDir: tfdir,
	}

	if planPath == "" {
		_, err = terraformCommand(opts, "init")
		if err != nil {
			return []byte{}, err
		}

		planfile, err := ioutil.TempFile(os.TempDir(), "tfplan")
		if err != nil {
			return []byte{}, err
		}
		defer os.Remove(planfile.Name())

		_, err = terraformCommand(opts, "plan", "-input=false", "-lock=false", fmt.Sprintf("-out=%s", planfile.Name()))
		if err != nil {
			return []byte{}, err
		}

		planPath = planfile.Name()
	}

	out, err := terraformCommand(opts, "show", "-json", planPath)
	if err != nil {
		return []byte{}, err
	}

	return out, nil
}

func ParsePlanJSON(planJSON []byte) ([]base.Resource, error) {
	resourceMap := make(map[string]base.Resource)

	providerConfig := gjson.GetBytes(planJSON, "configuration.provider_config")
	terraformResources := gjson.GetBytes(planJSON, "planned_values.root_module.resources")

	for _, terraformResource := range terraformResources.Array() {
		address := terraformResource.Get("address").String()
		resourceType := terraformResource.Get("type").String()
		rawValues := terraformResource.Get("values").Value().(map[string]interface{})
		resource := createResource(resourceType, address, rawValues, providerConfig)
		if resource != nil {
			resourceMap[address] = resource
		}
	}

	for _, resource := range resourceMap {
		query := fmt.Sprintf(`configuration.root_module.resources.#(address="%s")`, resource.Address())
		terraformResourceConfig := gjson.GetBytes(planJSON, query)
		addReferences(resource, terraformResourceConfig, resourceMap)
	}

	resources := make([]base.Resource, 0, len(resourceMap))
	for _, resource := range resourceMap {
		resources = append(resources, resource)
	}

	return resources, nil
}

func addReferences(r base.Resource, resourceConfig gjson.Result, resourceMap map[string]base.Resource) {
	gjson.Get(resourceConfig.String(), "expressions").ForEach(func(key gjson.Result, value gjson.Result) bool {
		var refAddr string
		if value.Get("references").Exists() {
			refAddr = value.Get("references").Array()[0].String()
		} else if len(value.Array()) > 0 {
			idVal := value.Array()[0].Get("id")
			if idVal.Get("references").Exists() {
				refAddr = idVal.Get("references").Array()[0].String()
			}
		}
		if resource, ok := resourceMap[refAddr]; ok {
			r.AddReference(key.String(), resource)
		}
		return true
	})
}
