package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
	"lazyas/internal/tui/components"
	"lazyas/internal/tui/styles"
)

// BrowseScreen displays the skill browser
type BrowseScreen struct {
	registry    SkillRegistry
	manifest    SkillManifest
	list        components.GroupedSkillList
	searchInput textinput.Model
	searching   bool
	query       string
	width       int
	height      int
}

// NewBrowseScreen creates a new browse screen
func NewBrowseScreen(reg SkillRegistry, mfst SkillManifest) *BrowseScreen {
	ti := textinput.New()
	ti.Placeholder = "Search skills..."
	ti.CharLimit = 50

	// Scan local skills directory to determine what's installed and modified
	localSkills := mfst.ScanLocalSkills()
	installed := make(map[string]bool)
	modified := make(map[string]bool)
	for name, local := range localSkills {
		installed[name] = true
		if local.IsModified {
			modified[name] = true
		}
	}

	// Merge registry skills with local-only skills
	skills := mergeSkills(reg.ListSkills(), localSkills)

	return &BrowseScreen{
		registry:    reg,
		manifest:    mfst,
		list:        components.NewGroupedSkillListWithStatus(skills, installed, modified),
		searchInput: ti,
		width:       80,
		height:      24,
	}
}

// mergeSkills combines registry skills with local-only skills
func mergeSkills(registrySkills []registry.SkillEntry, localSkills map[string]LocalSkill) []registry.SkillEntry {
	// Create a map for quick lookup
	seen := make(map[string]bool)
	result := make([]registry.SkillEntry, 0, len(registrySkills)+len(localSkills))

	// Add all registry skills
	for _, skill := range registrySkills {
		result = append(result, skill)
		seen[skill.Name] = true
	}

	// Add local skills that aren't in registry
	for name, local := range localSkills {
		if !seen[name] {
			result = append(result, registry.SkillEntry{
				Name:        name,
				Description: local.Description,
				Source: registry.SkillSource{
					Repo: local.Path, // Use path as "repo" for local skills
				},
			})
		}
	}

	return result
}

// LocalSkill mirrors manifest.LocalSkill for interface compatibility
type LocalSkill = manifest.LocalSkill

// Init initializes the screen
func (s *BrowseScreen) Init() tea.Cmd {
	return nil
}

// BrowseMsg types for screen communication
type (
	SelectSkillMsg struct {
		Skill *registry.SkillEntry
	}
	InstallSkillMsg struct {
		Skill *registry.SkillEntry
	}
	RemoveSkillMsg struct {
		Skill *registry.SkillEntry
	}
	BackMsg struct{}
)

// Update handles events
func (s *BrowseScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.list.SetHeight(s.height - 10) // Reserve space for header/footer
		return s, nil

	case tea.KeyMsg:
		if s.searching {
			return s.handleSearchInput(msg)
		}
		return s.handleNavigation(msg)
	}

	return s, nil
}

func (s *BrowseScreen) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		s.searching = false
		s.query = s.searchInput.Value()
		s.filterSkills()
		return s, nil

	case "esc":
		s.searching = false
		s.searchInput.SetValue(s.query)
		return s, nil
	}

	var cmd tea.Cmd
	s.searchInput, cmd = s.searchInput.Update(msg)
	return s, cmd
}

func (s *BrowseScreen) handleNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return s, func() tea.Msg { return BackMsg{} }

	case "/":
		s.searching = true
		s.searchInput.Focus()
		return s, textinput.Blink

	case "enter":
		if skill := s.list.Selected(); skill != nil {
			return s, func() tea.Msg { return SelectSkillMsg{Skill: skill} }
		}

	case "i":
		if skill := s.list.Selected(); skill != nil {
			if !s.manifest.IsInstalled(skill.Name) {
				return s, func() tea.Msg { return InstallSkillMsg{Skill: skill} }
			}
		}

	case "r":
		if skill := s.list.Selected(); skill != nil {
			if s.manifest.IsInstalled(skill.Name) {
				return s, func() tea.Msg { return RemoveSkillMsg{Skill: skill} }
			}
		}

	case "c":
		// Clear search
		s.query = ""
		s.searchInput.SetValue("")
		s.filterSkills()

	case "z", "tab":
		// Toggle group collapse (handled by list)
		s.list.Update(msg)

	default:
		s.list.Update(msg)
	}

	return s, nil
}

func (s *BrowseScreen) filterSkills() {
	localSkills := s.manifest.ScanLocalSkills()
	var skills []registry.SkillEntry
	if s.query == "" {
		skills = mergeSkills(s.registry.ListSkills(), localSkills)
	} else {
		skills = mergeSkills(s.registry.SearchSkills(s.query), localSkills)
	}
	s.list.SetSkills(skills)
}

// RefreshInstalled updates the installed and modified state by rescanning local skills
func (s *BrowseScreen) RefreshInstalled() {
	localSkills := s.manifest.ScanLocalSkills()
	installed := make(map[string]bool)
	modified := make(map[string]bool)
	for name, local := range localSkills {
		installed[name] = true
		if local.IsModified {
			modified[name] = true
		}
	}
	s.list.SetInstalled(installed)
	s.list.SetModified(modified)
}

// View renders the screen
func (s *BrowseScreen) View() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.Title.Render("lazyas - Browse Skills"))
	b.WriteString("\n")

	// Search bar
	if s.searching {
		b.WriteString(styles.SearchPrompt.Render("/") + " ")
		b.WriteString(s.searchInput.View())
	} else if s.query != "" {
		b.WriteString(styles.Muted.Render("Search: ") + s.query)
	}
	b.WriteString("\n\n")

	// Legend
	b.WriteString(styles.Muted.Render("● installed  ○ available"))
	b.WriteString("\n\n")

	// Skill list
	b.WriteString(s.list.View())
	b.WriteString("\n")

	// Help bar
	help := styles.FormatHelp(
		"j/k", "navigate",
		"z", "collapse",
		"enter", "details",
		"i", "install",
		"r", "remove",
		"/", "search",
		"q", "quit",
	)
	b.WriteString(help)

	return b.String()
}
