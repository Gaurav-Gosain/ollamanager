package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

func ModelPicker(tabs []string) (model *OllamaModel, err error) {
	ctx := context.Background()

	var models []OllamaModel

	loadModels := func() {
		models, err = GetAvailableModels()
		ctx.Done() // signal that model fetching is done
	}

	spinnerErr := spinner.
		New().
		Title("Loading models...").
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

	items := []list.Item{}

	for _, model := range models {
		items = append(items, list.Item(model))
	}

	listModel := list.New(items, list.NewDefaultDelegate(), 0, 0)

	listModel.SetShowHelp(false)

	helpModel := help.New()
	helpModel.ShowAll = true
	helpModel.Styles.FullDesc.UnsetForeground()
	helpModel.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"})

	m := ModelSelector{list: listModel, Tabs: tabs, help: helpModel}
	m.list.Title = "Pick a Model..."

	p := tea.NewProgram(m, tea.WithAltScreen())

	resModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	res := resModel.(ModelSelector).SelectedModel

	return &res, nil
}
