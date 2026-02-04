package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/tui/components"
	"lazyas/internal/tui/styles"
)

// LoadingScreen shows a loading spinner
type LoadingScreen struct {
	spinner components.Spinner
	message string
	width   int
	height  int
}

// NewLoadingScreen creates a new loading screen
func NewLoadingScreen(message string) *LoadingScreen {
	return &LoadingScreen{
		spinner: components.NewSpinner(message),
		message: message,
		width:   80,
		height:  24,
	}
}

// Init initializes the screen
func (s *LoadingScreen) Init() tea.Cmd {
	return s.spinner.Tick()
}

// Update handles events
func (s *LoadingScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	}

	cmd := s.spinner.Update(msg)
	return s, cmd
}

// SetMessage updates the loading message
func (s *LoadingScreen) SetMessage(msg string) {
	s.message = msg
	s.spinner.SetMessage(msg)
}

// View renders the screen
func (s *LoadingScreen) View() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("lazyas"))
	b.WriteString("\n\n")
	b.WriteString(s.spinner.View())
	b.WriteString("\n")

	return b.String()
}
