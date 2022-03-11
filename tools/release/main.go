package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v41/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

func main() {
	cli := newAuthedGithubClient()

	release, err := createDraftRelease(cli)
	if err != nil {
		log.Errorf("failed to create draft release %s", err)
		return
	}

	toUpload, err := findReleaseAssets()
	if err != nil {
		log.Errorf("failed to collect release assets %s", err)
		return
	}

	err = uploadAssets(toUpload, cli, release)
	if err != nil {
		log.Errorf("failed to upload release assets %s", err)
		return
	}

	log.Info("successfully created draft release")
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
		b, _ := ioutil.ReadAll(res.Body)
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
	log.Infof("uploading asset %s", file)

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
