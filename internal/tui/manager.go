package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// Style definitions.
var (
	// Paragraphs.

	layoutStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			AlignVertical(lipgloss.Center)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4"))

	descStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4"))

	// Page.

	docStyle          = lipgloss.NewStyle().Margin(1)
	helpStyle         = list.DefaultStyles().HelpStyle
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	dimTextColor      = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

const (
	INSTALL = "Install"
	UPDATE  = "Update"
	DELETE  = "Delete"
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

type ModelSelector struct {
	installableList          list.Model
	installedList            list.Model
	help                     help.Model
	SelectedInstallableModel OllamaModel
	SelectedInstalledModel   InstalledOllamaModel
	Action                   string
	Tabs                     []string
	width                    int
	height                   int
	ActiveTab                int
	infoVisible              bool
	helpVisible              bool
}

func (m ModelSelector) Init() tea.Cmd {
	return nil
}

func (m ModelSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	installAction := m.Tabs[m.ActiveTab] == INSTALL

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.installableList.FilterState() == list.Filtering || m.installedList.FilterState() == list.Filtering {
			break
		}

		if m.helpVisible {
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "?", "esc":
				m.helpVisible = !m.helpVisible
				return m, nil
			}
			break
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.helpVisible = !m.helpVisible
			return m, nil
		case "n", "tab":
			m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			return m, nil
		case "p", "shift+tab":
			m.ActiveTab = max(m.ActiveTab-1, 0)
			return m, nil
		case "enter":
			if installAction {
				m.SelectedInstallableModel = m.installableList.SelectedItem().(OllamaModel)
			} else {
				m.SelectedInstalledModel = m.installedList.SelectedItem().(InstalledOllamaModel)
			}
			m.Action = m.Tabs[m.ActiveTab]
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.height = msg.Height - v - 1
		m.width = msg.Width - h - 2
		listWidth := m.width
		if m.width > 90 {
			listWidth = 3 * m.width / 5
			m.infoVisible = true
		} else {
			m.infoVisible = false
		}
		m.help.Width = 8 * m.width / 10
		v = activeTabStyle.GetVerticalFrameSize()
		m.installableList.SetSize(listWidth, m.height-v)
		m.installedList.SetSize(listWidth, m.height-v)
	}

	var cmd tea.Cmd

	if !m.helpVisible {
		if installAction {
			m.installableList, cmd = m.installableList.Update(msg)
		} else {
			m.installedList, cmd = m.installedList.Update(msg)
		}
	}
	return m, cmd
}

func (m ModelSelector) View() string {
	var renderedTabs []string

	numTabs := len(m.Tabs)
	tabWidth := m.width / numTabs
	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isActive := i == 0, i == m.ActiveTab
		w := tabWidth
		if isFirst {
			w = m.width - ((numTabs - 1) * tabWidth) - 2
		}
		if isActive {
			style = activeTabStyle.Copy().Width(w)
		} else {
			style = inactiveTabStyle.Copy().Foreground(dimTextColor).Width(w)
		}

		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	activeTabContent := ""

	v := activeTabStyle.GetVerticalFrameSize()

	var list list.Model

	switch m.Tabs[m.ActiveTab] {
	case INSTALL:
		list = m.installableList
	default:
		list = m.installedList
	}

	frames := []string{
		layoutStyle.Copy().
			Padding(0, 2).
			Width(m.installableList.Width()).
			Height(m.height - v).
			BorderForeground(lipgloss.Color("69")).
			Render(list.View()),
	}

	selectedItem := list.SelectedItem()

	if m.infoVisible {

		info := fmt.Sprintf("%s not found", titleStyle.Render(fmt.Sprintf(" %s ", list.FilterValue())))

		switch m.Tabs[m.ActiveTab] {
		case INSTALL:
			if selectedItem != nil {
				selectedModel := m.installableList.SelectedItem().(OllamaModel)
				extraInfo := ""
				if len(selectedModel.ExtraInfo) > 0 {
					extraInfo = fmt.Sprintf("\n\n%s", strings.Join(selectedModel.ExtraInfo, " "))
				}
				info = fmt.Sprintf(
					"%s\n\n%s%s\n\n%s\n\n%s",
					titleStyle.Render(fmt.Sprintf(" %s ", selectedModel.Name)),
					lipgloss.NewStyle().Foreground(dimTextColor).Render(strings.TrimSpace(selectedModel.Updated)),
					extraInfo,
					wordwrap.String(selectedModel.Desc, m.width-m.installableList.Width()-8),
					lipgloss.NewStyle().Foreground(dimTextColor).Render(
						fmt.Sprintf(
							"%s Pulls • %s Tags",
							selectedModel.Pulls, selectedModel.Tags,
						),
					),
				)
			}
		default:
			if selectedItem != nil {
				selectedModel := list.SelectedItem().(InstalledOllamaModel)
				info = fmt.Sprintf(
					"%s\n\n%s\n\n%s\n\n%s",
					titleStyle.
						AlignHorizontal(lipgloss.Center).
						Render(fmt.Sprintf(" %s ", selectedModel.Name)),
					strings.Join([]string{
						titleStyle.
							Copy().
							Background(lipgloss.Color("242")).
							Render(fmt.Sprintf(" %s ", selectedModel.Details.Format)),
						titleStyle.
							Copy().
							Background(lipgloss.Color("242")).
							Render(fmt.Sprintf(" %s ", selectedModel.Details.QuantizationLevel)),
					}, " "),
					wordwrap.String(selectedModel.Digest, m.width-m.installableList.Width()),
					lipgloss.NewStyle().Foreground(dimTextColor).Render(
						selectedModel.Description(),
					),
				)
			}

		}

		frames = append(frames, layoutStyle.Copy().
			Width(m.width-list.Width()).
			Height(m.height-v).
			AlignHorizontal(lipgloss.Center).
			Padding(0, 2).
			BorderForeground(lipgloss.Color("#209fb5")).
			Render(info),
		)
	}

	activeTabContent = lipgloss.JoinHorizontal(
		lipgloss.Top,
		frames...,
	)
	if m.helpVisible {
		activeTabContent = PlaceOverlay(
			m.width/10,
			5*m.height/100,
			layoutStyle.
				Copy().
				Width(8*m.width/10).
				Height(90*m.height/100).
				AlignHorizontal(lipgloss.Center).
				BorderForeground(lipgloss.Color("#209fb5")).
				Render(
					titleStyle.Render(" Help Menu ")+
						"\n\n"+
						m.help.View(Keys)+
						"\n\n"+
						fmt.Sprintf("Press %s to close this menu", titleStyle.Render(" ? ")),
				),
			activeTabContent,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		row,
		activeTabContent,
	)
}
