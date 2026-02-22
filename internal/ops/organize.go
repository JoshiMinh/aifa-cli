package ops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type OrganizeInput struct {
	TargetPath string
	DryRun     bool
}

type OrganizeAction struct {
	From    string
	To      string
	Applied bool
	Error   string
}

type OrganizeResult struct {
	Actions []OrganizeAction
	DryRun  bool
}

func (r OrganizeResult) Print() {
	fmt.Println("Organize plan")
	for _, action := range r.Actions {
		status := "preview"
		if action.Applied {
			status = "applied"
		}
		if action.Error != "" {
			status = "error"
		}
		fmt.Printf("- [%s] %s -> %s", status, action.From, action.To)
		if action.Error != "" {
			fmt.Printf(" [error: %s]", action.Error)
		}
		fmt.Println()
	}
	if len(r.Actions) == 0 {
		fmt.Println("- no organization suggestions")
	}
}

type Organizer struct{}

func NewOrganizer() *Organizer {
	return &Organizer{}
}

func (o *Organizer) Plan(ctx context.Context, in OrganizeInput) (OrganizeResult, error) {
	_ = ctx
	entries, err := os.ReadDir(in.TargetPath)
	if err != nil {
		return OrganizeResult{}, err
	}

	result := OrganizeResult{DryRun: in.DryRun}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		source := filepath.Join(in.TargetPath, entry.Name())
		category := categorize(entry.Name())
		destinationDir := filepath.Join(in.TargetPath, category)
		destination := filepath.Join(destinationDir, entry.Name())
		action := OrganizeAction{From: source, To: destination}

		if in.DryRun {
			result.Actions = append(result.Actions, action)
			continue
		}

		if err := os.MkdirAll(destinationDir, 0o755); err != nil {
			action.Error = err.Error()
			result.Actions = append(result.Actions, action)
			continue
		}
		if _, err := os.Stat(destination); err == nil {
			action.Error = "target already exists"
			result.Actions = append(result.Actions, action)
			continue
		}
		if err := os.Rename(source, destination); err != nil {
			action.Error = err.Error()
		} else {
			action.Applied = true
		}
		result.Actions = append(result.Actions, action)
	}

	return result, nil
}

func categorize(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "Images"
	case ".mp4", ".mov", ".mkv", ".avi", ".webm":
		return "Videos"
	case ".mp3", ".wav", ".flac", ".m4a":
		return "Audio"
	case ".pdf", ".doc", ".docx", ".txt", ".md":
		return "Documents"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "Archives"
	case ".go", ".js", ".ts", ".py", ".java", ".cs", ".json", ".yaml", ".yml", ".toml":
		return "Code"
	default:
		return "Other"
	}
}
