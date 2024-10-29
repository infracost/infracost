package main

import (
	"fmt"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/comment"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

var validCommentBitbucketBehaviors = []string{"update", "new", "delete-and-new"}

func commentBitbucketCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bitbucket",
		Short: "Post an Infracost comment to Bitbucket",
		Long:  "Post an Infracost comment to Bitbucket",
		Example: `  Update comment on a pull request:

      infracost comment bitbucket --repo my-org/my-repo --pull-request 3 --path infracost.json --bitbucket-token $BITBUCKET_TOKEN

  Post a new comment to a commit:

      infracost comment bitbucket --repo my-org/my-repo --commit 2ca7182 --path infracost.json --behavior delete-and-new --bitbucket-token $BITBUCKET_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx.ContextValues.SetValue("platform", "bitbucket")

			var err error

			format, _ := cmd.Flags().GetString("format")
			format = strings.ToLower(format)
			if format != "" && !contains(validCommentOutputFormats, format) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--format only supports %s", strings.Join(validCommentOutputFormats, ", "))
			}

			serverURL, _ := cmd.Flags().GetString("bitbucket-server-url")
			token, _ := cmd.Flags().GetString("bitbucket-token")
			tag, _ := cmd.Flags().GetString("tag")
			omitDetails, _ := cmd.Flags().GetBool("exclude-cli-output")
			extra := comment.BitbucketExtra{
				ServerURL:   serverURL,
				Token:       token,
				Tag:         tag,
				OmitDetails: omitDetails,
			}

			commit, _ := cmd.Flags().GetString("commit")
			prNumber, _ := cmd.Flags().GetInt("pull-request")
			repo, _ := cmd.Flags().GetString("repo")

			var commentHandler *comment.CommentHandler
			if prNumber != 0 {
				ctx.ContextValues.SetValue("targetType", "pull-request")

				commentHandler, err = comment.NewBitbucketPRHandler(ctx.Context(), repo, strconv.Itoa(prNumber), extra)
				if err != nil {
					return err
				}
			} else if commit != "" {
				ctx.ContextValues.SetValue("targetType", "commit")

				commentHandler, err = comment.NewBitbucketCommitHandler(ctx.Context(), repo, commit, extra)
				if err != nil {
					return err
				}
			} else {
				ui.PrintUsage(cmd)
				return fmt.Errorf("either --commit or --pull-request is required")
			}

			behavior, _ := cmd.Flags().GetString("behavior")
			if behavior != "" && !contains(validCommentBitbucketBehaviors, behavior) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--behavior only supports %s", strings.Join(validCommentBitbucketBehaviors, ", "))
			}
			ctx.ContextValues.SetValue("behavior", behavior)

			paths, _ := cmd.Flags().GetStringArray("path")

			commentOut, commentErr := buildCommentOutput(cmd, ctx, paths, output.MarkdownOptions{
				WillUpdate:          prNumber != 0 && behavior == "update",
				WillReplace:         prNumber != 0 && behavior == "delete-and-new",
				IncludeFeedbackLink: !ctx.Config.IsSelfHosted(),
				OmitDetails:         extra.OmitDetails,
				BasicSyntax:         true,
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

				if res.Posted && ctx.IsCloudUploadExplicitlyEnabled() {
					dashboardClient := apiclient.NewDashboardAPIClient(ctx)
					if err := dashboardClient.SavePostedPrComment(ctx, commentOut.AddRunResponse.RunID, commentOut.Body); err != nil {
						logging.Logger.Err(err).Msg("could not save posted PR comment")
					}
				}

				pricingClient := apiclient.GetPricingAPIClient(ctx)
				err = pricingClient.AddEvent("infracost-comment", ctx.EventEnv())
				if err != nil {
					logging.Logger.Err(err).Msg("could not report infracost-comment event")
				}

				if format == "json" {
					b, err := jsoniter.MarshalIndent(commentOut.AddRunResponse, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to marshal result: %w", err)
					}
					cmd.Print(string(b))
				} else if res.Posted {
					cmd.Println("Comment posted to Bitbucket")
				} else {
					msg := "Comment not posted to Bitbucket"
					if res.SkipReason != "" {
						msg += fmt.Sprintf(": %s", res.SkipReason)
					}
					cmd.Println(msg)
				}
			} else {
				cmd.Println(commentOut.Body)
				cmd.Println("Comment not posted to Bitbucket (--dry-run was specified)")
			}

			return commentErr
		},
	}

	cmd.Flags().String("behavior", "update", `Behavior when posting comment, one of:
  update (default)  Update latest comment
  new               Create a new comment
  delete-and-new    Delete previous matching comments and create a new comment`)
	_ = cmd.RegisterFlagCompletionFunc("behavior", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validCommentBitbucketBehaviors, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().String("bitbucket-server-url", "https://bitbucket.org", "Bitbucket Server URL")
	cmd.Flags().String("bitbucket-token", "", "Bitbucket access token. Use 'username:app-password' for Bitbucket Cloud and HTTP access token for Bitbucket Server")
	_ = cmd.MarkFlagRequired("bitbucket-token")
	cmd.Flags().String("commit", "", "Commit SHA to post comment on, mutually exclusive with pull-request. Not available when bitbucket-server-url is set")
	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	var prNumber PRNumber
	cmd.Flags().Var(&prNumber, "pull-request", "Pull request number to post comment on")
	cmd.Flags().String("repo", "", "Repository in format workspace/repo")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().Bool("exclude-cli-output", false, "Exclude CLI output so comment has just the summary table")
	cmd.Flags().String("tag", "", "Customize special text used to detect comments posted by Infracost (placed at the bottom of a comment)")
	cmd.Flags().Bool("dry-run", false, "Generate comment without actually posting to Bitbucket")
	cmd.Flags().String("format", "", "Output format: json")

	return cmd
}
