package tui

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

type OllamaModel struct {
	Name      string
	Desc      string
	Pulls     string
	Tags      string
	Updated   string
	ExtraInfo []string
}

func removeWhitespace(input string) string {
	input = strings.TrimSpace(input)
	// Match any sequence of whitespace characters or newline characters
	regex := regexp.MustCompile(`\s+`)
	// Replace matched sequences with a single space
	cleaned := regex.ReplaceAllString(input, " ")
	return cleaned
}

// Function to find the first child node with a specific tag name
func findFirstNodeByTag(node *html.Node, tagName string) *html.Node {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			return c
		}
	}
	return nil
}

func findNthNodeByTag(node *html.Node, tagName string, n int) *html.Node {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			if n == 0 {
				return c
			} else {
				n -= 1
			}
		}
	}
	return nil
}

func findLastNodeByTag(node *html.Node, tagName string) *html.Node {
	for c := node.LastChild; c != nil; c = c.PrevSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			return c
		}
	}
	return nil
}

// Function to find all child nodes with a specific tag name
func findAllNodesByTag(node *html.Node, tagName string) []*html.Node {
	var nodes []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			nodes = append(nodes, c)
		}
	}
	return nodes
}

func getExtraInfo(node *html.Node) []string {
	infoTags := []string{}

	for _, infoTag := range findAllNodesByTag(node, "span") {
		infoTags = append(
			infoTags,
			titleStyle.
				Copy().
				Background(lipgloss.Color("242")).
				Render(fmt.Sprintf(" %s ", infoTag.FirstChild.Data)),
		)
	}

	return infoTags
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
				var desc, pulls, tags, updated string
				var extraInfo []string
				fmt.Println(n.LastChild.Data)
				descTag := n.LastChild.Data
				desc = removeWhitespace(descTag)
				divTag := findFirstNodeByTag(n, "div")
				if divTag != nil {
					pTag := findFirstNodeByTag(divTag, "p")
					if pTag != nil {
						desc = pTag.FirstChild.Data
					}

					extraInfoTag := findFirstNodeByTag(divTag, "div")

					if extraInfoTag != nil {
						extraInfo = getExtraInfo(extraInfoTag)
					}

					infoTag := findLastNodeByTag(divTag, "p")
					pulls = removeWhitespace(findNthNodeByTag(infoTag, "span", 0).
						FirstChild.
						NextSibling.
						NextSibling.
						Data,
					)
					tags = removeWhitespace(findNthNodeByTag(infoTag, "span", 1).
						FirstChild.
						NextSibling.
						NextSibling.
						Data,
					)
					updated = removeWhitespace(findNthNodeByTag(infoTag, "span", 2).
						LastChild.
						Data,
					)
				}
				models = append(models, OllamaModel{Name: name, Desc: desc, Pulls: pulls, Tags: tags, Updated: updated, ExtraInfo: extraInfo})
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
