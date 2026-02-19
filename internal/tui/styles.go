package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorBg       = lipgloss.Color("#0d0d0d")
	colorAccent   = lipgloss.Color("#8b5cf6") // violet — Poe's signature
	colorDim      = lipgloss.Color("#4b4b4b")
	colorText     = lipgloss.Color("#e2e2e2")
	colorPoeText  = lipgloss.Color("#c4b5fd")
	colorUserText = lipgloss.Color("#6ee7b7")

	styleHeader = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	styleInput = lipgloss.NewStyle().
			Foreground(colorText).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	stylePoeMsg  = lipgloss.NewStyle().Foreground(colorPoeText)
	styleUserMsg = lipgloss.NewStyle().Foreground(colorUserText)
)
