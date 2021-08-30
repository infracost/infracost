package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

var defaultTerragruntBinary = "terragrunt"
var minTerragruntVersion = "v0.31.0"

type TerragruntProvider struct {
	ctx              *config.ProjectContext
	TerragruntBinary string
	Path             string
}

type TerragruntInfo struct {
	WorkingDir string
}

func NewTerragruntProvider(ctx *config.ProjectContext) schema.Provider {

	terragruntBinary := ctx.ProjectConfig.TerraformBinary
	if terragruntBinary == "" {
		terragruntBinary = defaultTerragruntBinary
	}

	return &TerragruntProvider{
		ctx:              ctx,
		TerragruntBinary: terragruntBinary,
		Path:             ctx.ProjectConfig.Path,
	}
}

func (p *TerragruntProvider) Type() string {
	return "terragrunt"
}

func (p *TerragruntProvider) DisplayType() string {
	return "Terragrunt directory"
}

func (p *TerragruntProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TerragruntProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	paths, err := p.getProjectPaths()
	if err != nil {
		return []*schema.Project{}, err
	}

	parallelism := 4
	projects := make([]*schema.Project, 0)

	var waitGroup sync.WaitGroup
	var semaphore = make(chan struct{}, parallelism) // Make a semaphore from a buffered channel
	errs := make(chan error)

	for _, path := range paths {
		waitGroup.Add(1)
		go func(path string) {
			defer waitGroup.Done()
			pathProjects, err := p.loadProjectWhenReady(semaphore, path, usage)
			if err != nil {
				errs <- err
			}

			projects = append(projects, pathProjects...)
		}(path)
	}

	waitGroup.Wait()
	close(errs)

	var multiErr *multierror.Error
	for err := range errs {
		multiErr = multierror.Append(multiErr, err)
	}

	fmt.Println(len(projects))

	return projects, multiErr.ErrorOrNil()
}

func (p *TerragruntProvider) loadProjectWhenReady(semaphore chan struct{}, path string, usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	semaphore <- struct{}{} // Add one to the buffered channel. Will block if parallelism limit is met
	defer func() {
		<-semaphore // Remove one from the buffered channel
	}()

	projectCfg := &config.Project{
		Path:                path,
		TerraformPlanFlags:  p.ctx.ProjectConfig.TerraformPlanFlags,
		TerraformBinary:     p.TerragruntBinary,
		TerraformWorkspace:  p.ctx.ProjectConfig.TerraformWorkspace,
		TerraformCloudHost:  p.ctx.ProjectConfig.TerraformCloudHost,
		TerraformCloudToken: p.ctx.ProjectConfig.TerraformCloudToken,
		UsageFile:           p.ctx.ProjectConfig.UsageFile,
		TerraformUseState:   p.ctx.ProjectConfig.TerraformUseState,
	}

	ctx := config.NewProjectContext(p.ctx.RunContext, projectCfg)
	ctx.SetContextValues(p.ctx.ContextValues())

	dirProvider := NewDirProvider(ctx).(*DirProvider)

	return dirProvider.LoadResources(usage)
}

func (p *TerragruntProvider) getProjectPaths() ([]string, error) {
	opts := &CmdOptions{
		TerraformBinary: p.TerragruntBinary,
		Dir:             p.Path,
	}
	out, err := Cmd(opts, "run-all", "terragrunt-info")
	if err != nil {
		return []string{}, err
	}

	jsons := bytes.SplitAfter(out, []byte{'}', '\n'})
	if len(jsons) > 1 {
		jsons = jsons[:len(jsons)-1]
	}

	paths := make([]string, 0, len(jsons))
	for _, j := range jsons {
		fmt.Println(string(j))
		var info TerragruntInfo
		err = json.Unmarshal(j, &info)
		if err != nil {
			return paths, err
		}

		paths = append(paths, info.WorkingDir)
	}

	return paths, nil
}
