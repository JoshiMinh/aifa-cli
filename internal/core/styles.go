package core

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

var (
	HeaderStyle  = color.New(color.FgCyan, color.Bold)
	SuccessStyle = color.New(color.FgGreen)
	WarnStyle    = color.New(color.FgYellow)
	ErrorStyle   = color.New(color.FgRed, color.Bold)
	MutedStyle   = color.New(color.FgHiBlack)
	PathStyle    = color.New(color.FgHiBlue, color.Italic)
)

const (
	SparkleIcon = "✨"
	FolderIcon  = "📁"
	FileIcon    = "📄"
	InfoIcon    = "ℹ️"
	WarnIcon    = "⚠️"
	ErrorIcon   = "❌"
	SuccessIcon = "✅"
	RenameIcon  = "➡️"
	DeleteIcon  = "🗑️"
	EditIcon    = "✍️"
	CommandIcon = "💻"
)

type Thinking struct {
	pb *progressbar.ProgressBar
}

func StartThinking(msg string) *Thinking {
	pb := progressbar.Default(-1, msg)
	return &Thinking{pb: pb}
}

func (t *Thinking) Stop(finalMsg string) {
	t.pb.Describe(finalMsg)
	t.pb.Finish()
	fmt.Println()
}
