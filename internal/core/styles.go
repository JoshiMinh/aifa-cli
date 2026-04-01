package core

import (
	"fmt"
	"time"

	"github.com/fatih/color"
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
	SparkleIcon = "◆"
	FolderIcon  = "▸"
	FileIcon    = "▫"
	InfoIcon    = "ℹ"
	WarnIcon    = "⚠"
	ErrorIcon   = "✖"
	SuccessIcon = "✔"
	RenameIcon  = "→"
	DeleteIcon  = "DEL"
	EditIcon    = "✎"
	CommandIcon = "❯"
)

type Thinking struct {
	stop chan bool
	done chan bool
}

func StartThinking(msg string) *Thinking {
	t := &Thinking{
		stop: make(chan bool),
		done: make(chan bool),
	}
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		colors := []*color.Color{
			color.New(color.FgCyan),
			color.New(color.FgHiBlue),
			color.New(color.FgMagenta),
		}

		fmt.Print("\033[?25l") // Hide cursor
		i := 0
		for {
			select {
			case <-t.stop:
				fmt.Print("\033[?25h") // Show cursor
				close(t.done)
				return
			default:
				char := colors[i%len(colors)].Sprint(frames[i%len(frames)])
				fmt.Printf("\r%s %s", char, msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return t
}

func (t *Thinking) Stop(finalMsg string) {
	t.stop <- true
	<-t.done
	fmt.Printf("\r\033[K%s %s\n", SuccessIcon, SuccessStyle.Sprint(finalMsg))
}
