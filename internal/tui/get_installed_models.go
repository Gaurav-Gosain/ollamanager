package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	humanize "github.com/dustin/go-humanize"
)

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

type InstalledOllamaModel struct {
	Details struct {
		Format            string `json:"format"`
		Family            string `json:"family"`
		Families          any    `json:"families"`
		ParameterSize     string `json:"parameter_size"`
		QuantizationLevel string `json:"quantization_level"`
	} `json:"details"`
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Digest     string `json:"digest"`
	Size       int64  `json:"size"`
}

type TagResponse struct {
	Models []InstalledOllamaModel `json:"models"`
}

func GetInstalledModels(baseURL string) ([]InstalledOllamaModel, error) {
	url := baseURL + "/api/tags"

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	var tags TagResponse
	err = decoder.Decode(&tags)
	if err != nil {
		return nil, err
	}

	return tags.Models, nil
}

func (model InstalledOllamaModel) Title() string {
	return model.Name
}

func (model InstalledOllamaModel) Description() string {
	modifiedAt, _ := time.Parse(time.RFC3339Nano, model.ModifiedAt)

	return fmt.Sprintf(
		"%s • %s • %s",
		humanize.Bytes(uint64(model.Size)),
		model.Details.ParameterSize,
		humanize.Time(modifiedAt),
	)
}
func (model InstalledOllamaModel) FilterValue() string { return model.Name }
