package ollamanager

import (
	"github.com/gaurav-gosain/ollamanager/internal/install"
	"github.com/gaurav-gosain/ollamanager/internal/utils"
)

func Ollamanager(baseURL string) {
	utils.PrintInstallResult(
		install.Run(
			baseURL,
		),
	)
}
