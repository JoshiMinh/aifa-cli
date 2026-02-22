package ops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type MetadataSuggestion struct {
	Path        string
	Title       string
	Description string
	Tags        []string
}

type MetadataResult struct {
	Items []MetadataSuggestion
}

func (r MetadataResult) Print() {
	fmt.Println("Metadata suggestions")
	for _, item := range r.Items {
		fmt.Printf("- %s\n", item.Path)
		fmt.Printf("    title: %s\n", item.Title)
		fmt.Printf("    desc : %s\n", item.Description)
		fmt.Printf("    tags : %v\n", item.Tags)
	}
	if len(r.Items) == 0 {
		fmt.Println("- no metadata suggestions")
	}
}

type MetadataSuggester struct{}

func NewMetadataSuggester() *MetadataSuggester {
	return &MetadataSuggester{}
}

func (s *MetadataSuggester) Suggest(ctx context.Context, targetPath string) (MetadataResult, error) {
	_ = ctx
	info, err := os.Stat(targetPath)
	if err != nil {
		return MetadataResult{}, err
	}

	if !info.IsDir() {
		return MetadataResult{Items: []MetadataSuggestion{buildSuggestion(targetPath, info)}}, nil
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return MetadataResult{}, err
	}
	items := make([]MetadataSuggestion, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(targetPath, entry.Name())
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, buildSuggestion(fullPath, entryInfo))
	}

	return MetadataResult{Items: items}, nil
}

func buildSuggestion(path string, info os.FileInfo) MetadataSuggestion {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	stem := name[:len(name)-len(ext)]
	modified := info.ModTime().Format(time.RFC3339)
	return MetadataSuggestion{
		Path:        path,
		Title:       stem,
		Description: fmt.Sprintf("%s file updated %s", ext, modified),
		Tags:        []string{ext, "aifa", "suggested"},
	}
}
