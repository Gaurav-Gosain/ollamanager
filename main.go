package main

import (
	"flag"
	"fmt"

	"github.com/gaurav-gosain/ollamanager/internal/install"
	"github.com/gaurav-gosain/ollamanager/internal/utils"
)

func main() {
	var baseURL string
	flag.StringVar(&baseURL, "base-url", "http://localhost:11434", "Base URL for the API server")
	flag.Parse()

	utils.PrintInstallResult(
		install.Run(
			fmt.Sprintf("%s/api/pull", baseURL),
		),
	)
}
