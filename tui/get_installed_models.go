package tui

import (
	"context"
	"fmt"
	"log"

	humanize "github.com/dustin/go-humanize"
	"github.com/ollama/ollama/api"
)

type InstalledOllamaModel api.ListModelResponse

type TagResponse struct {
	Models []InstalledOllamaModel `json:"models"`
}

func GetInstalledModels() ([]InstalledOllamaModel, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	installedModels := make([]InstalledOllamaModel, len(list.Models))
	for i, model := range list.Models {
		installedModels[i] = InstalledOllamaModel(model)
	}

	return installedModels, nil
}

func (model InstalledOllamaModel) Title() string {
	return model.Name
}

func (model InstalledOllamaModel) Description() string {
	return fmt.Sprintf(
		"%s • %s • %s",
		humanize.Bytes(uint64(model.Size)),
		model.Details.ParameterSize,
		humanize.Time(model.ModifiedAt),
	)
}
func (model InstalledOllamaModel) FilterValue() string { return model.Name }
