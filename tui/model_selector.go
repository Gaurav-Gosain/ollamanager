package tui

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	humanize "github.com/dustin/go-humanize"
	"github.com/gaurav-gosain/ollamanager/tabs"
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
	titleBorder = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).Render

	tagBorder = titleStyle.UnsetBackground().Foreground(lipgloss.Color("242")).Render
	tagStyle  = titleStyle.Background(lipgloss.Color("242")).Render

	descStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4"))

	// Page.

	docStyle          = lipgloss.NewStyle().Margin(1)
	helpStyle         = list.DefaultStyles(lipgloss.HasDarkBackground()).HelpStyle
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	dimTextColor      = lipgloss.AdaptiveColor{Light: "250", Dark: "250"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
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
	runningList              list.Model
	help                     help.Model
	SelectedInstallableModel OllamaModel
	SelectedRunningModel     RunningOllamaModel
	SelectedInstalledModel   InstalledOllamaModel
	Action                   tabs.Tab
	ManageAction             tabs.ManageAction
	Tabs                     []tabs.Tab
	ApprovedActions          []tabs.ManageAction
	width                    int
	height                   int
	ActiveTab                int
	infoVisible              bool
	helpVisible              bool
}

func (m ModelSelector) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *ModelSelector) SetSelectedModel(installAction, manageAction, monitorAction bool) {
	if installAction {
		m.Action = tabs.INSTALL
		m.SelectedInstallableModel = m.installableList.SelectedItem().(OllamaModel)
	} else if monitorAction {
		m.Action = tabs.MONITOR
		m.SelectedRunningModel = m.runningList.SelectedItem().(RunningOllamaModel)
	} else if manageAction {
		m.Action = tabs.MANAGE
		m.SelectedInstalledModel = m.installedList.SelectedItem().(InstalledOllamaModel)
	}
}

func (m ModelSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	installAction := m.Tabs[m.ActiveTab] == tabs.INSTALL
	manageAction := m.Tabs[m.ActiveTab] == tabs.MANAGE
	monitorAction := m.Tabs[m.ActiveTab] == tabs.MONITOR

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
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "?":
			m.helpVisible = !m.helpVisible
			return m, nil
		case "n", "tab":
			m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			m.Action = tabs.Tab(m.Tabs[m.ActiveTab])
			return m, nil
		case "p", "shift+tab":
			m.ActiveTab = max(m.ActiveTab-1, 0)
			m.Action = tabs.Tab(m.Tabs[m.ActiveTab])
			return m, nil
		case "u":
			// if on manage tab, select the update `ManageAction` (if it is in the list of approved actions)
			if manageAction && slices.Contains(m.ApprovedActions, tabs.UPDATE) {
				m.ManageAction = tabs.UPDATE
				m.SetSelectedModel(installAction, manageAction, monitorAction)
				return m, tea.Quit
			}
		case "d":
			// if on manage tab, select the delete `ManageAction` (if it is in the list of approved actions)
			if manageAction && slices.Contains(m.ApprovedActions, tabs.DELETE) {
				m.ManageAction = tabs.DELETE
				m.SetSelectedModel(installAction, manageAction, monitorAction)
				return m, tea.Quit
			}
		case "c":
			// if on manage tab, select the chat `ManageAction` (if it is in the list of approved actions)
			if manageAction && slices.Contains(m.ApprovedActions, tabs.CHAT) {
				m.ManageAction = tabs.CHAT
				m.SetSelectedModel(installAction, manageAction, monitorAction)
				return m, tea.Quit
			}
		case "enter":
			m.SetSelectedModel(installAction, manageAction, monitorAction)
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
		if slices.Contains(m.Tabs, tabs.INSTALL) {
			m.installableList.SetSize(listWidth, m.height-v)
		}
		if slices.Contains(m.Tabs, tabs.MONITOR) {
			m.runningList.SetSize(listWidth, m.height-v)
		}
		if slices.Contains(m.Tabs, tabs.MANAGE) {
			m.installedList.SetSize(listWidth, m.height-v)
		}
	}

	var cmd tea.Cmd

	if !m.helpVisible {
		if installAction {
			m.installableList, cmd = m.installableList.Update(msg)
		} else if monitorAction {
			m.runningList, cmd = m.runningList.Update(msg)
		} else {
			m.installedList, cmd = m.installedList.Update(msg)
		}
	}
	return m, cmd
}

