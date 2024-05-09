package tui

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
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

type OllamaModel struct {
	Name    string
	Desc    string
	Pulls   string
	Tags    string
	Updated string
}

func removeExtraWhitespace(input string) string {
	// Match any sequence of whitespace characters or newline characters
	regex := regexp.MustCompile(`\s+`)
	// Replace matched sequences with a single space
	cleaned := regex.ReplaceAllString(input, " ")
	return cleaned
}

func extractModels(htmlString string) []OllamaModel {
	var models []OllamaModel

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return models
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// Check if the <a> tag has an href attribute starting with "/library/"
			var href, name string
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.HasPrefix(attr.Val, "/library/") {
					href = attr.Val
					name = strings.TrimPrefix(href, "/library/")
					break
				}
			}
			if href != "" {
				// Find the last <p> tag inside the <a> tag
				var desc, pulls, tags, updated string
				found := false
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if found {
						break
					}
					if c.Type == html.ElementNode && c.Data == "p" {
						if strings.TrimSpace(c.FirstChild.Data) != "" {
							desc = removeExtraWhitespace(c.FirstChild.Data)
						} else {
							// Find and extract pulls, tags, and updated numbers from <span> tags
							for span := c.FirstChild; span != nil; span = span.NextSibling {
								if found {
									break
								}
								if span.Type == html.ElementNode && span.Data == "span" {
									for spanContent := span.FirstChild; spanContent != nil; spanContent = spanContent.NextSibling {
										text := strings.TrimSpace(spanContent.Data)
										if text == "" || text == "span" || text == "svg" {
											continue
										}
										// fmt.Println(text)
										if pulls == "" {
											pulls = text
										} else if tags == "" {
											tags = text
										} else if updated == "" {
											updated = text
										} else {
											found = true
											break
										}
									}
								}
							}
						}
					}
				}
				models = append(models, OllamaModel{Name: name, Desc: desc, Pulls: pulls, Tags: tags, Updated: updated})
			}
		}
		// Recursively call the function for child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	// Start traversing the HTML tree
	traverse(doc)

	return models
}

func GetAvailableModels() ([]OllamaModel, error) {
	resp, err := http.Get("https://ollama.com/library")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	models := extractModels(string(body))
	return models, nil
}

func (model OllamaModel) Title() string {
	return model.Name
}

func (model OllamaModel) Description() string {
	return fmt.Sprintf(
		"↓ %s • %s tags • %s",
		model.Pulls, model.Tags, model.Updated,
	)
}
func (model OllamaModel) FilterValue() string { return model.Name }

type ModelSelector struct {
	list          list.Model
	SelectedModel OllamaModel
	Tabs          []string
	width         int
	height        int
	ActiveTab     int
	infoVisible   bool
}

func (m ModelSelector) Init() tea.Cmd {
	return nil
}

func (m ModelSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Tabs[m.ActiveTab] == INSTALL {
			if m.list.FilterState() == list.Filtering {
				break
			}
		}
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "n", "tab":
			m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			return m, nil
		case "p", "shift+tab":
			m.ActiveTab = max(m.ActiveTab-1, 0)
			return m, nil
		case "enter":
			if m.Tabs[m.ActiveTab] == INSTALL {
				m.SelectedModel = m.list.SelectedItem().(OllamaModel)
				return m, tea.Quit
			}
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
		v = activeTabStyle.GetVerticalFrameSize()
		m.list.SetSize(listWidth, m.height-v)
	}

	var cmd tea.Cmd

	if m.Tabs[m.ActiveTab] == INSTALL {
		m.list, cmd = m.list.Update(msg)
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

	frames := []string{
		layoutStyle.Copy().
			Padding(0, 2).
			Width(m.list.Width()).
			Height(m.height - v).
			BorderForeground(lipgloss.Color("69")).
			Render(m.list.View()),
	}
	if m.list.SelectedItem() != nil && m.infoVisible {

		selectedModel := m.list.SelectedItem().(OllamaModel)

		info := fmt.Sprintf(
			"%s\n%s\n\n%s\n\n%s",
			titleStyle.Render(selectedModel.Name),
			helpStyle.Foreground(dimTextColor).Render(strings.TrimSpace(selectedModel.Updated)),
			selectedModel.Desc,
			helpStyle.Foreground(dimTextColor).Render(
				fmt.Sprintf(
					"%s Pulls • %s Tags",
					selectedModel.Pulls, selectedModel.Tags,
				),
			),
		)

		frames = append(frames, layoutStyle.Copy().
			Width(m.width-m.list.Width()).
			Height(m.height-v).
			AlignHorizontal(lipgloss.Center).
			Padding(0, 2).
			BorderForeground(lipgloss.Color("#209fb5")).
			Render(info),
		)
	}

	switch m.Tabs[m.ActiveTab] {
	case INSTALL:
		activeTabContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			frames...,
		)
	case UPDATE:
		activeTabContent = "Update go brr..."
	case DELETE:
		activeTabContent = "Delete go brr..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		row,
		activeTabContent,
	)
}

func ModelPicker(tabs []string) (model OllamaModel, err error) {
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

	m := ModelSelector{list: listModel, Tabs: tabs}
	m.list.Title = "Pick a Model..."

	p := tea.NewProgram(m, tea.WithAltScreen())

	resModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return resModel.(ModelSelector).SelectedModel, nil
}
