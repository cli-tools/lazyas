package panels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"lazyas/internal/registry"
)

// ListItemType indicates whether a list item is a skill or a group header
type ListItemType int

const (
	ItemTypeSkill ListItemType = iota
	ItemTypeHeader
)

// ListItem represents an item in the flattened list
type ListItem struct {
	Type       ListItemType
	Skill      *registry.SkillEntry
	HeaderName string
	Collapsed  bool
	SkillCount int
}

// SkillGroup represents a group of skills
type SkillGroup struct {
	Name      string
	Skills    []registry.SkillEntry
	Collapsed bool
}

// SkillsPanel displays skills in a grouped list
type SkillsPanel struct {
	skills      []registry.SkillEntry
	groups      []SkillGroup
	flatItems   []ListItem
	installed   map[string]bool
	modified    map[string]bool
	cursor      int
	height      int
	width       int
	offset      int
	collapseMap map[string]bool
	focused     bool

	// Search
	searchInput textinput.Model
	searching   bool
	query       string

	// Styles
	styles SkillsPanelStyles
}

// SkillsPanelStyles holds the panel styles
type SkillsPanelStyles struct {
	Title                lipgloss.Style
	StatusInstalled      lipgloss.Style
	StatusAvailable      lipgloss.Style
	StatusModified       lipgloss.Style
	SelectedItem         lipgloss.Style
	NormalItem           lipgloss.Style
	GroupHeader          lipgloss.Style
	GroupHeaderInstalled lipgloss.Style
	Muted                lipgloss.Style
	SearchPrompt         lipgloss.Style
}

// DefaultSkillsPanelStyles returns the default styles
func DefaultSkillsPanelStyles() SkillsPanelStyles {
	return SkillsPanelStyles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")),
		StatusInstalled: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			SetString("●"),
		StatusAvailable: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			SetString("○"),
		StatusModified: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			SetString("◉"),
		SelectedItem: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")),
		NormalItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
		GroupHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6B7280")),
		GroupHeaderInstalled: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981")),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		SearchPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true),
	}
}

// NewSkillsPanel creates a new skills panel
func NewSkillsPanel(skills []registry.SkillEntry, installed, modified map[string]bool) *SkillsPanel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 50

	p := &SkillsPanel{
		skills:      skills,
		installed:   installed,
		modified:    modified,
		collapseMap: make(map[string]bool),
		styles:      DefaultSkillsPanelStyles(),
		searchInput: ti,
		height:      20,
		width:       30,
	}
	p.buildGroups()
	p.rebuildFlatList()
	return p
}

