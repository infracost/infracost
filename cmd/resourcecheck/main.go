package main

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/slack-go/slack"

	"github.com/infracost/infracost/internal/logging"
)

var (
	api = slack.New(os.Getenv("SLACK_API_TOKEN"))
)

func main() {
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

	f, err := decorator.ParseFile(token.NewFileSet(), "internal/resources/aws/util.go", nil, parser.ParseComments)
	if err != nil {
		logging.Logger.Fatal().Msgf("error loading aws util file %s", err)
	}

	currentRegions := make(map[string]struct{})
	for _, decl := range f.Decls {
		if v, ok := decl.(*dst.GenDecl); ok {
			for _, spec := range v.Specs {
				if vs, ok := spec.(*dst.ValueSpec); ok && vs.Names[0].Name == "RegionMapping" {

					cl := vs.Values[0].(*dst.CompositeLit)
					for _, e := range cl.Elts {
						val := strings.ReplaceAll(e.(*dst.KeyValueExpr).Key.(*dst.BasicLit).Value, `"`, "")
						currentRegions[val] = struct{}{}
					}
				}
			}
		}
	}

	if len(currentRegions) == 0 {
		logging.Logger.Fatal().Msg("error parsing aws RegionMapping from util.go, empty list found")
	}

	notFound := strings.Builder{}
	for _, r := range resp.Regions {
		if _, ok := currentRegions[*r.RegionName]; !ok {
			notFound.WriteString(*r.RegionName + ",")
		}
	}

	if notFound.Len() > 0 {
		sendSlackMessage(strings.TrimRight(notFound.String(), ","))
		return
	}
}

func sendSlackMessage(regions string) {
	attachment := slack.Attachment{
		Pretext: "found missing region configuration:",
		Text:    regions,
		Color:   "ff0000",
	}

	_, _, err := api.PostMessage(
		"production",
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		logging.Logger.Fatal().Msgf("error sending slack notifications %s", err)
	}
}
