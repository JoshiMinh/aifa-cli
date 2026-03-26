package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

var (
	headerStyle  = color.New(color.FgCyan, color.Bold)
	successStyle = color.New(color.FgGreen)
	warnStyle    = color.New(color.FgYellow)
	errorStyle   = color.New(color.FgRed, color.Bold)
	mutedStyle   = color.New(color.FgHiBlack)
	pathStyle    = color.New(color.FgHiBlue, color.Italic)
)

const (
	sparkleIcon = "✨"
	folderIcon  = "📁"
	fileIcon    = "📄"
	infoIcon    = "ℹ️"
	warnIcon    = "⚠️"
	errorIcon   = "❌"
	successIcon = "✅"
	renameIcon  = "➡️"
	deleteIcon  = "🗑️"
	editIcon    = "✍️"
	commandIcon = "💻"
)

type thinking struct {
	pb *progressbar.ProgressBar
}

func startThinking(msg string) *thinking {
	pb := progressbar.Default(-1, msg)
	return &thinking{pb: pb}
}

func (t *thinking) stop(finalMsg string) {
	t.pb.Describe(finalMsg)
	t.pb.Finish()
	fmt.Println()
}
