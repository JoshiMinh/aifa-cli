package core

import "github.com/fatih/color"

// Theme — sky-blue palette. All UI styles are derived from these constants.
// To retheme the entire app, only change values here.
const (
	// PrimaryColor is the sky-blue accent used for headers and highlights.
	PrimaryColor = color.FgCyan

	// SuccessColor is used for successful operations.
	SuccessColor = color.FgGreen

	// WarnColor is used for warnings.
	WarnColor = color.FgYellow

	// ErrorColor is used for error messages.
	ErrorColor = color.FgRed

	// MutedColor is used for secondary / de-emphasized text.
	MutedColor = color.FgHiBlack

	// PathColor is used for file and directory paths.
	PathColor = color.FgHiBlue
)
