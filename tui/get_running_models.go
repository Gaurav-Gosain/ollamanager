package tui

import (
	"context"
	"fmt"
	"log"

	humanize "github.com/dustin/go-humanize"
	"github.com/ollama/ollama/api"
)

type RunningOllamaModel api.ProcessModelResponse

type PSResponse struct {
	Models []RunningOllamaModel `json:"models"`
}

func GetRunningModels() ([]RunningOllamaModel, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	list, err := client.ListRunning(ctx)
	if err != nil {
		return nil, err
	}

	runningModels := make([]RunningOllamaModel, len(list.Models))
	for i, model := range list.Models {
		runningModels[i] = RunningOllamaModel(model)
	}

	return runningModels, nil
}

func (model RunningOllamaModel) Title() string {
	return model.Name
}

func (model RunningOllamaModel) Description() string {
	return fmt.Sprintf(
		"%s • %s",
		model.Details.ParameterSize,
		humanize.Time(model.ExpiresAt),
	)
}
func (model RunningOllamaModel) FilterValue() string { return model.Name }
