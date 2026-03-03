package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("#00D26A")
	colorRed    = lipgloss.Color("#FF4D4F")
	colorYellow = lipgloss.Color("#FAAD14")
	colorBlue   = lipgloss.Color("#4096FF")
	colorGray   = lipgloss.Color("#8C8C8C")
	colorWhite  = lipgloss.Color("#F0F0F0")
	colorAccent = lipgloss.Color("#722ED1")

	StyleSuccess = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	StyleError   = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	StyleWarn    = lipgloss.NewStyle().Foreground(colorYellow)
	StyleInfo    = lipgloss.NewStyle().Foreground(colorBlue)
	StyleSkip    = lipgloss.NewStyle().Foreground(colorGray)
	StyleBold    = lipgloss.NewStyle().Bold(true)
	StyleDim     = lipgloss.NewStyle().Foreground(colorGray)
	StyleAccent  = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	StyleStepHeader = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorAccent).
			Padding(0, 1).
			Bold(true)

	StyleSuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(0, 2)
)
