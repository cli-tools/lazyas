package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/registry"
	"lazyas/internal/tui/styles"
)

// ConfirmAction represents the action to confirm
type ConfirmAction int

const (
	ConfirmInstall ConfirmAction = iota
	ConfirmRemove
)

// ConfirmScreen shows a confirmation dialog
type ConfirmScreen struct {
	action   ConfirmAction
	skill    *registry.SkillEntry
	selected int // 0 = yes, 1 = no
	width    int
	height   int
}

// NewConfirmScreen creates a new confirm screen
func NewConfirmScreen(action ConfirmAction, skill *registry.SkillEntry) *ConfirmScreen {
	return &ConfirmScreen{
		action:   action,
		skill:    skill,
		selected: 0,
		width:    80,
		height:   24,
	}
}

// Init initializes the screen
func (s *ConfirmScreen) Init() tea.Cmd {
	return nil
}

// ConfirmedMsg is sent when action is confirmed
type ConfirmedMsg struct {
	Action ConfirmAction
	Skill  *registry.SkillEntry
}

// CancelledMsg is sent when action is cancelled
type CancelledMsg struct{}

// Update handles events
func (s *ConfirmScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			s.selected = 0
		case "right", "l":
			s.selected = 1
		case "y", "Y":
			s.selected = 0
			return s, s.confirm()
		case "n", "N", "esc", "q":
			return s, func() tea.Msg { return CancelledMsg{} }
		case "enter":
			return s, s.confirm()
		}
	}

	return s, nil
}

func (s *ConfirmScreen) confirm() tea.Cmd {
	if s.selected == 0 {
		return func() tea.Msg {
			return ConfirmedMsg{Action: s.action, Skill: s.skill}
		}
	}
	return func() tea.Msg { return CancelledMsg{} }
}

// View renders the screen
func (s *ConfirmScreen) View() string {
	var b strings.Builder

	// Title
	var title string
	var message string
	switch s.action {
	case ConfirmInstall:
		title = "Install Skill"
		message = fmt.Sprintf("Install %s?", styles.InfoValue.Render(s.skill.Name))
	case ConfirmRemove:
		title = "Remove Skill"
		message = fmt.Sprintf("Remove %s?", styles.InfoValue.Render(s.skill.Name))
	}

	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")
	b.WriteString(message)
	b.WriteString("\n\n")

	// Show source info for install
	if s.action == ConfirmInstall {
		b.WriteString(styles.Muted.Render("Source: " + s.skill.Source.Repo))
		if s.skill.Source.Path != "" {
			b.WriteString(styles.Muted.Render(" (" + s.skill.Source.Path + ")"))
		}
		b.WriteString("\n\n")
	}

	// Buttons
	yesBtn := "[ Yes ]"
	noBtn := "[ No ]"

	if s.selected == 0 {
		yesBtn = styles.SelectedItem.Render(" Yes ")
		noBtn = styles.NormalItem.Render(" No ")
	} else {
		yesBtn = styles.NormalItem.Render(" Yes ")
		noBtn = styles.SelectedItem.Render(" No ")
	}

	b.WriteString(yesBtn + "  " + noBtn)
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.FormatHelp(
		"y", "yes",
		"n", "no",
		"←/→", "select",
		"enter", "confirm",
	))

	return b.String()
}
