package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/registry"
	"lazyas/internal/tui/styles"
)

// ListItemType indicates whether a list item is a skill or a group header
type ListItemType int

const (
	ItemTypeSkill ListItemType = iota
	ItemTypeHeader
)

// ListItem represents an item in the flattened list (either a skill or a header)
type ListItem struct {
	Type       ListItemType
	Skill      *registry.SkillEntry
	HeaderName string
	Collapsed  bool
	SkillCount int // Number of skills in this group (for headers)
}

// SkillGroup represents a group of skills from the same source
type SkillGroup struct {
	Name      string
	Skills    []registry.SkillEntry
	Collapsed bool
}

// GroupedSkillList displays skills grouped by repository with Installed section at top
type GroupedSkillList struct {
	skills      []registry.SkillEntry
	groups      []SkillGroup
	flatItems   []ListItem
	installed   map[string]bool
	modified    map[string]bool // Skills with local modifications
	cursor      int
	height      int
	offset      int
	collapseMap map[string]bool // Tracks collapsed state per group name
}

// NewGroupedSkillList creates a new grouped skill list
func NewGroupedSkillList(skills []registry.SkillEntry, installed map[string]bool) GroupedSkillList {
	gl := GroupedSkillList{
		skills:      skills,
		installed:   installed,
		modified:    make(map[string]bool),
		cursor:      0,
		height:      10,
		offset:      0,
		collapseMap: make(map[string]bool),
	}
	gl.buildGroups()
	gl.rebuildFlatList()
	return gl
}

// NewGroupedSkillListWithStatus creates a grouped list with modification status
func NewGroupedSkillListWithStatus(skills []registry.SkillEntry, installed, modified map[string]bool) GroupedSkillList {
	gl := GroupedSkillList{
		skills:      skills,
		installed:   installed,
		modified:    modified,
		cursor:      0,
		height:      10,
		offset:      0,
		collapseMap: make(map[string]bool),
	}
	gl.buildGroups()
	gl.rebuildFlatList()
	return gl
}

