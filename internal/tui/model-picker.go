package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

func ModelPicker(tabs []string, baseURL string) (result ModelSelector, err error) {
	ctx := context.Background()

	var models []OllamaModel

	loadModels := func() {
		models, err = GetAvailableModels()
		ctx.Done() // signal that model fetching is done
	}

	spinnerErr := spinner.
		New().
		Title("Loading available models...").
		Action(loadModels).
		Run()

	if spinnerErr != nil {
		return
	}

	if err != nil {
		return
	}

	if len(models) == 0 {
		return
	}

	installableItems := []list.Item{}

	for _, model := range models {
		installableItems = append(installableItems, list.Item(model))
	}

	ctx = context.Background()

	var installedModels []InstalledOllamaModel

	loadModels = func() {
		installedModels, err = GetInstalledModels(baseURL)
		ctx.Done() // signal that model fetching is done
	}

	spinnerErr = spinner.
		New().
		Title("Loading available models...").
		Action(loadModels).
		Run()

	if spinnerErr != nil {
		return
	}

	if err != nil {
		return
	}

	if len(models) == 0 {
		return
	}

	installedItems := []list.Item{}

	for _, model := range installedModels {
		installedItems = append(installedItems, list.Item(model))
	}

	installableModelsList := list.New(installableItems, list.NewDefaultDelegate(), 0, 0)
	installableModelsList.Title = "Pick a Model..."
	installableModelsList.SetShowHelp(false)

	installedModelsList := list.New(installedItems, list.NewDefaultDelegate(), 0, 0)
	installedModelsList.Title = "Pick a Model..."
	installedModelsList.SetShowHelp(false)

	helpModel := help.New()
	helpModel.ShowAll = true
	helpModel.Styles.FullDesc.UnsetForeground()
	helpModel.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"})

	m := ModelSelector{
		installableList: installableModelsList,
		installedList:   installedModelsList,
		Tabs:            tabs,
		help:            helpModel,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	resModel, err := p.Run()
	if err != nil {
		return
	}

	return resModel.(ModelSelector), nil
}
