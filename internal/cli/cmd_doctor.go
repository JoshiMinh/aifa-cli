package cli

import (
	"fmt"
	"os"
	"strings"

	"aifiler/internal/models"
)

func (a *App) runDoctor() int {
	headerStyle.Println("aifiler diagnostics")

	if cwd, err := os.Getwd(); err == nil {
		fmt.Printf("cwd: %s\n", cwd)
	}
	if exePath, err := os.Executable(); err == nil {
		fmt.Printf("executable: %s\n", exePath)
	}

	configured := strings.TrimSpace(os.Getenv(models.RegistryPathEnvVar))
	if configured == "" {
		fmt.Printf("%s: (not set)\n", models.RegistryPathEnvVar)
	} else {
		fmt.Printf("%s: %s\n", models.RegistryPathEnvVar, configured)
	}

	resolved, err := models.ResolveRegistryPath(models.DefaultRegistryPath)
	if err != nil {
		errorStyle.Printf("registry: unresolved (%v)\n", err)
		return 1
	}

	successStyle.Printf("registry: %s\n", resolved)
	return 0
}
