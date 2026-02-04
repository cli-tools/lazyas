package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/tui/styles"
)

// Spinner wraps the bubbles spinner
type Spinner struct {
	spinner spinner.Model
	message string
}

// NewSpinner creates a new spinner
func NewSpinner(message string) Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle
	return Spinner{
		spinner: s,
		message: message,
	}
}

// SetMessage updates the spinner message
func (s *Spinner) SetMessage(msg string) {
	s.message = msg
}

// Init initializes the spinner
func (s Spinner) Init() tea.Cmd {
	return s.spinner.Tick
}

// Update handles spinner updates
func (s *Spinner) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return cmd
}

// View renders the spinner
func (s Spinner) View() string {
	return s.spinner.View() + " " + s.message
}

// Tick returns the tick command
func (s Spinner) Tick() tea.Cmd {
	return s.spinner.Tick
}