// buildGroups partitions skills into Installed section and repo groups
func (p *SkillsPanel) buildGroups() {
	p.groups = nil

	var installedSkills []registry.SkillEntry
	repoGroups := make(map[string][]registry.SkillEntry)

	for _, skill := range p.skills {
		if p.installed[skill.Name] {
			installedSkills = append(installedSkills, skill)
		} else {
			repo := skill.Source.Repo
			repoGroups[repo] = append(repoGroups[repo], skill)
		}
	}

	// Add Installed group first (if any)
	if len(installedSkills) > 0 {
		p.groups = append(p.groups, SkillGroup{
			Name:      "Installed",
			Skills:    installedSkills,
			Collapsed: p.collapseMap["Installed"],
		})
	}

	// Sort repo names
	var repos []string
	for repo := range repoGroups {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	// Add repo groups
	for _, repo := range repos {
		displayName := formatRepoName(repo)
		p.groups = append(p.groups, SkillGroup{
			Name:      displayName,
			Skills:    repoGroups[repo],
			Collapsed: p.collapseMap[displayName],
		})
	}
}

func formatRepoName(repo string) string {
	name := repo
	if idx := strings.Index(name, "://"); idx != -1 {
		name = name[idx+3:]
	}
	name = strings.TrimSuffix(name, ".git")
	return name
}

// rebuildFlatList creates the flat item list from groups
func (p *SkillsPanel) rebuildFlatList() {
	p.flatItems = nil

	for i := range p.groups {
		group := &p.groups[i]

		p.flatItems = append(p.flatItems, ListItem{
			Type:       ItemTypeHeader,
			HeaderName: group.Name,
			Collapsed:  group.Collapsed,
			SkillCount: len(group.Skills),
		})

		if !group.Collapsed {
			for j := range group.Skills {
				p.flatItems = append(p.flatItems, ListItem{
					Type:  ItemTypeSkill,
					Skill: &group.Skills[j],
				})
			}
		}
	}

	p.adjustCursor()
}

func (p *SkillsPanel) adjustCursor() {
	if len(p.flatItems) == 0 {
		p.cursor = 0
		return
	}
	if p.cursor >= len(p.flatItems) {
		p.cursor = len(p.flatItems) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// SetSize sets the panel dimensions
func (p *SkillsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetFocused sets whether the panel is focused
func (p *SkillsPanel) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns whether the panel is focused
func (p *SkillsPanel) IsFocused() bool {
	return p.focused
}

// SetSkills updates the skills list
func (p *SkillsPanel) SetSkills(skills []registry.SkillEntry) {
	p.skills = skills
	p.buildGroups()
	p.rebuildFlatList()
}

// SetInstalled updates the installed map
func (p *SkillsPanel) SetInstalled(installed map[string]bool) {
	p.installed = installed
	p.buildGroups()
	p.rebuildFlatList()
}

// SetModified updates the modified map
func (p *SkillsPanel) SetModified(modified map[string]bool) {
	p.modified = modified
}

// Selected returns the currently selected skill
func (p *SkillsPanel) Selected() *registry.SkillEntry {
	if len(p.flatItems) == 0 || p.cursor >= len(p.flatItems) {
		return nil
	}
	item := p.flatItems[p.cursor]
	if item.Type == ItemTypeSkill {
		return item.Skill
	}
	return nil
}

// IsSearching returns whether the panel is in search mode
func (p *SkillsPanel) IsSearching() bool {
	return p.searching
}

// KeyMap for the skills panel
type SkillsKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
}

func DefaultSkillsKeyMap() SkillsKeyMap {
	return SkillsKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
		),
	}
}

// Update handles key events
func (p *SkillsPanel) Update(msg tea.Msg) tea.Cmd {
	if p.searching {
		return p.handleSearchInput(msg)
	}

	km := DefaultSkillsKeyMap()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, km.Up):
			p.moveUp()
		case key.Matches(msg, km.Down):
			p.moveDown()
		case key.Matches(msg, km.Top):
			p.moveToTop()
		case key.Matches(msg, km.Bottom):
			p.moveToBottom()
		case msg.String() == "z" || msg.String() == "tab":
			p.toggleCurrentGroup()
		case msg.String() == "/":
			p.searching = true
			p.searchInput.Focus()
			return textinput.Blink
		}
	}
	return nil
}

func (p *SkillsPanel) handleSearchInput(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			p.searching = false
			p.query = p.searchInput.Value()
			return nil
		case "esc":
			p.searching = false
			p.searchInput.SetValue(p.query)
			return nil
		}
	}

	var cmd tea.Cmd
	p.searchInput, cmd = p.searchInput.Update(msg)
	return cmd
}

// GetQuery returns the current search query
func (p *SkillsPanel) GetQuery() string {
	return p.query
}

// ClearSearch clears the search
func (p *SkillsPanel) ClearSearch() {
	p.query = ""
	p.searchInput.SetValue("")
}

func (p *SkillsPanel) moveUp() {
	if p.cursor > 0 {
		p.cursor--
		p.adjustOffset()
	}
}

func (p *SkillsPanel) moveDown() {
	if p.cursor < len(p.flatItems)-1 {
		p.cursor++
		p.adjustOffset()
	}
}

