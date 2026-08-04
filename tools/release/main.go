package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/infracost/infracost/internal/logging"
)

func main() {
	cli := newAuthedGithubClient()

	releaseId := strings.TrimSpace(os.Getenv("RELEASE_ID"))
	var release *github.RepositoryRelease
	var err error
	if releaseId != "" {
		release, err = createPrerelease(cli, releaseId)
	} else {
		release, err = createDraftRelease(cli)
	}

	if err != nil {
		logging.Logger.Error().Msgf("failed to create release %s", err)
		os.Exit(1)
	}
	toUpload, err := findReleaseAssets()
	if err != nil {
		logging.Logger.Error().Msgf("failed to collect release assets %s", err)
		os.Exit(1)
	}

	err = uploadAssets(toUpload, cli, release)
	if err != nil {
		logging.Logger.Error().Msgf("failed to upload release assets %s", err)
		os.Exit(1)
	}

	logging.Logger.Info().Msg("successfully created release")
}

// createPrerelease creates a new immutable prerelease at the current commit.
// The tag must not start with "v", or scripts/get_version.sh (git tag --list
// 'v*') would select it and corrupt version detection.
func createPrerelease(cli *github.Client, tag string) (*github.RepositoryRelease, error) {
	o, res, err := cli.Repositories.CreateRelease(
		context.Background(),
		"infracost",
		"infracost",
		&github.RepositoryRelease{
			Name:                 github.String(tag),
			TagName:              github.String(tag),
			TargetCommitish:      github.String(os.Getenv("GITHUB_SHA")),
			Prerelease:           github.Bool(true),
			GenerateReleaseNotes: github.Bool(true),
		},
	)
	if err != nil {
		body := ""
		if res != nil {
			b, _ := io.ReadAll(res.Body)
			body = string(b)
		}
		return nil, fmt.Errorf("could not create prerelease %s: body: %s %w", tag, body, err)
	}

	return o, nil
}

func createDraftRelease(cli *github.Client) (*github.RepositoryRelease, error) {
	name := github.String(strings.Join(strings.Split(os.Getenv("GITHUB_REF"), "/")[2:], "/"))
	o, res, err := cli.Repositories.CreateRelease(
		context.Background(),
		"infracost",
		"infracost",
		&github.RepositoryRelease{
			Name:                 name,
			TagName:              name,
			TargetCommitish:      github.String(os.Getenv("GITHUB_SHA")),
			Draft:                github.Bool(true),
			GenerateReleaseNotes: github.Bool(true),
		},
	)
	if err != nil {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("body: %s status: %d %w", b, res.StatusCode, err)
	}

	return o, nil
}

func newAuthedGithubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	cli := github.NewClient(tc)
	return cli
}

func findReleaseAssets() ([]string, error) {
	arguments := []string{
		"./build/*.tar.gz",
		"./build/*.zip",
		"./build/*.sha256",
		"./docs/generated/docs.tar.gz",
	}

	var toUpload []string
	for _, argument := range arguments {
		files, err := filepath.Glob(filepath.Clean(argument))
		if err != nil {
			return nil, fmt.Errorf("error loading file %s from filesystem %s", argument, err)
		}

		for _, file := range files {
			if file != "." {
				toUpload = append(toUpload, file)
			}
		}
	}

	if len(toUpload) == 0 {
		return nil, errors.New("failed to find any valid release assets")
	}

	return toUpload, nil
}

func uploadAssets(toUpload []string, cli *github.Client, release *github.RepositoryRelease) error {
	errGroup := &errgroup.Group{}
	ch := make(chan string, len(toUpload))
	for _, file := range toUpload {
		ch <- file
	}
	close(ch)

	id := release.GetID()

	for i := 0; i < 4; i++ {
		errGroup.Go(func() error {
			for file := range ch {
				err := uploadAsset(file, cli, id)
				if err != nil {
					return err
				}
			}

			return nil
		})
	}

	return errGroup.Wait()
}

func uploadAsset(file string, cli *github.Client, id int64) error {
	logging.Logger.Info().Msgf("uploading asset %s", file)

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("could not open upload asset %s %s", file, err)
	}

	_, _, err = cli.Repositories.UploadReleaseAsset(
		context.Background(),
		"infracost",
		"infracost",
		id,
		&github.UploadOptions{
			Name: filepath.Base(file),
		},
		f,
	)

	if err != nil {
		return fmt.Errorf("could not upload release asset %s %s", file, err)
	}

	return nil
}
