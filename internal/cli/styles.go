package cli

import "github.com/fatih/color"

var (
	headerStyle     = color.New(color.FgHiCyan, color.Bold)
	successStyle    = color.New(color.FgHiGreen, color.Bold)
	warnStyle       = color.New(color.FgHiYellow)
	errorStyle      = color.New(color.FgHiRed, color.Bold)
	mutedStyle      = color.New(color.FgHiBlack)
	infoStyle       = color.New(color.FgHiBlue)
	thinkingStyle   = color.New(color.FgHiMagenta, color.Bold)
	pathStyle       = color.New(color.FgHiWhite)
	commandStyle    = color.New(color.FgHiCyan)
	opCreateStyle   = color.New(color.FgHiGreen)
	opUpdateStyle   = color.New(color.FgHiYellow)
	opRenameStyle   = color.New(color.FgHiBlue)
	opCommandStyle  = color.New(color.FgHiMagenta)
	treeBranchStyle = color.New(color.FgHiBlack)
)
