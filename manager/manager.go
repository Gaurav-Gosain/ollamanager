package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/gaurav-gosain/ollamanager/tabs"
	"github.com/gaurav-gosain/ollamanager/tui"
	"github.com/gaurav-gosain/ollamanager/utils"
	"github.com/ollama/ollama/api"
)

func installModel(modelName string, p *tea.Program) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}

	ctx := context.Background()

	req := &api.PullRequest{
		Model: modelName,
	}
	progressFunc := func(resp api.ProgressResponse) error {
		p.Send(resp)
		return nil
	}

	err = client.Pull(ctx, req, progressFunc)
	if err != nil {
		fmt.Println("Error pulling model:", err.Error())
		return
	}
}

// deleteModel deletes a model by name It returns an error if the model is not
// found or if any other error occurs.
func deleteModel(modelName string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %s", err.Error())
	}

	ctx := context.Background()

	req := &api.DeleteRequest{
		Model: modelName,
	}

	err = client.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete model: %s", err.Error())
	}

	return nil
}

func loadModel(modelName string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %s", err.Error())
	}

	ctx := context.Background()

	req := &api.GenerateRequest{
		Model: modelName,
		KeepAlive: &api.Duration{
			Duration: -1,
		},
	}

	err = client.Generate(ctx, req, func(g api.GenerateResponse) error { return nil })
	if err != nil {
		return fmt.Errorf("failed to load model: %s", err.Error())
	}

	fmt.Println(
		lipgloss.NewStyle().Padding(0, 2).Render(
			fmt.Sprintln(
				"Model",
				tui.StatusStyle.Render(modelName),
				"will stay loaded in memory",
				tui.StatusStyle.Render("indefinitely"),
			),
		),
	)

	return nil
}

func freeModel(modelName string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %s", err.Error())
	}

	ctx := context.Background()

	req := &api.GenerateRequest{
		Model: modelName,
		KeepAlive: &api.Duration{
			Duration: 0,
		},
	}

	err = client.Generate(ctx, req, func(g api.GenerateResponse) error { return nil })
	if err != nil {
		return fmt.Errorf("failed to free model: %s", err.Error())
	}

	fmt.Println(
		lipgloss.NewStyle().Padding(0, 2).Render(
			fmt.Sprintln(
				"Model",
				tui.StatusStyle.Render(modelName),
				tui.StatusStyle.Render("unloaded"),
				"from memory",
			),
		),
	)

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

func Run(
	selectedTabs []tabs.Tab,
	approvedActions []tabs.InstalledAction,
) (
	action tabs.Tab,
	manageAction tabs.InstalledAction,
	modelName string,
	err error,
) {
	ctx := context.Background()

	modelSelector, err := tui.ModelPicker(
		selectedTabs,
		approvedActions,
	)

	// TODO: could be cleaner
	if err != nil ||
		(modelSelector.SelectedInstallableModel.Name == "" &&
			modelSelector.SelectedInstalledModel.Name == "" &&
			modelSelector.SelectedRunningModel.Name == "") {
		utils.ClearTerminal()
		err = errors.New(`failed to pick a model :(`)
		return
	}

	if modelSelector.Action == tabs.INSTALL {

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
			err = spinnerErr
			return
		}

		if err != nil {
			return
		}

		if len(modelTags) == 0 {
			err = errors.New("couldn't load tags for " + modelName + "  :(")
			return
		}

		options := []huh.Option[string]{}

		for _, modelTag := range modelTags {
			options = append(options, huh.NewOption(modelTag, modelTag))
		}

		var tag string

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
		).WithProgramOptions(tea.WithAltScreen())

		err = form.Run()
		if err != nil {
			return
		}

		if !confirm {
			err = errors.New("see you")
			return
		}
		modelName = fmt.Sprintf("%s:%s", modelName, tag)
	} else {
		modelName = modelSelector.SelectedInstalledModel.Name
	}

	switch tabs.Tab(modelSelector.Action) {
	case tabs.INSTALLED:
		modelName = modelSelector.SelectedInstalledModel.Name
	case tabs.RUNNING:
		modelName = modelSelector.SelectedRunningModel.Name
	}

	utils.ClearTerminal()

	fmt.Println(
		lipgloss.NewStyle().Padding(1, 2).Render(
			fmt.Sprintf(
				"Picked action %s on model %s",
				tui.StatusStyle.SetString(string(modelSelector.Action)).String(),
				tui.StatusStyle.SetString(modelName).String(),
			),
		),
	)

	m := tui.InstallModel{
		Progress: progress.New(progress.WithDefaultGradient()),
		Spinner:  tui.InitSpinner(),
	}
	// Start Bubble Tea
	p := tea.NewProgram(m)

	var actionErr error
	var res tea.Model

	switch modelSelector.Action {
	case tabs.INSTALL:
		go installModel(modelName, p)
		res, err = p.Run()
		if err != nil {
			fmt.Println("error running program:", err.Error())
			return
		}

		actionErr = res.(tui.InstallModel).Err
	case tabs.INSTALLED:
		switch modelSelector.ManageAction {
		case tabs.UPDATE:
			go installModel(modelName, p)

			res, err = p.Run()
			if err != nil {
				fmt.Println("error running program:", err.Error())
				return
			}

			actionErr = res.(tui.InstallModel).Err
		case tabs.DELETE:
			modelName = modelSelector.SelectedInstalledModel.Name
			actionErr = deleteModel(modelName)
		}
	case tabs.RUNNING:
		// TODO: Implement running model
		modelName = modelSelector.SelectedRunningModel.Name
		fmt.Println(
			lipgloss.NewStyle().Padding(0, 2).Render(
				fmt.Sprintln("Model", modelSelector.SelectedRunningModel.Name, "is running..."),
			),
		)

		var runningAction string
		var confirm bool

		form := huh.NewForm(
			huh.NewGroup(
				// Ask the user for a base burger and toppings.
				huh.NewSelect[string]().
					Title("Choose your tag for "+modelName).
					Options(
						huh.NewOption("Keep loaded in memory (indefinitely)", "load"),
						huh.NewOption("Free up memory by unloading", "free"),
						huh.NewOption("Do nothing", "none"),
					).
					Value(&runningAction),
				huh.NewConfirm().
					Title("Would you like to continue?").
					Value(&confirm),
			),
		).WithProgramOptions(tea.WithAltScreen())

		err = form.Run()
		if err != nil {
			return
		}

		if !confirm {
			break
		}

		switch runningAction {
		case "load":
			actionErr = loadModel(modelName)
		case "free":
			actionErr = freeModel(modelName)
		case "none":
			actionErr = nil
		}

	}

	action = modelSelector.Action
	manageAction = modelSelector.ManageAction
	err = actionErr
	return
}
