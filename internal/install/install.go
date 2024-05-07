package install

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/gaurav-gosain/ollamanager/internal/tui"
	"golang.org/x/net/html"
)

type OllamaModel struct {
	Title   string
	Desc    string
	Pulls   string
	Tags    string
	Updated string
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
							desc = strings.TrimSpace(c.FirstChild.Data)
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
				models = append(models, OllamaModel{Title: name, Desc: desc, Pulls: pulls, Tags: tags, Updated: updated})
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

func Run(apiUrl string) (string, string, error) {
	ctx := context.Background()

	var models []OllamaModel
	var err error

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
		return "", "", spinnerErr
	}

	if err != nil {
		return "", "", err
	}

	if len(models) == 0 {
		return "", "", errors.New("couldn't load model names :(")
	}

	var modelName string
	confirm := false

	options := []huh.Option[string]{}

	for _, model := range models {
		options = append(options, huh.NewOption(model.Title, model.Title))
	}

	form := huh.NewForm(
		huh.NewGroup(
			// Ask the user for a base burger and toppings.
			huh.NewSelect[string]().
				Title("Choose your model").
				Options(
					options...,
				).
				Value(&modelName), // store the chosen option in the "modelName" variable
		),
	)

	err = form.Run()
	if err != nil {
		return "", "", err
	}

	var tag string
	confirm = false

	ctx = context.Background()

	var modelTags []string

	loadModelTags := func() {
		modelTags, err = GetAvailableTags(modelName)
		ctx.Done() // signal that model fetching is done
	}

	spinnerErr = spinner.
		New().
		Title("Loading tags for " + modelName + "...").
		Action(loadModelTags).
		Run()

	if spinnerErr != nil {
		return "", "", spinnerErr
	}

	if err != nil {
		return "", "", err
	}

	if len(modelTags) == 0 {
		return "", "", errors.New("couldn't load tags for " + modelName + "  :(")
	}

	options = []huh.Option[string]{}

	for _, modelTag := range modelTags {
		options = append(options, huh.NewOption(modelTag, modelTag))
	}

	form = huh.NewForm(
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
		return "", "", err
	}

	if !confirm {
		return "", "", errors.New("see you")
	}
	downloadingModel := fmt.Sprintf("%s:%s", modelName, tag)

	fmt.Println("Starting to download ", tui.StatusStyle.SetString(downloadingModel).String())

	m := tui.InstallModel{
		Progress: progress.New(progress.WithDefaultGradient()),
		Spinner:  tui.InitSpinner(),
	}
	// Start Bubble Tea
	p := tea.NewProgram(m)

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
	}(apiUrl, downloadingModel, p)

	res, err := p.Run()
	if err != nil {
		fmt.Println("error running program:", err)
		return "", "", err
	}

	return modelName, tag, res.(tui.InstallModel).Err
}