func (p *SkillsPanel) adjustOffset() {
	visibleHeight := p.height - 4 // Account for header and padding
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+visibleHeight {
		p.offset = p.cursor - visibleHeight + 1
	}
}

func (p *SkillsPanel) moveToTop() {
	p.cursor = 0
	p.offset = 0
}

func (p *SkillsPanel) moveToBottom() {
	for i := len(p.flatItems) - 1; i >= 0; i-- {
		if p.flatItems[i].Type == ItemTypeSkill {
			p.cursor = i
			p.adjustOffset()
			return
		}
	}
}

func (p *SkillsPanel) toggleCurrentGroup() {
	if len(p.flatItems) == 0 {
		return
	}

	groupName := p.findCurrentGroupName()
	if groupName == "" {
		return
	}

	p.collapseMap[groupName] = !p.collapseMap[groupName]

	for i := range p.groups {
		if p.groups[i].Name == groupName {
			p.groups[i].Collapsed = p.collapseMap[groupName]
			break
		}
	}

	p.rebuildFlatList()
}

func (p *SkillsPanel) findCurrentGroupName() string {
	var currentGroup string
	for i := 0; i <= p.cursor && i < len(p.flatItems); i++ {
		if p.flatItems[i].Type == ItemTypeHeader {
			currentGroup = p.flatItems[i].HeaderName
		}
	}
	return currentGroup
}

// View renders the skills panel
func (p *SkillsPanel) View() string {
	var b strings.Builder

	// Search bar
	if p.searching {
		b.WriteString(p.styles.SearchPrompt.Render("/") + " ")
		b.WriteString(p.searchInput.View())
		b.WriteString("\n")
	} else if p.query != "" {
		b.WriteString(p.styles.Muted.Render("Search: " + p.query))
		b.WriteString("\n")
	}

	if len(p.flatItems) == 0 {
		b.WriteString(p.styles.Muted.Render("No skills found"))
		return b.String()
	}

	visibleHeight := p.height - 4
	if p.searching || p.query != "" {
		visibleHeight--
	}

	end := p.offset + visibleHeight
	if end > len(p.flatItems) {
		end = len(p.flatItems)
	}

	for i := p.offset; i < end; i++ {
		item := p.flatItems[i]

		if item.Type == ItemTypeHeader {
			line := p.renderHeader(item, i == p.cursor)
			b.WriteString(line)
		} else {
			line := p.renderSkill(item.Skill, i == p.cursor)
			b.WriteString(line)
		}

		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (p *SkillsPanel) renderHeader(item ListItem, selected bool) string {
	indicator := "▼"
	if item.Collapsed {
		indicator = "▶"
	}

	headerText := fmt.Sprintf("%s %s (%d)", indicator, item.HeaderName, item.SkillCount)

	// Truncate if too wide
	maxWidth := p.width - 2
	if len(headerText) > maxWidth {
		headerText = headerText[:maxWidth-3] + "..."
	}

	if selected && p.focused {
		return p.styles.SelectedItem.Render(headerText)
	}
	if item.HeaderName == "Installed" {
		return p.styles.GroupHeaderInstalled.Render(headerText)
	}
	return p.styles.GroupHeader.Render(headerText)
}

func (p *SkillsPanel) renderSkill(skill *registry.SkillEntry, selected bool) string {
	var status string
	if p.installed[skill.Name] {
		if p.modified[skill.Name] {
			status = p.styles.StatusModified.String()
		} else {
			status = p.styles.StatusInstalled.String()
		}
	} else {
		status = p.styles.StatusAvailable.String()
	}

	name := skill.Name
	if p.modified[skill.Name] {
		name = name + "*"
	}

	// Truncate if too wide
	maxWidth := p.width - 6
	if len(name) > maxWidth {
		name = name[:maxWidth-3] + "..."
	}

	line := fmt.Sprintf("  %s %s", status, name)

	if selected && p.focused {
		return p.styles.SelectedItem.Render(line)
	}
	return p.styles.NormalItem.Render(line)
}
