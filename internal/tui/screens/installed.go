package screens

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/registry"
	"lazyas/internal/tui/components"
	"lazyas/internal/tui/styles"
)

// InstalledScreen displays installed skills
type InstalledScreen struct {
	manifest SkillManifest
	registry SkillRegistry
	list     components.SkillList
	width    int
	height   int
}

// NewInstalledScreen creates a new installed screen
func NewInstalledScreen(mfst SkillManifest, reg SkillRegistry) *InstalledScreen {
	s := &InstalledScreen{
		manifest: mfst,
		registry: reg,
		width:    80,
		height:   24,
	}
	s.refreshList()
	return s
}

func (s *InstalledScreen) refreshList() {
	installed := s.manifest.ListInstalled()
	skills := make([]registry.SkillEntry, 0, len(installed))
	installedMap := make(map[string]bool)

	// Get skill names sorted
	names := make([]string, 0, len(installed))
	for name := range installed {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		info := installed[name]
		installedMap[name] = true

		// Try to get full info from registry
		if regSkill := s.registry.GetSkill(name); regSkill != nil {
			skills = append(skills, *regSkill)
		} else {
			// Create minimal entry from installed info
			skills = append(skills, registry.SkillEntry{
				Name: name,
				Source: registry.SkillSource{
					Repo: info.SourceRepo,
					Path: info.SourcePath,
					Tag:  info.Version,
				},
			})
		}
	}

	s.list = components.NewSkillList(skills, installedMap)
}

// Init initializes the screen
func (s *InstalledScreen) Init() tea.Cmd {
	return nil
}

// Update handles events
func (s *InstalledScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.list.SetHeight(s.height - 8)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, func() tea.Msg { return BackMsg{} }

		case "enter":
			if skill := s.list.Selected(); skill != nil {
				return s, func() tea.Msg { return SelectSkillMsg{Skill: skill} }
			}

		case "r":
			if skill := s.list.Selected(); skill != nil {
				return s, func() tea.Msg { return RemoveSkillMsg{Skill: skill} }
			}

		case "u":
			if skill := s.list.Selected(); skill != nil {
				return s, func() tea.Msg { return UpdateSkillMsg{Skill: skill} }
			}

		default:
			s.list.Update(msg)
		}
	}

	return s, nil
}

// UpdateSkillMsg requests an update
type UpdateSkillMsg struct {
	Skill *registry.SkillEntry
}

// Refresh updates the list
func (s *InstalledScreen) Refresh() {
	s.refreshList()
}

// View renders the screen
func (s *InstalledScreen) View() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.Title.Render("Installed Skills"))
	b.WriteString("\n\n")

	if s.list.Len() == 0 {
		b.WriteString(styles.Muted.Render("No skills installed"))
		b.WriteString("\n\n")
		b.WriteString(styles.Muted.Render("Press 'b' to browse available skills"))
	} else {
		b.WriteString(s.list.View())
	}

	b.WriteString("\n")

	// Help bar
	b.WriteString(styles.FormatHelp(
		"j/k", "navigate",
		"enter", "details",
		"u", "update",
		"r", "remove",
		"q", "quit",
	))

	return b.String()
}
