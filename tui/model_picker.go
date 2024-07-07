package tui

import (
	"context"
	"errors"
	"slices"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/gaurav-gosain/ollamanager/tabs"
)

func ModelPicker(
	selectedTabs []tabs.Tab,
	approvedActions []tabs.ManageAction,
) (result ModelSelector, err error) {
	ctx := context.Background()

	var spinnerErr error

	var installableItems, installedItems, runningItems []list.Item

	var installableModelsList, installedModelsList, runningModelsList list.Model

	var models []OllamaModel
	var installedModels []InstalledOllamaModel
	var runningModels []RunningOllamaModel

	var loadModels func()

	hasInstallTab := slices.Contains(selectedTabs, tabs.INSTALL)

	if len(selectedTabs) == 0 {
		return
	}

	if hasInstallTab {
		loadModels = func() {
			models, err = GetAvailableModels()
			ctx.Done() // signal that model fetching is done
		}
		spinnerErr = spinner.
			New().
			Title("Loading installable models...").
			Action(loadModels).
			Run()

		if spinnerErr != nil {
			return
		}

		if err != nil {
			return
		}

		for _, model := range models {
			installableItems = append(installableItems, list.Item(model))
		}

		installableModelsList = list.New(installableItems, list.NewDefaultDelegate(), 0, 0)
		installableModelsList.Title = "Pick a Model to install..."
		installableModelsList.SetShowHelp(false)
	}

	loadModels = func() {
		installedModels, err = GetInstalledModels()
		ctx.Done() // signal that model fetching is done
	}

	spinnerErr = spinner.
		New().
		Title("Loading installed models...").
		Action(loadModels).
		Run()

	if spinnerErr != nil {
		return
	}

	if err != nil {
		return
	}

	for _, model := range installedModels {
		installedItems = append(installedItems, list.Item(model))
	}

	installedModelsList = list.New(installedItems, list.NewDefaultDelegate(), 0, 0)
	installedModelsList.Title = "Pick an installed Model..."
	installedModelsList.SetShowHelp(false)

	loadModels = func() {
		runningModels, err = GetRunningModels()
		ctx.Done() // signal that model fetching is done
	}

	spinnerErr = spinner.
		New().
		Title("Loading running models...").
		Action(loadModels).
		Run()

	if spinnerErr != nil {
		return
	}

	// if err != nil {
	// 	return
	// }

	for _, model := range runningModels {
		runningItems = append(runningItems, list.Item(model))
	}

	runningModelsList = list.New(runningItems, list.NewDefaultDelegate(), 0, 0)
	runningModelsList.Title = "Pick a running Model..."
	runningModelsList.SetShowHelp(false)

	helpModel := help.New()
	helpModel.ShowAll = true
	helpModel.Styles.FullDesc.UnsetForeground()
	helpModel.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"})

	m := ModelSelector{
		installableList: installableModelsList,
		installedList:   installedModelsList,
		runningList:     runningModelsList,
		Tabs:            selectedTabs,
		ApprovedActions: approvedActions,
		help:            helpModel,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	resModel, err := p.Run()
	if err != nil {
		return
	}

	model := resModel.(ModelSelector)

	// TODO: could be cleaner
	if err != nil ||
		(model.SelectedInstallableModel.Name == "" &&
			model.SelectedInstalledModel.Name == "" &&
			model.SelectedRunningModel.Name == "") {
		err = errors.New(`failed to pick a model :(`)
		return
	}

	// if manage tab was selected, but no valid action was selected, show a form, otherwise continue
	if model.Action == tabs.MANAGE && !slices.Contains(model.ApprovedActions, model.ManageAction) {

		var validOptions []huh.Option[tabs.ManageAction]

		for _, option := range model.ApprovedActions {
			validOptions = append(validOptions, huh.NewOption(string(option), option))
		}

		form := huh.NewForm(
			huh.NewGroup(
				// Ask the user for a base burger and toppings.
				huh.NewSelect[tabs.ManageAction]().
					Title("Choose the action for " + model.SelectedInstalledModel.Name).
					Options(validOptions...).
					Value(&model.ManageAction),
			),
		).WithProgramOptions(tea.WithAltScreen())

		err = form.Run()
		if err != nil {
			return
		}
	}

	return model, nil
}
