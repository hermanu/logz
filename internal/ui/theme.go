package ui

import "github.com/charmbracelet/lipgloss"

// Theme Colors
var (
	Background = lipgloss.Color("#0d0d0e")
	Foreground = lipgloss.Color("#e6e4dd")
	Muted      = lipgloss.Color("#7a7a80")

	Fatal = lipgloss.Color("#e85a4f")
	Error = lipgloss.Color("#e85a4f")
	Warn  = lipgloss.Color("#e0a44c")
	Info  = lipgloss.Color("#7da7c8")
	Debug = lipgloss.Color("#7a7a80")
	Trace = lipgloss.Color("#5a5a60")

	Selection    = lipgloss.Color("#2a2a2f")
	Cursor      = lipgloss.Color("#e85a4f")
	Border      = lipgloss.Color("#1f1f22")
	BorderFocus = lipgloss.Color("#4a4a50")
)

// LevelColor returns color for level
func LevelColor(level string) lipgloss.Color {
	switch level {
	case "FATAL", "fatal":
		return Fatal
	case "ERROR", "error":
		return Error
	case "WARN", "warn", "WARNING":
		return Warn
	case "INFO", "info":
		return Info
	case "DEBUG", "debug":
		return Debug
	case "TRACE", "trace":
		return Trace
	default:
		return Foreground
	}
}

// Layout constants
const (
	PaneSidebar = iota
	PaneList
	PaneDetail

	SidebarWidth = 18
	DetailWidth  = 28
	HeaderHeight = 1
	StatusHeight = 1
)