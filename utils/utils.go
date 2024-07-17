package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/gaurav-gosain/ollamanager/tabs"
)

type OllamanagerResult struct {
	Err          error
	Action       tabs.Tab
	ManageAction tabs.ManageAction
	ModelName    string
	IsMultiModal bool
}

func PrintError(err error) {
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
	}
}

func PrintActionResult(result OllamanagerResult, err error) error {
	if err != nil {
		PrintError(err)
		return err
	}

	if result.ManageAction == tabs.CHAT {
		return nil
	}

	Padding := lipgloss.NewStyle().Padding(1, 2)
	SuccessHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F1F1")).
		Background(lipgloss.Color("#8839ef")).
		Bold(true).
		Padding(0, 1)

	actionStr := string(result.Action)

	if result.Action != tabs.MONITOR {
		if result.Action == tabs.MANAGE {
			actionStr = string(result.ManageAction)
		}
		fmt.Println(
			Padding.Render(
				fmt.Sprintf(
					"Performed action %s on model %s successfully!",
					SuccessHeader.Render(actionStr),
					SuccessHeader.Render(result.ModelName),
				),
			),
		)
	}

	return nil
}

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ClearTerminal() {
	switch runtime.GOOS {
	case "darwin":
		runCmd("clear")
	case "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}
