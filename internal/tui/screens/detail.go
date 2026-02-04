package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/registry"
	"lazyas/internal/tui/styles"
)

// DetailScreen shows skill details
type DetailScreen struct {
	skill     *registry.SkillEntry
	manifest  SkillManifest
	installed bool
	width     int
	height    int
}

// NewDetailScreen creates a new detail screen
func NewDetailScreen(skill *registry.SkillEntry, mfst SkillManifest) *DetailScreen {
	return &DetailScreen{
		skill:     skill,
		manifest:  mfst,
		installed: mfst.IsInstalled(skill.Name),
		width:     80,
		height:    24,
	}
}

// Init initializes the screen
func (s *DetailScreen) Init() tea.Cmd {
	return nil
}

// Update handles events
func (s *DetailScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "backspace":
			return s, func() tea.Msg { return BackMsg{} }

		case "i":
			if !s.installed {
				return s, func() tea.Msg { return InstallSkillMsg{Skill: s.skill} }
			}

		case "r":
			if s.installed {
				return s, func() tea.Msg { return RemoveSkillMsg{Skill: s.skill} }
			}
		}
	}

	return s, nil
}

// View renders the screen
func (s *DetailScreen) View() string {
	var b strings.Builder

	// Title
	title := s.skill.Name
	if s.skill.Source.Tag != "" {
		title = fmt.Sprintf("%s@%s", s.skill.Name, s.skill.Source.Tag)
	}
	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n")

	// Status
	if s.installed {
		info, _ := s.manifest.GetInstalled(s.skill.Name)
		b.WriteString(styles.InstalledBadge.Render("● INSTALLED"))
		b.WriteString(styles.Muted.Render(fmt.Sprintf(" (commit: %s)", truncate(info.Commit, 7))))
	} else {
		b.WriteString(styles.Muted.Render("○ Not installed"))
	}
	b.WriteString("\n\n")

	// Info box
	var info strings.Builder

	// Description
	info.WriteString(styles.InfoLabel.Render("Description"))
	info.WriteString(styles.InfoValue.Render(s.skill.Description))
	info.WriteString("\n\n")

	// Author
	info.WriteString(styles.InfoLabel.Render("Author"))
	info.WriteString(styles.InfoValue.Render(s.skill.Author))
	info.WriteString("\n\n")

	// Source
	info.WriteString(styles.InfoLabel.Render("Repository"))
	info.WriteString(styles.InfoValue.Render(s.skill.Source.Repo))
	info.WriteString("\n")

	if s.skill.Source.Path != "" {
		info.WriteString(styles.InfoLabel.Render("Path"))
		info.WriteString(styles.InfoValue.Render(s.skill.Source.Path))
		info.WriteString("\n")
	}

	info.WriteString(styles.InfoLabel.Render("Version"))
	version := s.skill.Source.Tag
	if version == "" {
		version = "latest"
	}
	info.WriteString(styles.InfoValue.Render(version))
	info.WriteString("\n\n")

	// Tags
	if len(s.skill.Tags) > 0 {
		info.WriteString(styles.InfoLabel.Render("Tags"))
		for _, tag := range s.skill.Tags {
			info.WriteString(styles.Tag.Render(tag))
		}
	}

	b.WriteString(styles.InfoBox.Render(info.String()))
	b.WriteString("\n")

	// Help bar
	var helpItems []string
	if s.installed {
		helpItems = append(helpItems, "r", "remove")
	} else {
		helpItems = append(helpItems, "i", "install")
	}
	helpItems = append(helpItems, "esc", "back", "q", "quit")

	b.WriteString(styles.FormatHelp(helpItems...))

	return b.String()
}

// Refresh updates the installed state
func (s *DetailScreen) Refresh() {
	s.installed = s.manifest.IsInstalled(s.skill.Name)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
