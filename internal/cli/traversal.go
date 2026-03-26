package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func buildWorkspaceContext(maxDepth int, showAll bool) string {
	cwd, _ := os.Getwd()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Current Directory: %s\n\n", cwd))
	sb.WriteString("File Tree:\n")

	fileCount := 0
	err := filepath.WalkDir(cwd, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == cwd {
			return nil
		}

		// Skip hidden files/dirs (except .aifiler if needed, but usually skip all)
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(cwd, path)
		depth := strings.Count(rel, string(os.PathSeparator)) + 1

		if maxDepth > 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		indent := strings.Repeat("  ", depth-1)
		icon := "📄"
		if d.IsDir() {
			icon = "📁"
		}

		// Limit context size to prevent token blowup unless -all is set
		fileCount++
		if !showAll && fileCount > 100 {
			if fileCount == 101 {
				sb.WriteString("... (truncated, use -all to see more)\n")
			}
			return nil
		}

		sb.WriteString(fmt.Sprintf("%s%s %s\n", indent, icon, rel))
		return nil
	})

	if err != nil {
		sb.WriteString(fmt.Sprintf("Error scanning directory: %v\n", err))
	}

	return sb.String()
}
