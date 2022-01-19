package main

import (
	"fmt"
	"github.com/infracost/infracost/internal/comment"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/spf13/cobra"
	"strconv"
)

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
				commentHandler, err = comment.NewGitHubPRHandler(ctx.Context(), repo, strconv.Itoa(prNumber), extra)
				if err != nil {
					return err
				}
			} else if commit != "" {
				commentHandler, err = comment.NewGitHubCommitHandler(ctx.Context(), repo, commit, extra)
				if err != nil {
					return err
				}
			} else {
				ui.PrintUsage(cmd)
				return fmt.Errorf("either --commit or --pull-request is required")
			}

			behavior, err := getBehaviorFlag(cmd)
			if err != nil {
				return err
			}

			paths, _ := cmd.Flags().GetStringArray("path")

			body, err := buildCommentBody(ctx, paths)
			if err != nil {
				return err
			}

			err = commentHandler.CommentWithBehavior(ctx.Context(), behavior, string(body))
			if err != nil {
				return err
			}

			if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
				err = saveOutFile(cmd, outFile, body)
				if err != nil {
					return err
				}
			} else {
				cmd.Println(string(body))
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
		return validCommentBehaviors, cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().String("commit", "", "Commit SHA to post/get the comment, mutually exclusive with pull-request")
	cmd.Flags().String("github-api-url", "https://api.github.com", "GitHub API URL, defaults to https://api.github.com")
	cmd.Flags().String("github-token", "", "GitHub token")
	_ = cmd.MarkFlagRequired("github-token")
	cmd.Flags().StringP("out-file", "o", "", "Save output to a file, helpful with format flag")
	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	cmd.Flags().Int("pull-request", 0, "Pull request number to post the comment on, mutually exclusive with commit")
	cmd.Flags().String("repo", "", "Repository in the format owner/repo")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().String("tag", "", "Customize the embedded tag that is used for detecting comments posted by Infracost")

	return cmd
}
