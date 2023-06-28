package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/comment"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

var validCommentGitHubBehaviors = []string{"update", "new", "hide-and-new", "delete-and-new"}

func commentGitHubCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Post an Infracost comment to GitHub",
		Long:  "Post an Infracost comment to GitHub",
		Example: `  Update comment on a pull request:

      infracost comment github --repo my-org/my-repo --pull-request 3 --path infracost.json --github-token $GITHUB_TOKEN

  Post a new comment to a commit:

      infracost comment github --repo my-org/my-repo --commit 2ca7182 --path infracost.json --behavior hide-and-new --github-token $GITHUB_TOKEN`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx.SetContextValue("platform", "github")

			var err error

			apiURL, _ := cmd.Flags().GetString("github-api-url")
			token, _ := cmd.Flags().GetString("github-token")
			tag, _ := cmd.Flags().GetString("tag")

			tlsCertFile, _ := cmd.Flags().GetString("github-tls-cert-file")
			tlsKeyFile, _ := cmd.Flags().GetString("github-tls-key-file")
			tlsInsecureSkipVerify, _ := cmd.Flags().GetBool("github-tls-insecure-skip-verify")

			tlsConfig := tls.Config{} // nolint: gosec

			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			tlsConfig.RootCAs = rootCAs
			tlsConfig.InsecureSkipVerify = tlsInsecureSkipVerify // nolint: gosec

			if tlsCertFile != "" && tlsKeyFile != "" {
				cert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
				if err != nil {
					return errors.Wrap(err, "Error loading TLS certificate and key")
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
			}

			extra := comment.GitHubExtra{
				APIURL:    apiURL,
				Token:     token,
				Tag:       tag,
				TLSConfig: &tlsConfig,
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

			body, hasDiff, err := buildCommentBody(cmd, ctx, paths, output.MarkdownOptions{
				WillUpdate:          prNumber != 0 && behavior == "update",
				WillReplace:         prNumber != 0 && behavior == "delete-and-new",
				IncludeFeedbackLink: !ctx.Config.IsSelfHosted(),
				MaxMessageSize:      output.GitHubMaxMessageSize,
			})
			var policyFailure output.PolicyCheckFailures
			var guardrailFailure output.GuardrailFailures
			var tagPolicyFailure *output.TagPolicyCheck
			if err != nil {
				if v, ok := err.(output.PolicyCheckFailures); ok {
					policyFailure = v
				} else if v, ok := err.(output.GuardrailFailures); ok {
					guardrailFailure = v
				} else if v, ok := err.(output.TagPolicyCheck); ok {
					tagPolicyFailure = &v
				} else {
					return err
				}
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if !dryRun {
				skipNoDiff, _ := cmd.Flags().GetBool("skip-no-diff")

				posted, err := commentHandler.CommentWithBehavior(ctx.Context(), !hasDiff && skipNoDiff, behavior, string(body))
				if err != nil {
					return err
				}

				pricingClient := apiclient.NewPricingAPIClient(ctx)
				err = pricingClient.AddEvent("infracost-comment", ctx.EventEnv())
				if err != nil {
					logging.Logger.WithError(err).Error("could not report infracost-comment event")
				}

				if posted {
					cmd.Println("Comment posted to GitHub")
				} else {
					cmd.Println("Comment not posted to GitHub (skipped)")
				}
			} else {
				cmd.Println(string(body))
				cmd.Println("Comment not posted to GitHub (--dry-run was specified)")
			}

			if policyFailure != nil {
				cmd.Printf("\n")
				return policyFailure
			}
			if guardrailFailure != nil {
				cmd.Printf("\n")
				return guardrailFailure
			}
			if tagPolicyFailure != nil {
				cmd.Printf("\n")
				return tagPolicyFailure
			}

			return nil
		},
	}

	cmd.Flags().String("behavior", "update", `Behavior when posting comment, one of:
  update (default)  Update latest comment
  new               Create a new comment
  hide-and-new      Hide previous matching comments and create a new comment
  delete-and-new    Delete previous matching comments and create a new comment`)
	_ = cmd.RegisterFlagCompletionFunc("behavior", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validCommentGitHubBehaviors, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().String("commit", "", "Commit SHA to post comment on, mutually exclusive with pull-request")
	cmd.Flags().String("github-api-url", "https://api.github.com", "GitHub API URL")
	cmd.Flags().String("github-token", "", "GitHub token")
	_ = cmd.MarkFlagRequired("github-token")
	cmd.Flags().String("github-tls-cert-file", "", "Path to optional client certificate file when communicating with GitHub Enterprise API")
	cmd.Flags().String("github-tls-key-file", "", "Path to optional client key file when communicating with GitHub Enterprise API")
	cmd.Flags().Bool("github-tls-insecure-skip-verify", false, "Skip TLS certificate checks for GitHub Enterprise API")
	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	var prNumber PRNumber
	cmd.Flags().Var(&prNumber, "pull-request", "Pull request number to post comment on, mutually exclusive with commit")
	cmd.Flags().String("repo", "", "Repository in format owner/repo")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().String("tag", "", "Customize hidden markdown tag used to detect comments posted by Infracost")
	cmd.Flags().Bool("dry-run", false, "Generate comment without actually posting to GitHub")

	return cmd
}