func (m ModelSelector) View() string {
	var renderedTabs []string
	var row string
	var v int

	numTabs := len(m.Tabs)

	if numTabs > 1 {
		tabWidth := int(math.Floor(float64(m.width) / float64(numTabs)))
		for i, tab := range m.Tabs {
			t := string(tab)
			var style lipgloss.Style
			isFirst, isActive := i == 0, i == m.ActiveTab
			w := tabWidth
			if isFirst {
				w = m.width - 1 - ((numTabs - 1) * tabWidth) - (numTabs - 1)
			}
			if isActive {
				style = activeTabStyle.Width(w)
			} else {
				style = inactiveTabStyle.Foreground(dimTextColor).Width(w)
			}

			// If the tab is too long, truncate it
			// TODO: This is a hack, we should probably do something better
			if len(t) > w-2 {
				if w > 2 {
					t = t[:w-2]
				} else {
					t = t[:1]
				}
			}

			renderedTabs = append(renderedTabs, style.Render(t))
		}

		row = lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
		v = activeTabStyle.GetVerticalFrameSize()
	}

	activeTabContent := ""

	var list list.Model

	switch tabs.Tab(m.Tabs[m.ActiveTab]) {
	case tabs.INSTALL:
		list = m.installableList
	case tabs.MONITOR:
		list = m.runningList
	case tabs.MANAGE:
		list = m.installedList
	default:
		list = m.installedList
	}

	frames := []string{
		layoutStyle.
			Padding(0, 2).
			Width(list.Width()).
			Height(m.height - v).
			BorderForeground(lipgloss.Color("69")).
			Render(list.View()),
	}

	selectedItem := list.SelectedItem()

	if m.infoVisible {

		info := fmt.Sprintf("%s not found", titleBorder(LEFT_HALF_CIRCLE)+
			titleStyle.Render(fmt.Sprintf(" %s ", list.FilterValue()))+
			titleBorder(RIGHT_HALF_CIRCLE))

		switch tabs.Tab(m.Tabs[m.ActiveTab]) {
		case tabs.INSTALL:
			if selectedItem != nil {
				selectedModel := selectedItem.(OllamaModel)
				extraInfo := ""
				if len(selectedModel.ExtraInfo) > 0 {
					extraInfo = "\n\n"
					for _, i := range selectedModel.ExtraInfo {
						if lipgloss.Width(extraInfo+i) >= m.width-list.Width()-8 {
							extraInfo += "\n\n"
						}
						extraInfo += " " + i
					}
				}
				info = fmt.Sprintf(
					"%s\n\n%s%s\n\n%s\n\n%s",
					titleBorder(LEFT_HALF_CIRCLE)+
						titleStyle.Render(fmt.Sprintf(" %s ", selectedModel.Name))+
						titleBorder(RIGHT_HALF_CIRCLE),
					lipgloss.NewStyle().Foreground(dimTextColor).Render(strings.TrimSpace(selectedModel.Updated)),
					extraInfo,
					wordwrap.String(selectedModel.Desc, m.width-list.Width()-8),
					lipgloss.NewStyle().Foreground(dimTextColor).Render(
						fmt.Sprintf(
							"%s Pulls • %s Tags",
							selectedModel.Pulls, selectedModel.Tags,
						),
					),
				)
			}
		case tabs.MONITOR:
			if selectedItem != nil {
				selectedModel := selectedItem.(RunningOllamaModel)
				totalSize := float64(selectedModel.Size)
				gpu := float64(selectedModel.SizeVRAM)
				cpu := totalSize - gpu

				gpuPercentage := fmt.Sprintf("%.2f", gpu*100/totalSize)
				cpuPercentage := fmt.Sprintf("%.2f", cpu*100/totalSize)
				info = fmt.Sprintf(
					"%s\n\n%s\n\n%s\n\n%s",
					titleBorder(LEFT_HALF_CIRCLE)+
						titleStyle.Render(fmt.Sprintf(" %s ", selectedModel.Name))+
						titleBorder(RIGHT_HALF_CIRCLE),
					strings.Join([]string{
						tagBorder(LEFT_HALF_CIRCLE) +
							tagStyle(fmt.Sprintf(" %s ", selectedModel.Details.Format)) +
							tagBorder(RIGHT_HALF_CIRCLE),
						tagBorder(LEFT_HALF_CIRCLE) +
							tagStyle(fmt.Sprintf(" %s ", selectedModel.Details.QuantizationLevel)) +
							tagBorder(RIGHT_HALF_CIRCLE),
					}, " "),
					titleBorder(LEFT_HALF_CIRCLE)+
						titleStyle.
							Render(
								fmt.Sprintf(
									" Expires in %s ",
									humanize.Time(selectedModel.ExpiresAt),
								),
							)+
						titleBorder(RIGHT_HALF_CIRCLE),
					fmt.Sprintf(
						"Total Size %s | GPU %s%% | CPU %s%%",
						humanize.Bytes(uint64(selectedModel.Size)),
						gpuPercentage,
						cpuPercentage,
					),
				)
			}
		default:
			if selectedItem != nil {
				selectedModel := selectedItem.(InstalledOllamaModel)

				isMultiModal := ""
				if len(selectedModel.Details.Families) > 1 {
					isMultiModal = titleBorder(LEFT_HALF_CIRCLE) + titleStyle.
						AlignHorizontal(lipgloss.Center).
						Render(fmt.Sprintf("  %s ", "Vision")) +
						titleBorder(RIGHT_HALF_CIRCLE)
				}

				info = fmt.Sprintf(
					"%s\n\n%s\n\n%s\n\n%s\n\n%s",
					titleBorder(LEFT_HALF_CIRCLE)+
						titleStyle.Render(fmt.Sprintf(" %s ", selectedModel.Name))+
						titleBorder(RIGHT_HALF_CIRCLE),
					strings.Join([]string{
						tagBorder(LEFT_HALF_CIRCLE) +
							tagStyle(fmt.Sprintf(" %s ", selectedModel.Details.Format)) +
							tagBorder(RIGHT_HALF_CIRCLE),
						tagBorder(LEFT_HALF_CIRCLE) +
							tagStyle(fmt.Sprintf(" %s ", selectedModel.Details.QuantizationLevel)) +
							tagBorder(RIGHT_HALF_CIRCLE),
					}, " "),
					wordwrap.String(selectedModel.Digest, m.width-list.Width()),
					lipgloss.NewStyle().Foreground(dimTextColor).Render(
						selectedModel.Description(),
					),
					isMultiModal,
				)
			}

		}

		frames = append(frames, layoutStyle.
			Width(m.width-list.Width()-1).
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

		defaultKeys := Keys.DefaultFullHelpKeys()

		if len(m.Tabs) == 1 {
			defaultKeys = Keys.DefaultFullHelpKeysSingleTab()
		}

		if m.Tabs[m.ActiveTab] == tabs.MANAGE {
			keyMap := defaultKeys
			for _, action := range m.ApprovedActions {

				keyBind := string(strings.ToLower(string(action))[0])

				keyMap[0] = append(keyMap[0], key.NewBinding(
					key.WithKeys(keyBind),
					key.WithHelp(keyBind, string(action)),
				))
			}
			Keys.SetFullHelpKeys(keyMap)
		} else {
			Keys.SetFullHelpKeys(defaultKeys)
		}

		activeTabContent = PlaceOverlay(
			m.width/10,
			m.height/10,
			layoutStyle.
				Width(8*m.width/10).
				Height(8*m.height/10).
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
