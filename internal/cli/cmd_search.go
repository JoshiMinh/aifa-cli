package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func (a *App) runSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	nameFlag := fs.String("name", "", "Search by file name (substring)")
	extFlag := fs.String("ext", "", "Search by file extension (e.g. .go or go)")
	contentFlag := fs.String("content", "", "Search by file content (substring)")

	if err := fs.Parse(args); err != nil {
		errorStyle.Printf("%s failed to parse search arguments\n", errorIcon)
		return 2
	}

	name := strings.ToLower(*nameFlag)
	ext := strings.ToLower(*extFlag)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	content := *contentFlag

	if name == "" && ext == "" && content == "" {
		errorStyle.Printf("%s Usage: aifiler search -name \"foo\" -ext \".go\" -content \"func main\"\n", errorIcon)
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		errorStyle.Printf("%s failed to get current directory: %v\n", errorIcon, err)
		return 1
	}

	headerStyle.Println("Searching for files...")
	
	var wg sync.WaitGroup
	results := make(chan string, 100)

	// A simple worker pool for reading contents
	type job struct {
		path string
		rel  string
	}
	jobs := make(chan job, 100)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				data, err := os.ReadFile(j.path)
				if err == nil {
					if strings.Contains(string(data), content) {
						results <- j.rel
					}
				}
			}
		}()
	}

	go func() {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				if d != nil && d.IsDir() {
					name := d.Name()
					if name == ".git" || name == "node_modules" {
						return filepath.SkipDir
					}
				}
				return nil
			}

			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}

			if name != "" && !strings.Contains(strings.ToLower(d.Name()), name) {
				return nil
			}

			if ext != "" && strings.ToLower(filepath.Ext(d.Name())) != ext {
				return nil
			}

			if content != "" {
				info, err := d.Info()
				if err == nil && info.Size() < 5*1024*1024 { // skip files > 5MB
					jobs <- job{path: path, rel: rel}
				}
			} else {
				results <- rel
			}
			return nil
		})
		close(jobs)
		wg.Wait()
		close(results)
	}()

	found := 0
	for res := range results {
		fmt.Printf("%s %s\n", fileIcon, pathStyle.Sprint(res))
		found++
	}

	if found == 0 {
		warnStyle.Printf("%s No matching files found.\n", warnIcon)
	} else {
		successStyle.Printf("\n%s Found %d matching files.\n", successIcon, found)
	}

	return 0
}