// buildGroups partitions skills into Installed section and repo groups
func (l *GroupedSkillList) buildGroups() {
	l.groups = nil

	// Separate installed and available skills
	var installedSkills []registry.SkillEntry
	repoGroups := make(map[string][]registry.SkillEntry)

	for _, skill := range l.skills {
		if l.installed[skill.Name] {
			installedSkills = append(installedSkills, skill)
		} else {
			repo := skill.Source.Repo
			repoGroups[repo] = append(repoGroups[repo], skill)
		}
	}

	// Add Installed group first (if any)
	if len(installedSkills) > 0 {
		l.groups = append(l.groups, SkillGroup{
			Name:      "Installed",
			Skills:    installedSkills,
			Collapsed: l.collapseMap["Installed"],
		})
	}

	// Sort repo names for consistent ordering
	var repos []string
	for repo := range repoGroups {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	// Add repo groups
	for _, repo := range repos {
		// Use a shorter display name for the repo
		displayName := formatRepoName(repo)
		l.groups = append(l.groups, SkillGroup{
			Name:      displayName,
			Skills:    repoGroups[repo],
			Collapsed: l.collapseMap[displayName],
		})
	}
}

// formatRepoName extracts a readable name from a repo URL
func formatRepoName(repo string) string {
	// Remove protocol prefix
	name := repo
	if idx := strings.Index(name, "://"); idx != -1 {
		name = name[idx+3:]
	}
	// Remove .git suffix
	name = strings.TrimSuffix(name, ".git")
	return name
}

// rebuildFlatList creates the flat item list from groups
func (l *GroupedSkillList) rebuildFlatList() {
	l.flatItems = nil

	for i := range l.groups {
		group := &l.groups[i]

		// Add header
		l.flatItems = append(l.flatItems, ListItem{
			Type:       ItemTypeHeader,
			HeaderName: group.Name,
			Collapsed:  group.Collapsed,
			SkillCount: len(group.Skills),
		})

		// Add skills if not collapsed
		if !group.Collapsed {
			for j := range group.Skills {
				l.flatItems = append(l.flatItems, ListItem{
					Type:  ItemTypeSkill,
					Skill: &group.Skills[j],
				})
			}
		}
	}

	// Ensure cursor is on a valid skill item
	l.adjustCursor()
}

// adjustCursor ensures cursor is within valid bounds
func (l *GroupedSkillList) adjustCursor() {
	if len(l.flatItems) == 0 {
		l.cursor = 0
		return
	}
	if l.cursor >= len(l.flatItems) {
		l.cursor = len(l.flatItems) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

// SetSkills updates the skills list and rebuilds groups
func (l *GroupedSkillList) SetSkills(skills []registry.SkillEntry) {
	l.skills = skills
	l.buildGroups()
	l.rebuildFlatList()
	l.cursor = 0
	l.offset = 0
	l.adjustCursor()
}

// SetInstalled updates the installed map and rebuilds groups
func (l *GroupedSkillList) SetInstalled(installed map[string]bool) {
	l.installed = installed
	l.buildGroups()
	l.rebuildFlatList()
}

// SetModified updates the modified status map
func (l *GroupedSkillList) SetModified(modified map[string]bool) {
	l.modified = modified
}

// SetHeight sets the visible height
func (l *GroupedSkillList) SetHeight(h int) {
	l.height = h
}

// Selected returns the currently selected skill (or nil if on a header)
func (l *GroupedSkillList) Selected() *registry.SkillEntry {
	if len(l.flatItems) == 0 || l.cursor >= len(l.flatItems) {
		return nil
	}
	item := l.flatItems[l.cursor]
	if item.Type == ItemTypeSkill {
		return item.Skill
	}
	return nil
}

// SelectedIndex returns the current cursor position
func (l *GroupedSkillList) SelectedIndex() int {
	return l.cursor
}

// Len returns the number of visible items
func (l *GroupedSkillList) Len() int {
	return len(l.flatItems)
}

// SkillCount returns the total number of skills
func (l *GroupedSkillList) SkillCount() int {
	return len(l.skills)
}

// Update handles key events
func (l *GroupedSkillList) Update(msg tea.Msg) tea.Cmd {
	km := DefaultKeyMap()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, km.Up):
			l.MoveUp()
		case key.Matches(msg, km.Down):
			l.MoveDown()
		case key.Matches(msg, km.Top):
			l.MoveToTop()
		case key.Matches(msg, km.Bottom):
			l.MoveToBottom()
		case msg.String() == "z" || msg.String() == "tab":
			l.ToggleCurrentGroup()
		}
	}
	return nil
}

// MoveUp moves the cursor up
func (l *GroupedSkillList) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
		l.adjustOffset()
	}
}

// MoveDown moves the cursor down
func (l *GroupedSkillList) MoveDown() {
	if l.cursor < len(l.flatItems)-1 {
		l.cursor++
		l.adjustOffset()
	}
}

// adjustOffset ensures the cursor is visible within the viewport
func (l *GroupedSkillList) adjustOffset() {
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+l.height {
		l.offset = l.cursor - l.height + 1
	}
}

// MoveToTop moves to the first skill
func (l *GroupedSkillList) MoveToTop() {
	l.cursor = 0
	l.offset = 0
	l.adjustCursor()
}

// MoveToBottom moves to the last skill
func (l *GroupedSkillList) MoveToBottom() {
	// Find the last skill
	for i := len(l.flatItems) - 1; i >= 0; i-- {
		if l.flatItems[i].Type == ItemTypeSkill {
			l.cursor = i
			l.adjustOffset()
			return
		}
	}
}

// ToggleCurrentGroup toggles collapse on the group containing the current selection
func (l *GroupedSkillList) ToggleCurrentGroup() {
	if len(l.flatItems) == 0 {
		return
	}

	// Find which group the cursor is in
	groupName := l.findCurrentGroupName()
	if groupName == "" {
		return
	}

	l.ToggleGroup(groupName)
}

