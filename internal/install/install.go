package install

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/gaurav-gosain/ollamanager/internal/tui"
)

// deleteModel deletes a model by name via a DELETE request to the specified API endpoint.
// It returns an error if the model is not found or if any other error occurs.
func deleteModel(url, modelName string) error {
	// Create the request body
	requestBody, err := json.Marshal(map[string]string{"name": modelName})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the DELETE request
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode == http.StatusNotFound {
		return errors.New("model not found")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func GetAvailableTags(modelName string) ([]string, error) {
	resp, err := http.Get("https://ollama.com/library/" + modelName + "/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	var items []string
	for _, line := range lines {
		if strings.Contains(line, `href="/library/`+modelName+":") {
			href := strings.Split(line, `href="`)[1]
			item := strings.Split(href, `"`)[0]
			item = strings.TrimPrefix(item, "/library/"+modelName+":")
			items = append(items, item)
		}
	}

	return items, nil
}

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ClearTerminal() {
	switch runtime.GOOS {
	case "darwin":
		runCmd("clear")
	case "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}

func Run(baseURL string) (string, string, string, error) {
	ctx := context.Background()

	modelSelector, err := tui.ModelPicker(
		[]string{
			tui.INSTALL,
			tui.UPDATE,
			tui.DELETE,
		},
		baseURL,
	)
	if err != nil || (modelSelector.SelectedInstallableModel.Name == "" && modelSelector.SelectedInstalledModel.Name == "") {
		ClearTerminal()
		return "", "", "", errors.New(`failed to pick a model :(`)
	}

	var downloadingModel string
	var modelName string
	var tag string

	if modelSelector.Action == tui.INSTALL {

		model := modelSelector.SelectedInstallableModel

		modelName = model.Name

		confirm := false

		ctx = context.Background()

		var modelTags []string

		loadModelTags := func() {
			modelTags, err = GetAvailableTags(modelName)
			ctx.Done() // signal that model fetching is done
		}

		spinnerErr := spinner.
			New().
			Title("Loading tags for " + modelName + "...").
			Action(loadModelTags).
			Run()

		if spinnerErr != nil {
			return "", "", "", spinnerErr
		}

		if err != nil {
			return "", "", "", err
		}

		if len(modelTags) == 0 {
			return "", "", "", errors.New("couldn't load tags for " + modelName + "  :(")
		}

		options := []huh.Option[string]{}

		for _, modelTag := range modelTags {
			options = append(options, huh.NewOption(modelTag, modelTag))
		}

		form := huh.NewForm(
			huh.NewGroup(
				// Ask the user for a base burger and toppings.
				huh.NewSelect[string]().
					Title("Choose your tag for "+modelName).
					Options(
						options...,
					).
					Value(&tag), // store the chosen option in the "modelName" variable

				huh.NewConfirm().
					Title("Would you like to continue?").
					Value(&confirm),
			),
		)

		err = form.Run()
		if err != nil {
			return "", "", "", err
		}

		if !confirm {
			return "", "", "", errors.New("see you")
		}
		downloadingModel = fmt.Sprintf("%s:%s", modelName, tag)
	} else {
		downloadingModel = modelSelector.SelectedInstalledModel.Name
		modelInfo := strings.Split(downloadingModel, ":")
		if len(modelInfo) == 2 {
			modelName = modelInfo[0]
			tag = modelInfo[1]
		} else {
			modelName = downloadingModel
			tag = "?"
		}
	}

	ClearTerminal()

	fmt.Println("Starting to", modelSelector.Action, tui.StatusStyle.SetString(downloadingModel).String())

	m := tui.InstallModel{
		Progress: progress.New(progress.WithDefaultGradient()),
		Spinner:  tui.InitSpinner(),
	}
	// Start Bubble Tea
	p := tea.NewProgram(m)

	var actionErr error

	if modelSelector.Action == tui.INSTALL || modelSelector.Action == tui.UPDATE {
		go func(apiUrl string, modelName string, p *tea.Program) {
			// Prepare request body
			requestBody, err := json.Marshal(map[string]string{
				"name": modelName,
			})
			if err != nil {
				fmt.Println("Error marshalling request body:", err)
				return
			}

			// Send POST request to the API
			resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				fmt.Println("Error sending POST request:", err)
				return
			}
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)

			for {
				var response tui.Response
				if err := decoder.Decode(&response); err != nil {
					fmt.Println("Error decoding response:", err)
					p.Quit()
					return
				}

				p.Send(response)

				if response.Status == "success" {
					break
				}
			}
		}(fmt.Sprintf("%s/api/pull", baseURL), downloadingModel, p)
		res, err := p.Run()
		if err != nil {
			fmt.Println("error running program:", err)
			return "", "", "", err
		}

		actionErr = res.(tui.InstallModel).Err
	} else {
		actionErr = deleteModel(fmt.Sprintf("%s/api/delete", baseURL), downloadingModel)
	}

	return modelSelector.Action, modelName, tag, actionErr
}
