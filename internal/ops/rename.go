package ops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aifa/internal/llm"
)

type RenamePlanInput struct {
	TargetPath string
	Recursive  bool
	MaxFiles   int
	DryRun     bool
}

type RenamePlanner struct {
	llm llm.Client
}

func NewRenamePlanner(client llm.Client) *RenamePlanner {
	return &RenamePlanner{llm: client}
}

type RenameAction struct {
	From    string
	To      string
	Reason  string
	Applied bool
	Error   string
}

type RenameResult struct {
	Actions []RenameAction
	DryRun  bool
}

func (r RenameResult) Print() {
	fmt.Println("Rename plan")
	for _, action := range r.Actions {
		status := "preview"
		if action.Applied {
			status = "applied"
		}
		if action.Error != "" {
			status = "error"
		}
		fmt.Printf("- [%s] %s -> %s", status, action.From, action.To)
		if action.Reason != "" {
			fmt.Printf(" (%s)", action.Reason)
		}
		if action.Error != "" {
			fmt.Printf(" [error: %s]", action.Error)
		}
		fmt.Println()
	}
	if len(r.Actions) == 0 {
		fmt.Println("- no rename suggestions")
	}
}

func (p *RenamePlanner) Plan(ctx context.Context, in RenamePlanInput) (RenameResult, error) {
	info, err := os.Stat(in.TargetPath)
	if err != nil {
		return RenameResult{}, err
	}

	files, err := collectFiles(in.TargetPath, info.IsDir(), in.Recursive, in.MaxFiles)
	if err != nil {
		return RenameResult{}, err
	}

	result := RenameResult{DryRun: in.DryRun}
	for _, file := range files {
		dir := filepath.Dir(file)
		base := filepath.Base(file)
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)

		suggested, err := p.llm.SuggestName(ctx, stem, "file rename")
		if err != nil {
			result.Actions = append(result.Actions, RenameAction{From: file, To: file, Reason: "llm suggestion failed", Error: err.Error()})
			continue
		}

		suggested = sanitizeFilename(suggested)
		if suggested == "" || suggested == stem {
			continue
		}
		target := filepath.Join(dir, suggested+ext)
		action := RenameAction{From: file, To: target, Reason: "semantic rename"}

		if !in.DryRun {
			if _, err := os.Stat(target); err == nil {
				action.Error = "target already exists"
			} else if err := os.Rename(file, target); err != nil {
				action.Error = err.Error()
			} else {
				action.Applied = true
			}
		}
		result.Actions = append(result.Actions, action)
	}
	return result, nil
}

func collectFiles(path string, isDir bool, recursive bool, maxFiles int) ([]string, error) {
	if maxFiles <= 0 {
		maxFiles = 100
	}
	if !isDir {
		return []string{path}, nil
	}

	files := make([]string, 0, maxFiles)
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if len(files) >= maxFiles {
			return filepath.SkipDir
		}
		if d.IsDir() {
			if p != path && !recursive {
				return filepath.SkipDir
			}
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ":", "-")
	name = replacer.Replace(name)
	name = strings.Join(strings.Fields(name), "-")
	name = strings.Trim(name, ".-")
	return name
}
