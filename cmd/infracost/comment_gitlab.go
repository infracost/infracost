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

var validCommentGitLabBehaviors = []string{"update", "new", "delete-and-new"}

func commentGitLabCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitlab",
		Short: "Post an Infracost comment to GitLab",
		Long:  "Post an Infracost comment to GitLab",
		Example: `  Update comment on a merge request:

      infracost comment gitlab --repo my-org/my-repo --merge-request 3 --path infracost.json --gitlab-token $GITLAB_TOKEN

  Post a new comment to a commit:

      infracost comment gitlab --repo my-org/my-repo --commit 2ca7182 --path infracost.json --behavior delete-and-new --gitlab-token $GITLAB_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx.ContextValues.SetValue("platform", "gitlab")

			var err error

			format, _ := cmd.Flags().GetString("format")
			format = strings.ToLower(format)
			if format != "" && !contains(validCommentOutputFormats, format) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--format only supports %s", strings.Join(validCommentOutputFormats, ", "))
			}

			serverURL, _ := cmd.Flags().GetString("gitlab-server-url")
			token, _ := cmd.Flags().GetString("gitlab-token")
			tag, _ := cmd.Flags().GetString("tag")
			extra := comment.GitLabExtra{
				ServerURL: serverURL,
				Token:     token,
				Tag:       tag,
			}

			commit, _ := cmd.Flags().GetString("commit")
			mrNumber, _ := cmd.Flags().GetInt("merge-request")
			repo, _ := cmd.Flags().GetString("repo")

			var commentHandler *comment.CommentHandler
			if mrNumber != 0 {
				ctx.ContextValues.SetValue("targetType", "merge-request")

				commentHandler, err = comment.NewGitLabPRHandler(ctx.Context(), repo, strconv.Itoa(mrNumber), extra)
				if err != nil {
					return err
				}
			} else if commit != "" {
				ctx.ContextValues.SetValue("targetType", "commit")

				commentHandler, err = comment.NewGitLabCommitHandler(ctx.Context(), repo, commit, extra)
				if err != nil {
					return err
				}
			} else {
				ui.PrintUsage(cmd)
				return fmt.Errorf("either --commit or --merge-request is required")
			}

			behavior, _ := cmd.Flags().GetString("behavior")
			if behavior != "" && !contains(validCommentGitLabBehaviors, behavior) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--behavior only supports %s", strings.Join(validCommentGitLabBehaviors, ", "))
			}
			ctx.ContextValues.SetValue("behavior", behavior)

			paths, _ := cmd.Flags().GetStringArray("path")

			commentOut, commentErr := buildCommentOutput(cmd, ctx, paths, output.MarkdownOptions{
				WillUpdate:          mrNumber != 0 && behavior == "update",
				WillReplace:         mrNumber != 0 && behavior == "delete-and-new",
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
					cmd.Println("Comment posted to GitLab")
				} else {
					msg := "Comment not posted to GitLab"
					if res.SkipReason != "" {
						msg += fmt.Sprintf(": %s", res.SkipReason)
					}
					cmd.Println(msg)
				}
			} else {
				cmd.Println(commentOut.Body)
				cmd.Println("Comment not posted to GitLab (--dry-run was specified)")
			}

			return commentErr
		},
	}

	cmd.Flags().String("behavior", "update", `Behavior when posting comment, one of:
  update (default)  Update latest comment
  new               Create a new comment
  delete-and-new    Delete previous matching comments and create a new comment`)
	_ = cmd.RegisterFlagCompletionFunc("behavior", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validCommentGitLabBehaviors, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().String("commit", "", "Commit SHA to post comment on, mutually exclusive with merge-request")
	cmd.Flags().String("gitlab-server-url", "https://gitlab.com", "GitLab Server URL")
	cmd.Flags().String("gitlab-token", "", "GitLab token")
	_ = cmd.MarkFlagRequired("gitlab-token")
	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	var mrNumber PRNumber
	cmd.Flags().Var(&mrNumber, "merge-request", "Merge request number to post comment on, mutually exclusive with commit")
	cmd.Flags().String("repo", "", "Repository in format owner/repo")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().String("tag", "", "Customize hidden markdown tag used to detect comments posted by Infracost")
	cmd.Flags().Bool("dry-run", false, "Generate comment without actually posting to GitLab")
	cmd.Flags().String("format", "", "Output format: json")

	return cmd
}
