package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/gaurav-gosain/ollamanager/internal/tui"
	"golang.org/x/net/html"
)

type OllamaModel struct {
	Name        string
	Description string
}

func extractModels(htmlString string) []OllamaModel {
	var models []OllamaModel

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return models
	}

	// Define a function to traverse the HTML tree and extract models
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
				// Find the <p> tag inside the <a> tag
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "p" {
						// Extract the description text
						description := strings.TrimSpace(c.FirstChild.Data)
						models = append(models, OllamaModel{Name: name, Description: description})
						break
					}
				}
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
	models, err := GetAvailableModels()
	if err != nil {
		fmt.Println("Error getting models")
		return "", "", err
	}

	var modelName string
	confirm := false

	options := []huh.Option[string]{}

	for _, model := range models {
		options = append(options, huh.NewOption(model.Name, model.Name))
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

	modelTags, err := GetAvailableTags(modelName)
	if err != nil {
		fmt.Println("Error getting tags")
		return "", "", err
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
