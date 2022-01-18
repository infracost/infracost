package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

func commentCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Post an Infracost comment to GitHub, GitLab, Azure Repos or Bitbucket",
		Long:  "Post an Infracost comment to GitHub, GitLab, Azure Repos or Bitbucket",
		Example: `  Update the Infracost comment on a GitHub pull request:

      infracost comment github --repo my-org/my-repo --pull-request 3 --path infracost.json --behavior update --github-token $GITHUB_TOKEN

  Delete old Infracost comments and post a new comment to a GitLab commit:

      infracost comment gitlab --repo my-org/my-repo --commit 2ca7182 --path infracost.json --behavior delete-and-new --gitlab-token $GITLAB_TOKEN

  Post a new comment to an Azure Repos pull request:

      infracost comment azure-repos --repo-url https://dev.azure.com/my-org/my-project/_git/my-repo --pull-request 3 --path infracost.json --behavior new --azure-access-token $AZURE_ACCESS_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmds := []*cobra.Command{commentGitHubCmd(ctx), commentGitLabCmd(ctx), commentAzureReposCmd(ctx), commentBitbucketCmd(ctx)}
	for _, subCmd := range cmds {
		subCmd.Flags().StringArray("policy-path", nil, "Paths to any Infracost cost policies (experimental)")
	}

	cmd.AddCommand(cmds...)

	return cmd
}

func buildCommentBody(cmd *cobra.Command, ctx *config.RunContext, paths []string, mdOpts output.MarkdownOptions) ([]byte, error) {
	inputs, err := output.LoadPaths(paths)
	if err != nil {
		return nil, err
	}

	combined, err := output.Combine(inputs)
	if err != nil {
		return nil, err
	}
	combined.IsCIRun = ctx.IsCIRun()

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if ctx.Config.EnableDashboard && !dryRun {
		if ctx.Config.IsSelfHosted() {
			ui.PrintWarning(cmd.ErrOrStderr(), "The dashboard is part of Infracost's hosted services. Contact hello@infracost.io for help.")
		}

		combined.RunID, combined.ShareURL = shareCombinedRun(ctx, combined, inputs)
	}

	var policyChecks output.PolicyCheck
	policyPaths, _ := cmd.Flags().GetStringArray("policy-path")
	if len(policyPaths) > 0 {
		policyChecks, err = queryPolicy(policyPaths, combined)
		if err != nil {
			return nil, err
		}
	}

	opts := output.Options{
		DashboardEnabled: ctx.Config.EnableDashboard,
		NoColor:          ctx.Config.NoColor,
		ShowSkipped:      true,
		PolicyChecks:     policyChecks,
	}

	b, err := output.ToMarkdown(combined, opts, mdOpts)
	if err != nil {
		return nil, err
	}

	if policyChecks.HasFailed() {
		return b, policyChecks.Failures
	}

	return b, nil
}

type PRNumber int

func (p *PRNumber) Set(value string) error {
	if value == "" {
		return nil
	}

	v, err := strconv.Atoi(value)
	*p = PRNumber(v)

	if err != nil {
		return errors.New("must be integer")
	}

	return nil
}

func (p *PRNumber) String() string {
	return fmt.Sprintf("%d", *p)
}

func (p *PRNumber) Type() string {
	return "int"
}

func queryPolicy(policyPaths []string, input output.Root) (output.PolicyCheck, error) {
	checks := output.PolicyCheck{
		Enabled: true,
	}

	inputValue, err := ast.InterfaceToValue(input)
	if err != nil {
		return checks, fmt.Errorf("Unable to process infracost output into rego input: %s", err.Error())
	}

	ctx := context.Background()
	r := rego.New(
		rego.Query("data.infracost.deny"),
		rego.ParsedInput(inputValue),
		rego.Load(policyPaths, func(abspath string, info os.FileInfo, depth int) bool {
			return false
		}),
	)
	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		return checks, fmt.Errorf("Unable to query cost policy: %s", err.Error())
	}

	res, err := pq.Eval(ctx)
	if err != nil {
		return checks, err
	}

	for _, e := range res[0].Expressions {
		switch v := e.Value.(type) {
		case map[string]interface{}:
			readPolicyOut(v, &checks)
		case []interface{}:
			for _, ii := range v {
				if m, ok := ii.(map[string]interface{}); ok {
					readPolicyOut(m, &checks)
				}
			}
		}
	}

	return checks, nil
}

func readPolicyOut(v map[string]interface{}, checks *output.PolicyCheck) {
	if _, ok := v["failed"]; !ok {
		log.Debugf("skipping policy output as did not contain [failed] property in output object")
		return
	}

	if _, ok := v["msg"]; !ok {
		log.Debugf("skipping policy output as did not contain [msg] property in output object")
		return
	}

	failed := v["failed"].(bool)
	msg := v["msg"].(string)

	if failed {
		checks.Failures = append(checks.Failures, msg)
		return
	}

	checks.Passed = append(checks.Passed, msg)
}
