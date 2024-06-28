package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/gaurav-gosain/ollamanager/manager"
	"github.com/gaurav-gosain/ollamanager/tabs"
	"github.com/gaurav-gosain/ollamanager/utils"
)

func main() {
	for {
		selectedTabs := []tabs.Tab{
			tabs.INSTALL,
			tabs.INSTALLED,
			tabs.RUNNING,
		}
		approvedActions := []tabs.InstalledAction{
			tabs.UPDATE,
			tabs.DELETE,

			// INFO: Other actions
			// tabs.CHAT,
		}

		action, manageAction, selectedModel, err := manager.Run(selectedTabs, approvedActions)

		err = utils.PrintActionResult(
			action,
			manageAction,
			selectedModel,
			err,
		)
		if err != nil {
			break
		}

		var confirm bool

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Would you like to continue?").
					Value(&confirm),
			),
		).WithProgramOptions(tea.WithAltScreen())

		err = form.Run()
		if err != nil {
			return
		}

		if !confirm {
			break
		}
	}
}
