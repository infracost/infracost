package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/spf13/cobra"
)

func commentCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Post an Infracost comment to GitHub, GitLab or Azure DevOps",
		Long:  "Post an Infracost comment to GitHub, GitLab or Azure DevOps",
		Example: `  Update a comment on a GitHub pull request:

      infracost comment github --repo my-org/my-github-repo --pull-request 3 --path infracost.json --github-token $GITHUB_TOKEN

  Post a new comment to a GitLab commit:

      infracost comment gitlab --repo my-org/my-gitlab-repo --commit 2ca7182 --path infracost.json --behavior new --gitlab-token $GITLAB_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(commentGitHubCmd(ctx))

	return cmd
}

func buildCommentBody(ctx *config.RunContext, paths []string, mdOpts output.MarkdownOptions) ([]byte, error) {
	inputs, err := output.LoadPaths(paths)
	if err != nil {
		return nil, err
	}

	combined, err := output.Combine(inputs)
	if err != nil {
		return nil, err
	}

	opts := output.Options{
		DashboardEnabled: ctx.Config.EnableDashboard,
		NoColor:          ctx.Config.NoColor,
		IncludeHTML:      true,
		ShowSkipped:      true,
	}

	b, err := output.ToMarkdown(combined, opts, mdOpts)
	if err != nil {
		return nil, err
	}

	return b, nil
}
