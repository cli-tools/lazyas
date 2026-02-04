package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/registry"
	"lazyas/internal/tui/styles"
)

// SkillList is a component for displaying a list of skills
type SkillList struct {
	skills    []registry.SkillEntry
	installed map[string]bool
	cursor    int
	height    int
	offset    int
}

// NewSkillList creates a new skill list
func NewSkillList(skills []registry.SkillEntry, installed map[string]bool) SkillList {
	return SkillList{
		skills:    skills,
		installed: installed,
		cursor:    0,
		height:    10,
		offset:    0,
	}
}

// SetSkills updates the skills list
func (l *SkillList) SetSkills(skills []registry.SkillEntry) {
	l.skills = skills
	l.cursor = 0
	l.offset = 0
}

// SetInstalled updates the installed map
func (l *SkillList) SetInstalled(installed map[string]bool) {
	l.installed = installed
}

// SetHeight sets the visible height
func (l *SkillList) SetHeight(h int) {
	l.height = h
}

// Selected returns the currently selected skill
func (l *SkillList) Selected() *registry.SkillEntry {
	if len(l.skills) == 0 {
		return nil
	}
	return &l.skills[l.cursor]
}

// SelectedIndex returns the current cursor position
func (l *SkillList) SelectedIndex() int {
	return l.cursor
}

// Len returns the number of skills
func (l *SkillList) Len() int {
	return len(l.skills)
}

// KeyMap defines key bindings for the list
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
	}
}

// Update handles key events
func (l *SkillList) Update(msg tea.Msg) tea.Cmd {
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
		}
	}
	return nil
}

// MoveUp moves the cursor up
func (l *SkillList) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
		if l.cursor < l.offset {
			l.offset = l.cursor
		}
	}
}

// MoveDown moves the cursor down
func (l *SkillList) MoveDown() {
	if l.cursor < len(l.skills)-1 {
		l.cursor++
		if l.cursor >= l.offset+l.height {
			l.offset = l.cursor - l.height + 1
		}
	}
}

// MoveToTop moves to the first item
func (l *SkillList) MoveToTop() {
	l.cursor = 0
	l.offset = 0
}

// MoveToBottom moves to the last item
func (l *SkillList) MoveToBottom() {
	l.cursor = len(l.skills) - 1
	if l.cursor >= l.height {
		l.offset = l.cursor - l.height + 1
	}
}

// View renders the list
func (l *SkillList) View() string {
	if len(l.skills) == 0 {
		return styles.Muted.Render("No skills found")
	}

	var b strings.Builder

	end := l.offset + l.height
	if end > len(l.skills) {
		end = len(l.skills)
	}

	for i := l.offset; i < end; i++ {
		skill := l.skills[i]

		// Status indicator
		var status string
		if l.installed[skill.Name] {
			status = styles.StatusInstalled.String()
		} else {
			status = styles.StatusAvailable.String()
		}

		// Format name and version
		name := skill.Name
		if skill.Source.Tag != "" {
			name = fmt.Sprintf("%s@%s", skill.Name, skill.Source.Tag)
		}

		// Truncate description
		desc := skill.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		// Build line
		line := fmt.Sprintf("%s %s", status, name)
		if desc != "" {
			line = fmt.Sprintf("%-35s %s", line, styles.Muted.Render(desc))
		}

		// Apply style based on selection
		if i == l.cursor {
			line = styles.SelectedItem.Render(line)
		} else {
			line = styles.NormalItem.Render(line)
		}

		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Show scroll indicator if needed
	if len(l.skills) > l.height {
		scrollInfo := fmt.Sprintf("\n%s", styles.Muted.Render(
			fmt.Sprintf("  [%d/%d]", l.cursor+1, len(l.skills)),
		))
		b.WriteString(scrollInfo)
	}

	return b.String()
}
