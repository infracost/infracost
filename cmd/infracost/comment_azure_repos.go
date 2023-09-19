package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/comment"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

var validCommentAzureReposBehaviors = []string{"update", "new", "delete-and-new"}

func commentAzureReposCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure-repos",
		Short: "Post an Infracost comment to Azure Repos",
		Long:  "Post an Infracost comment to Azure Repos",
		Example: `  Update comment on a pull request:

      infracost comment azure-repos --repo-url https://dev.azure.com/my-org/my-project/_git/my-repo --pull-request 3 --path infracost.json --azure-access-token $AZURE_ACCESS_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx.ContextValues.SetValue("platform", "azure-repos")

			var err error

			token, _ := cmd.Flags().GetString("azure-access-token")
			tag, _ := cmd.Flags().GetString("tag")
			extra := comment.AzureReposExtra{
				Token: token,
				Tag:   tag,
			}

			prNumber, _ := cmd.Flags().GetInt("pull-request")
			repoURL, _ := cmd.Flags().GetString("repo-url")

			var commentHandler *comment.CommentHandler
			if prNumber != 0 {
				ctx.ContextValues.SetValue("targetType", "pull-request")

				commentHandler, err = comment.NewAzureReposPRHandler(ctx.Context(), repoURL, strconv.Itoa(prNumber), extra)
				if err != nil {
					return err
				}
			} else {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--pull-request is required")
			}

			behavior, _ := cmd.Flags().GetString("behavior")
			if behavior != "" && !contains(validCommentAzureReposBehaviors, behavior) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--behavior only supports %s", strings.Join(validCommentAzureReposBehaviors, ", "))
			}
			ctx.ContextValues.SetValue("behavior", behavior)

			paths, _ := cmd.Flags().GetStringArray("path")

			commentOut, commentErr := buildCommentOutput(cmd, ctx, paths, output.MarkdownOptions{
				WillUpdate:          prNumber != 0 && behavior == "update",
				WillReplace:         prNumber != 0 && behavior == "delete-and-new",
				IncludeFeedbackLink: !ctx.Config.IsSelfHosted(),
			})
			if isErrorUnhandled(commentErr) {
				return commentErr
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if !dryRun {
				skipNoDiff, _ := cmd.Flags().GetBool("skip-no-diff")

				res, err := commentHandler.CommentWithBehavior(ctx.Context(), behavior, commentOut.Body, &comment.CommentOpts{
					ValidAt:    commentOut.ValidAt,
					SkipNoDiff: !commentOut.HasDiff && skipNoDiff,
				})
				if err != nil {
					return err
				}

				pricingClient := apiclient.GetPricingAPIClient(ctx)
				err = pricingClient.AddEvent("infracost-comment", ctx.EventEnv())
				if err != nil {
					logging.Logger.WithError(err).Error("could not report infracost-comment event")
				}

				if res.Posted {
					cmd.Println("Comment posted to Azure Repos")
				} else {
					msg := "Comment not posted to Azure Repos"
					if res.SkipReason != "" {
						msg += fmt.Sprintf(": %s", res.SkipReason)
					}
					cmd.Println(msg)
				}
			} else {
				cmd.Println(commentOut.Body)
				cmd.Println("Comment not posted to Azure Repos (--dry-run was specified)")
			}

			return commentErr
		},
	}

	cmd.Flags().String("behavior", "update", `Behavior when posting comment, one of:
  update (default)  Update latest comment
  new               Create a new comment
  delete-and-new    Delete previous matching comments and create a new comment`)
	_ = cmd.RegisterFlagCompletionFunc("behavior", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validCommentAzureReposBehaviors, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().String("azure-access-token", "", "Azure DevOps access token")
	_ = cmd.MarkFlagRequired("azure-access-token")
	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	var prNumber PRNumber
	cmd.Flags().Var(&prNumber, "pull-request", "Pull request number to post comment on")
	_ = cmd.MarkFlagRequired("pull-request")
	cmd.Flags().String("repo-url", "", "Repository URL, e.g. https://dev.azure.com/my-org/my-project/_git/my-repo")
	_ = cmd.MarkFlagRequired("repo-url")
	cmd.Flags().String("tag", "", "Customize hidden markdown tag used to detect comments posted by Infracost")
	cmd.Flags().Bool("dry-run", false, "Generate comment without actually posting to Azure Repos")

	return cmd
}
