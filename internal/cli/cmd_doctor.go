package cli

import (
	"fmt"
	"os"
)

func (a *App) runDoctor() int {
	headerStyle.Println("aifiler diagnostics")

	if cwd, err := os.Getwd(); err == nil {
		fmt.Printf("cwd: %s\n", cwd)
	}
	if exePath, err := os.Executable(); err == nil {
		fmt.Printf("executable: %s\n", exePath)
	}

	return 0
}
