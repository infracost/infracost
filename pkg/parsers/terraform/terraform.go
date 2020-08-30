package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"infracost/internal/terraform/aws"
	"infracost/pkg/config"
	"infracost/pkg/resource"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func createResource(resourceType string, address string, rawValues map[string]interface{}, providerConfig gjson.Result) resource.Resource {
	awsRegion := "us-east-1" // Use as fallback

	// Find region from terraform provider config
	awsRegionConfig := providerConfig.Get("aws.expressions.region.constant_value").String()
	if awsRegionConfig != "" {
		awsRegion = awsRegionConfig
	}

	// Override the region with the region from the arn if it
	arn := rawValues["arn"]
	if arn != nil {
		awsRegion = strings.Split(arn.(string), ":")[3]
	}

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
		return resource.NewBaseResource(address, rawValues, false) // has no cost
	case "aws_launch_template":
		return resource.NewBaseResource(address, rawValues, false) // has no cost
	case "aws_autoscaling_group":
		return aws.NewEc2AutoscalingGroup(address, awsRegion, rawValues)
	case "aws_db_instance":
		return aws.NewRdsInstance(address, awsRegion, rawValues)
	case "aws_elb":
		return aws.NewElb(address, awsRegion, rawValues, true) // is classic
	case "aws_lb":
		return aws.NewElb(address, awsRegion, rawValues, false)
	case "aws_alb": // alias for aws_lb
		return aws.NewElb(address, awsRegion, rawValues, false)
	case "aws_nat_gateway":
		return aws.NewNatGateway(address, awsRegion, rawValues)
	case "aws_lambda_function":
		return aws.NewLambdaFunction(address, awsRegion, rawValues)
	case "aws_dynamodb_table":
		return aws.NewDynamoDBTable(address, awsRegion, rawValues)
	case "aws_ecs_cluster":
		return resource.NewBaseResource(address, rawValues, false) // has no cost
	case "aws_ecs_task_definition":
		return resource.NewBaseResource(address, rawValues, false) // has no cost
	case "aws_ecs_service":
		return aws.NewEcsService(address, awsRegion, rawValues)
	}
	return nil
}

type TerraformOptions struct {
	TerraformDir string
}

func terraformCommand(options *TerraformOptions, args ...string) ([]byte, error) {
	terraformBinary := os.Getenv("TERRAFORM_BINARY")
	if terraformBinary == "" {
		terraformBinary = "terraform"
	}

	cmd := exec.Command(terraformBinary, args...)
	if config.Config.NoColor {
		log.Infof("Running command: %s", cmd.String())
	} else {
		log.Info(color.HiBlackString("Running command: %s", cmd.String()))
	}
	cmd.Dir = options.TerraformDir

	var outbuf bytes.Buffer
	cmd.Stdout = bufio.NewWriter(&outbuf)
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

func ParsePlanJSON(j []byte) ([]resource.Resource, error) {
	planJSON := gjson.ParseBytes(j)
	providerConfig := planJSON.Get("configuration.provider_config")
	plannedValuesJSON := planJSON.Get("planned_values.root_module")
	configurationJSON := planJSON.Get("configuration.root_module")

	resources, err := parseModule(planJSON, providerConfig, plannedValuesJSON, configurationJSON)
	return resources, err
}

func parseModule(planJSON gjson.Result, providerConfig gjson.Result, plannedValuesJSON gjson.Result, configurationJSON gjson.Result) ([]resource.Resource, error) {
	resourceMap := make(map[string]resource.Resource)
	moduleAddr := plannedValuesJSON.Get("address").String()
	terraformResources := plannedValuesJSON.Get("resources").Array()

	for _, terraformResource := range terraformResources {
		address := terraformResource.Get("address").String()
		resourceType := terraformResource.Get("type").String()
		var rawValues map[string]interface{}
		if terraformResource.Get("values").Value() != nil {
			rawValues = terraformResource.Get("values").Value().(map[string]interface{})
		} else {
			rawValues = make(map[string]interface{})
		}
		resource := createResource(resourceType, address, rawValues, providerConfig)
		if resource != nil {
			resourceMap[getInternalName(resource.Address(), moduleAddr)] = resource
		}
	}

	for _, resource := range resourceMap {
		resourceJSON := configurationJSON.Get(fmt.Sprintf(`resources.#(address="%s")`, stripResourceArrayPart(getInternalName(resource.Address(), moduleAddr))))
		addReferences(resource, resourceJSON, resourceMap)
	}

	resources := make([]resource.Resource, 0, len(resourceMap))
	for _, resource := range resourceMap {
		resources = append(resources, resource)
	}

	for _, pvJSON := range plannedValuesJSON.Get("child_modules").Array() {
		moduleName := parseModuleName(pvJSON.Get("address").String())
		cJSON := configurationJSON.Get(fmt.Sprintf("module_calls.%s.module", moduleName))
		moduleResources, err := parseModule(planJSON, providerConfig, pvJSON, cJSON)
		if err != nil {
			return resources, err
		}
		resources = append(resources, moduleResources...)
	}

	return resources, nil
}

func getInternalName(resourceAddr string, moduleAddr string) string {
	return strings.TrimPrefix(resourceAddr, moduleAddr+".")
}

func stripResourceArrayPart(addr string) string {
	r := regexp.MustCompile("([^[]+)")
	match := r.FindStringSubmatch(addr)
	if len(match) <= 1 {
		return ""
	}
	return match[1]
}

func parseModuleName(moduleAddr string) string {
	if moduleAddr == "" {
		return "root_module"
	}
	r := regexp.MustCompile("module.([^[]+)")
	match := r.FindStringSubmatch(moduleAddr)
	if len(match) <= 1 {
		return ""
	}
	return match[1]
}

func addReferencesHelper(r resource.Resource, key string, valueJSON gjson.Result, resourceMap map[string]resource.Resource) {
	var refAddr string
	if valueJSON.Get("references").Exists() {
		refAddr = valueJSON.Get("references").Array()[0].String()
	} else if len(valueJSON.Array()) > 0 && valueJSON.Get("0.id.references").Exists() {
		refAddr = valueJSON.Get("0.id.references").Array()[0].String()
	}

	if resource, ok := resourceMap[refAddr]; ok {
		r.AddReference(key, resource)
	} else if valueJSON.Type.String() == "JSON" {
		valueJSON.ForEach(func(k gjson.Result, v gjson.Result) bool {
			addReferencesHelper(r, k.String(), v, resourceMap)
			return true
		})
	}
}

func addReferences(r resource.Resource, resourceJSON gjson.Result, resourceMap map[string]resource.Resource) {
	addReferencesHelper(r, "expressions", resourceJSON.Get("expressions"), resourceMap)
}