// findCurrentGroupName finds the name of the group containing the current cursor
func (l *GroupedSkillList) findCurrentGroupName() string {
	var currentGroup string
	for i := 0; i <= l.cursor && i < len(l.flatItems); i++ {
		if l.flatItems[i].Type == ItemTypeHeader {
			currentGroup = l.flatItems[i].HeaderName
		}
	}
	return currentGroup
}

// ToggleGroup toggles the collapse state of a group by name
func (l *GroupedSkillList) ToggleGroup(name string) {
	l.collapseMap[name] = !l.collapseMap[name]

	// Update the group
	for i := range l.groups {
		if l.groups[i].Name == name {
			l.groups[i].Collapsed = l.collapseMap[name]
			break
		}
	}

	l.rebuildFlatList()
}

// View renders the grouped list
func (l *GroupedSkillList) View() string {
	if len(l.flatItems) == 0 {
		return styles.Muted.Render("No skills found")
	}

	var b strings.Builder

	end := l.offset + l.height
	if end > len(l.flatItems) {
		end = len(l.flatItems)
	}

	for i := l.offset; i < end; i++ {
		item := l.flatItems[i]

		if item.Type == ItemTypeHeader {
			// Render group header
			line := l.renderHeader(item, i == l.cursor)
			b.WriteString(line)
		} else {
			// Render skill item
			line := l.renderSkill(item.Skill, i == l.cursor)
			b.WriteString(line)
		}

		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Show scroll indicator if needed
	totalSkills := l.SkillCount()
	if len(l.flatItems) > l.height && totalSkills > 0 {
		// Count current skill position
		skillIdx := 0
		for i := 0; i <= l.cursor && i < len(l.flatItems); i++ {
			if l.flatItems[i].Type == ItemTypeSkill {
				skillIdx++
			}
		}
		scrollInfo := fmt.Sprintf("\n%s", styles.Muted.Render(
			fmt.Sprintf("  [%d/%d skills]", skillIdx, totalSkills),
		))
		b.WriteString(scrollInfo)
	}

	return b.String()
}

// renderHeader renders a group header
func (l *GroupedSkillList) renderHeader(item ListItem, selected bool) string {
	indicator := "▼"
	if item.Collapsed {
		indicator = "▶"
	}

	headerText := fmt.Sprintf("%s %s (%d)", indicator, item.HeaderName, item.SkillCount)

	if selected {
		return styles.SelectedItem.Render(headerText)
	}
	if item.HeaderName == "Installed" {
		return styles.GroupHeaderInstalled.Render(headerText)
	}
	return styles.GroupHeader.Render(headerText)
}

// renderSkill renders a skill item
func (l *GroupedSkillList) renderSkill(skill *registry.SkillEntry, selected bool) string {
	// Status indicator
	var status string
	if l.installed[skill.Name] {
		if l.modified[skill.Name] {
			status = styles.StatusModified.String()
		} else {
			status = styles.StatusInstalled.String()
		}
	} else {
		status = styles.StatusAvailable.String()
	}

	// Format name and version
	name := skill.Name
	if skill.Source.Tag != "" {
		name = fmt.Sprintf("%s@%s", skill.Name, skill.Source.Tag)
	}

	// Add modified indicator
	if l.modified[skill.Name] {
		name = name + " [modified]"
	}

	// Truncate description
	desc := skill.Description
	if len(desc) > 40 {
		desc = desc[:37] + "..."
	}

	// Build line
	line := fmt.Sprintf("  %s %s", status, name)
	if desc != "" {
		line = fmt.Sprintf("%-42s %s", line, styles.Muted.Render(desc))
	}

	// Apply style based on selection
	if selected {
		line = styles.SelectedItem.Render(line)
	} else {
		line = styles.NormalItem.Render(line)
	}

	return line
}
