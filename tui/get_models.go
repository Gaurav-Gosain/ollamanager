package tui

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/lipgloss"
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
	// Match any sequence of whitespace characters or newline characters
	regex := regexp.MustCompile(`\s+`)
	// Replace matched sequences with a single space
	cleaned := regex.ReplaceAllString(input, " ")
	return strings.TrimSpace(cleaned)
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

	doc, err := goquery.
		NewDocumentFromReader(
			strings.NewReader(
				string(body),
			),
		)
	if err != nil {
		return nil, err
	}

	var models []OllamaModel

	doc.Find("#repo > ul > li").Each(func(i int, s *goquery.Selection) {
		var model OllamaModel

		root := s.Find("a > div")

		model.Name = removeWhitespace(
			root.
				Find("h2 > div > span").
				Text(),
		)
		model.Desc = removeWhitespace(
			root.
				Find("p").
				First().
				Text(),
		)

		infoP := s.Find("p > span")

		model.Pulls = removeWhitespace(
			infoP.
				First().
				Find("span").
				First().
				Contents().
				Not("*").
				Text(),
		)
		model.Tags = removeWhitespace(
			infoP.
				First().
				Next().
				First().
				Find("span").
				First().
				Contents().
				Not("*").
				Text(),
		)
		model.Updated = removeWhitespace(
			infoP.
				Last().
				First().
				Find("span").
				Last().
				Contents().
				Not("*").
				Text(),
		)

		tagStyle := titleStyle.
			Background(lipgloss.Color("242")).Render

		tagBorder := titleStyle.Foreground(lipgloss.Color("242")).UnsetBackground().Render

		root.Find("div > span").Each(func(i int, span *goquery.Selection) {
			model.ExtraInfo = append(
				model.ExtraInfo,
				tagBorder(LEFT_HALF_CIRCLE)+
					tagStyle(
						fmt.Sprintf(
							" %s ",
							removeWhitespace(
								span.Text(),
							),
						),
					)+tagBorder(RIGHT_HALF_CIRCLE),
			)
		})

		models = append(models, model)
	})

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
