package infracost

import (
	"context"
	"github.com/infracost/infracost/internal/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

const ConfigFilePathKey = "INFRACOST_CONFIG"

func BreakdownJsonFile(rootCtx context.Context, jsonPlan []byte) ([]byte, error) {
	return processRequest(rootCtx, jsonPlan, false)
}

func DiffFile(rootCtx context.Context, jsonPlan []byte) ([]byte, error) {
	return processRequest(rootCtx, jsonPlan, false)
}

func processRequest(rootCtx context.Context, jsonPlan []byte, diff bool) ([]byte, error) {
	ctx, err := config.NewRunContextFromEnv(rootCtx)
	if err != nil {
		return nil, err
	}

	if confFile, ok := rootCtx.Value(ConfigFilePathKey).(string); ok {
		err = loadConfiguration(ctx, confFile)
	}

	err = configureExecution(ctx)
	if err != nil {
		return nil, err
	}
	planFile, err := temporarySavePlan(jsonPlan)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Remove(planFile)
		if err != nil {
			log.Infof("failed to remove a temporary plan file %s ", err)
		}
	}()
	err = assignProject(ctx, planFile)
	if err != nil {
		return nil, err
	}

	return runMain(ctx, diff)
}

func configureExecution(ctx *config.RunContext) error {
	ctx.Config.NoColor = true
	ctx.Config.LogLevel = "debug"
	ctx.SetContextValue("isDefaultPricingAPIEndpoint", ctx.Config.PricingAPIEndpoint == ctx.Config.DefaultPricingAPIEndpoint)
	ctx.Config.Format = "json"
	ctx.Config.ShowSkipped = true
	ctx.SetContextValue("outputFormat", ctx.Config.Format)

	return nil
}

func assignProject(ctx *config.RunContext, tmpPath string) error {
	//TODO: refactor it. This is a workaround. Use byte and custom terraform Json provider instead of creating a temp file
	projectCfg := &config.Project{}

	ctx.Config.Projects = []*config.Project{
		projectCfg,
	}
	projectCfg.Path = tmpPath
	return nil
}

func temporarySavePlan(jsonPlan []byte) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "plan*.tf.json")
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(file.Name(), jsonPlan, 0644)
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func loadConfiguration(ctx *config.RunContext, configPath string) error {
	return ctx.Config.LoadFromConfigFile(configPath)
}
