package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

func (a *App) runLs(args []string) int {
	dir, err := os.Getwd()
	if err != nil {
		errorStyle.Printf("%s failed to get current directory: %v\n  %s Tip: Check if the directory still exists or if you have permissions.\n", errorIcon, err, infoIcon)
		return 1
	}

	maxDepth := 1
	if len(args) > 0 {
		switch args[0] {
		case "-r":
			maxDepth = 2
		case "-ra":
			maxDepth = 0 // infinite depth
		}
	}

	type EntryInfo struct {
		Name    string
		IsDir   bool
		Size    int64
		ModTime time.Time
	}
	var entries []EntryInfo

	maxLength := 4 // Minimum padding for "Name"
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == dir {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(filepath.Clean(rel))

		depth := strings.Count(rel, "/") + 1
		if maxDepth > 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		length := len(rel)
		if d.IsDir() {
			length++ // trailing slash
		}
		if length > maxLength {
			maxLength = length
		}

		entries = append(entries, EntryInfo{
			Name:    rel,
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		errorStyle.Printf("%s failed to read directory contents: %v\n", errorIcon, err)
		return 1
	}

	if len(entries) == 0 {
		fmt.Printf("%s Directory is empty.\n", infoIcon)
		return 0
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir && !entries[j].IsDir {
			return true
		}
		if !entries[i].IsDir && entries[j].IsDir {
			return false
		}
		return entries[i].Name < entries[j].Name
	})

	paddingStr := fmt.Sprintf("%%-%ds %%-15s %%s\n", maxLength+2)
	headerStyle.Printf(paddingStr, "Name", "Size", "Modified")
	fmt.Println(strings.Repeat("-", maxLength+40))

	for _, entry := range entries {
		name := entry.Name
		size := formatSize(entry.Size)
		if entry.IsDir {
			size = "-"
		}
		modTime := entry.ModTime.Format(time.DateTime)

		var icon string
		var nameFormatted string

		if entry.IsDir {
			icon = dirIcon
			nameFormatted = color.New(color.FgHiCyan).Sprint(name) + "/"
		} else {
			icon = fileIcon
			nameFormatted = pathStyle.Sprint(name)
		}

		plainLength := 2 + len(name)
		if entry.IsDir {
			plainLength++
		}
		padding := (maxLength + 2) - plainLength
		if padding < 1 {
			padding = 1
		}

		fmt.Printf("%s %s%*s %-15s %s\n", icon, nameFormatted, padding, "", color.New(color.FgYellow).Sprint(size), color.New(color.FgHiBlack).Sprint(modTime))
	}

	return 0
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
