package tui

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

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
