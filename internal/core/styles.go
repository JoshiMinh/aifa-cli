package core

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// UI styles — derived from theme constants in theme.go.
var (
	HeaderStyle  = color.New(PrimaryColor, color.Bold)
	SuccessStyle = color.New(SuccessColor)
	WarnStyle    = color.New(WarnColor)
	ErrorStyle   = color.New(ErrorColor, color.Bold)
	MutedStyle   = color.New(MutedColor)
	PathStyle    = color.New(PathColor, color.Italic)
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
