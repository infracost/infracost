package main

import (
	"bytes"
	"context"
	"flag"
	"go/format"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/api/compute/v1"

	"github.com/infracost/infracost/internal/logging"
)

func main() {
	flag.Parse()
	provider := flag.Arg(0)
	switch provider {
	case "gcp":
		describeGCPZones()
	default:
		describeAWSZones()
	}
}

func describeAWSZones() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logging.Logger.Fatal().Msgf("error loading aws config %s", err)
	}

	svc := ec2.NewFromConfig(cfg)

	resp, err := svc.DescribeRegions(context.Background(), &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		logging.Logger.Fatal().Msgf("error describing ec2 regions %s", err)
	}
	m := make(map[string]map[string]cty.Value)
	for _, region := range resp.Regions {
		name := *region.RegionName
		regionalConf, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(name),
		)
		if err != nil {
			logging.Logger.Fatal().Msg(err.Error())
		}

		regionalSvc := ec2.NewFromConfig(regionalConf)

		result, err := regionalSvc.DescribeAvailabilityZones(context.Background(), &ec2.DescribeAvailabilityZonesInput{
			AllAvailabilityZones: aws.Bool(true),
		})
		if err != nil {
			log.Println(name, err)
			continue
		}

		var names, zoneIds, groupNames []cty.Value

		for _, zone := range result.AvailabilityZones {
			names = append(names, cty.StringVal(*zone.ZoneName))
			zoneIds = append(zoneIds, cty.StringVal(*zone.ZoneId))
			groupNames = append(groupNames, cty.StringVal(*zone.GroupName))
		}

		m[*region.RegionName] = map[string]cty.Value{
			"id":          cty.StringVal(*region.RegionName),
			"names":       cty.ListVal(names),
			"zone_ids":    cty.ListVal(zoneIds),
			"group_names": cty.ListVal(groupNames),
		}
	}

	tmpl, err := template.New("test").Parse(`
package hcl

import "github.com/zclconf/go-cty/cty"

var awsZones = map[string]cty.Value{
	{{- range $regionName, $region := . }}
	"{{ $regionName }}": cty.ObjectVal(map[string]cty.Value{
		"id": {{ $region.id.GoString }},
		"names": {{ $region.names.GoString }},
		"zone_ids": {{ $region.zone_ids.GoString }},
		"group_names": {{ $region.group_names.GoString }},
	}),
	{{- end }}
}
	`)
	if err != nil {
		logging.Logger.Fatal().Msgf("failed to create template: %s", err)
	}

	buf := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buf, m)
	if err != nil {
		logging.Logger.Fatal().Msgf("error executing template: %s", err)
	}

	writeOutput("aws", buf.Bytes())
}
func describeGCPZones() {
	ctx := context.Background()

	computeService, err := compute.NewService(ctx)
	if err != nil {
		logging.Logger.Fatal().Msgf("failed to create compute service: %s", err)
	}

	projectID := "691877312977"

	req := computeService.Regions.List(projectID)

	regions := make(map[string]cty.Value)
	if err := req.Pages(ctx, func(page *compute.RegionList) error {
		for _, region := range page.Items {
			var zones []cty.Value
			for _, zoneURL := range region.Zones {
				zones = append(zones, cty.StringVal(path.Base(zoneURL)))
			}

			regions[region.Name] = cty.ListVal(zones)
		}

		return nil
	}); err != nil {
		logging.Logger.Fatal().Msgf("Failed to list regions: %s", err)
	}

	tmpl, err := template.New("test").Parse(`
package hcl

import "github.com/zclconf/go-cty/cty"

var gcpZones = map[string]cty.Value{
	{{- range $regionName, $zones := . }}
	"{{ $regionName }}": cty.ObjectVal(map[string]cty.Value{
		"names": {{ $zones.GoString }},
	}),
	{{- end }}
}
	`)
	if err != nil {
		logging.Logger.Fatal().Msgf("failed to create template: %s", err)
	}

	buf := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buf, regions)
	if err != nil {
		logging.Logger.Fatal().Msgf("error executing template: %s", err)
	}

	writeOutput("gcp", buf.Bytes())
}

func writeOutput(provider string, input []byte) {
	f, err := os.Create("zones_" + provider + ".go")
	if err != nil {
		logging.Logger.Fatal().Msgf("could not create output file: %s", err)
	}
	defer f.Close()

	formatted, err := format.Source(input)
	if err != nil {
		logging.Logger.Fatal().Msgf("could not format output: %s", err)
	}

	_, err = f.Write(formatted)
	if err != nil {
		logging.Logger.Fatal().Msgf("could not write output: %s", err)
	}
}
