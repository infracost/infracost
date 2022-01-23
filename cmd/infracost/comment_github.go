package main

import (
	"fmt"
	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/comment"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

var validCommentGitHubBehaviors = []string{"update", "new", "hide-and-new", "delete-and-new"}

func commentGitHubCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Post an Infracost comment to GitHub",
		Long:  "Post an Infracost comment to GitHub",
		Example: `  Update a comment on a pull request:

      infracost comment github --repo my-org/my-github-repo --pull-request 3 --path infracost.json --github-token $GITHUB_TOKEN

  Post a new comment to a commit:

      infracost comment github --repo my-org/my-github-repo --commit 2ca7182 --path infracost.json --behavior hide-and-new --github-token $GITHUB_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx.SetContextValue("platform", "github")

			var err error

			apiURL, _ := cmd.Flags().GetString("github-api-url")
			token, _ := cmd.Flags().GetString("github-token")
			tag, _ := cmd.Flags().GetString("tag")
			extra := comment.GitHubExtra{
				APIURL: apiURL,
				Token:  token,
				Tag:    tag,
			}

			commit, _ := cmd.Flags().GetString("commit")
			prNumber, _ := cmd.Flags().GetInt("pull-request")
			repo, _ := cmd.Flags().GetString("repo")

			var commentHandler *comment.CommentHandler
			if prNumber != 0 {
				ctx.SetContextValue("targetType", "pull-request")

				commentHandler, err = comment.NewGitHubPRHandler(ctx.Context(), repo, strconv.Itoa(prNumber), extra)
				if err != nil {
					return err
				}
			} else if commit != "" {
				ctx.SetContextValue("targetType", "commit")

				commentHandler, err = comment.NewGitHubCommitHandler(ctx.Context(), repo, commit, extra)
				if err != nil {
					return err
				}
			} else {
				ui.PrintUsage(cmd)
				return fmt.Errorf("either --commit or --pull-request is required")
			}

			behavior, _ := cmd.Flags().GetString("behavior")
			if behavior != "" && !contains(validCommentGitHubBehaviors, behavior) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--behavior only supports %s", strings.Join(validCommentGitHubBehaviors, ", "))
			}
			ctx.SetContextValue("behavior", behavior)

			paths, _ := cmd.Flags().GetStringArray("path")

			body, err := buildCommentBody(ctx, paths, output.MarkdownOptions{
				WillUpdate:          prNumber != 0 && behavior == "update",
				WillReplace:         prNumber != 0 && behavior == "delete-and-new",
				IncludeFeedbackLink: true,
			})
			if err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if !dryRun {
				err = commentHandler.CommentWithBehavior(ctx.Context(), behavior, string(body))
				if err != nil {
					return err
				}

				pricingClient := apiclient.NewPricingAPIClient(ctx)
				err = pricingClient.AddEvent("infracost-comment", ctx.EventEnv())
				if err != nil {
					log.Errorf("Error reporting event: %s", err)
				}

				cmd.Println("Comment posted to GitHub")
			} else {
				cmd.Println(string(body))
				cmd.Println("Comment not posted to GitHub (--dry-run was specified)")
			}

			return nil
		},
	}

	cmd.Flags().String("behavior", "update", `Behavior when posting the comment, one of:
  update (default)  Update the latest comment
  new               Create a new comment
  hide-and-new      Hide previous matching comments and create a new comment
  delete-and-new    Delete previous matching comments and create a new comment`)
	_ = cmd.RegisterFlagCompletionFunc("behavior", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validCommentGitHubBehaviors, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().String("commit", "", "Commit SHA to post/get the comment, mutually exclusive with pull-request")
	cmd.Flags().String("github-api-url", "https://api.github.com", "GitHub API URL, defaults to https://api.github.com")
	cmd.Flags().String("github-token", "", "GitHub token")
	_ = cmd.MarkFlagRequired("github-token")
	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	cmd.Flags().Int("pull-request", 0, "Pull request number to post the comment on, mutually exclusive with commit")
	cmd.Flags().String("repo", "", "Repository in the format owner/repo")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().String("tag", "", "Customize the embedded tag that is used for detecting comments posted by Infracost")
	cmd.Flags().Bool("dry-run", false, "Generate the comment without actually posting to GitHub.")

	return cmd
}
