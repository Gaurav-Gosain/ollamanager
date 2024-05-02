package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func PrintError(err error, exitOnErr bool) {
	ErrPadding := lipgloss.NewStyle().Padding(1, 2)
	ErrorHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F1F1")).
		Background(lipgloss.Color("#FF5F87")).
		Bold(true).
		Padding(0, 1).
		SetString("ERROR")

	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			ErrPadding.Render(
				fmt.Sprintf(
					"\n%s %s",
					ErrorHeader.String(),
					err.Error(),
				),
			),
		)
		if exitOnErr {
			os.Exit(1)
		}
	}
}

func PrintInstallResult(modelName, tag string, err error) {
	if err != nil {
		PrintError(err, true)
		return
	}

	Padding := lipgloss.NewStyle().Padding(1, 2)
	SuccessHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F1F1")).
		Background(lipgloss.Color("#8839ef")).
		Bold(true).
		Padding(0, 1)

	fmt.Fprintln(
		os.Stdout,
		Padding.Render(
			fmt.Sprintf(
				"Downloaded model %s with tag %s successfully!",
				SuccessHeader.SetString(modelName).String(),
				SuccessHeader.SetString(tag).String(),
			),
		),
	)
}
