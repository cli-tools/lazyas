package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Primary    = lipgloss.Color("#7C3AED") // Purple
	Secondary  = lipgloss.Color("#10B981") // Green
	Accent     = lipgloss.Color("#F59E0B") // Amber
	Danger     = lipgloss.Color("#EF4444") // Red
	MutedColor = lipgloss.Color("#6B7280") // Gray
	Subtle     = lipgloss.Color("#374151") // Dark gray

	// Muted style (for rendering)
	Muted = lipgloss.NewStyle().
		Foreground(MutedColor)

	// Base styles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginBottom(1)

	// List styles
	SelectedItem = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(Primary).
			Padding(0, 1)

	NormalItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	InstalledBadge = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Status indicators
	StatusInstalled = lipgloss.NewStyle().
			Foreground(Secondary).
			SetString("●")

	StatusAvailable = lipgloss.NewStyle().
			Foreground(MutedColor).
			SetString("○")

	StatusModified = lipgloss.NewStyle().
			Foreground(Accent). // Yellow/Amber for modified
			SetString("◉")

	// Help bar
	HelpBar = lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1)

	HelpKey = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	// Info panel
	InfoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Subtle).
		Padding(1, 2).
		MarginTop(1)

	InfoLabel = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(12)

	InfoValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Tags
	Tag = lipgloss.NewStyle().
		Foreground(Accent).
		Background(lipgloss.Color("#1F2937")).
		Padding(0, 1).
		MarginRight(1)

	// Messages
	ErrorMsg = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	SuccessMsg = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Primary)

	// Search
	SearchPrompt = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	SearchInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Description
	Description = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(60)

	// Group headers for grouped list
	GroupHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(MutedColor)

	GroupHeaderInstalled = lipgloss.NewStyle().
				Bold(true).
				Foreground(Secondary).
				MarginTop(0)

	// Collapse indicator
	CollapseIndicator = lipgloss.NewStyle().
				Foreground(MutedColor)
)

// FormatHelp formats help text with highlighted keys
func FormatHelp(pairs ...string) string {
	var result string
	for i := 0; i < len(pairs); i += 2 {
		if i > 0 {
			result += "  "
		}
		result += HelpKey.Render(pairs[i]) + " " + pairs[i+1]
	}
	return HelpBar.Render(result)
}
