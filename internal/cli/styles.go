package cli

import "github.com/fatih/color"

var (
	headerStyle  = color.New(color.FgHiCyan, color.Bold)
	successStyle = color.New(color.FgHiGreen)
	warnStyle    = color.New(color.FgHiYellow)
	errorStyle   = color.New(color.FgHiRed, color.Bold)
	mutedStyle   = color.New(color.FgHiBlack)
)
