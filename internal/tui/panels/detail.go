package panels

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

// Tab represents the current detail tab
type Tab int

const (
	TabInfo Tab = iota
	TabSkillMD
)

// DetailPanel displays skill details with tabs
type DetailPanel struct {
	skill        *registry.SkillEntry
	installed    *manifest.InstalledSkill
	localInfo    *manifest.LocalSkill
	tab          Tab
	height       int
	width        int
	focused      bool
	viewport     viewport.Model
	infoViewport viewport.Model
	skillMD      string
	isOutdated   bool

	// Styles
	styles DetailPanelStyles
}

// DetailPanelStyles holds the panel styles
type DetailPanelStyles struct {
	Title         lipgloss.Style
	TabActive     lipgloss.Style
	TabInactive   lipgloss.Style
	TabBar        lipgloss.Style
	Label         lipgloss.Style
	Value         lipgloss.Style
	Muted         lipgloss.Style
	Tag           lipgloss.Style
	Badge         lipgloss.Style
	BadgeModified lipgloss.Style
	BadgeOutdated lipgloss.Style
}

// DefaultDetailPanelStyles returns the default styles
func DefaultDetailPanelStyles() DetailPanelStyles {
	return DetailPanelStyles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1),
		TabBar: lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#374151")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Tag: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1).
			MarginRight(1),
		Badge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true),
		BadgeModified: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true),
		BadgeOutdated: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#818CF8")).
			Bold(true),
	}
}

// NewDetailPanel creates a new detail panel
func NewDetailPanel() *DetailPanel {
	vp := viewport.New(80, 20)
	ivp := viewport.New(80, 20)
	return &DetailPanel{
		tab:          TabInfo,
		styles:       DefaultDetailPanelStyles(),
		viewport:     vp,
		infoViewport: ivp,
		height:       24,
		width:        60,
	}
}

// SetSkill sets the skill to display
func (p *DetailPanel) SetSkill(skill *registry.SkillEntry, installed *manifest.InstalledSkill, local *manifest.LocalSkill, skillsDir string) {
	p.skill = skill
	p.installed = installed
	p.localInfo = local
	p.skillMD = ""

	// Try to load SKILL.md if installed
	if skill != nil && local != nil {
		skillMDPath := filepath.Join(skillsDir, skill.Name, "SKILL.md")
		if content, err := os.ReadFile(skillMDPath); err == nil {
			p.skillMD = string(content)
		}
	}

	// Update viewport content
	if skill != nil {
		p.infoViewport.SetContent(p.renderInfo())
		p.infoViewport.GotoTop()
	}
	if p.tab == TabSkillMD {
		p.viewport.SetContent(p.skillMD)
	}
}

// SetOutdated sets whether the current skill has an update available
func (p *DetailPanel) SetOutdated(outdated bool) {
	p.isOutdated = outdated
	if p.skill != nil {
		p.infoViewport.SetContent(p.renderInfo())
	}
}

// SetSize sets the panel dimensions
func (p *DetailPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 4
	p.viewport.Height = height - 8 // Account for tabs and padding
	p.infoViewport.Width = width - 4
	p.infoViewport.Height = height - 8
}

// SetFocused sets whether the panel is focused
func (p *DetailPanel) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns whether the panel is focused
func (p *DetailPanel) IsFocused() bool {
	return p.focused
}

// DetailKeyMap for the detail panel
type DetailKeyMap struct {
	PrevTab key.Binding
	NextTab key.Binding
	Up      key.Binding
	Down    key.Binding
}

func DefaultDetailKeyMap() DetailKeyMap {
	return DetailKeyMap{
		PrevTab: key.NewBinding(key.WithKeys("[")),
		NextTab: key.NewBinding(key.WithKeys("]")),
		Up:      key.NewBinding(key.WithKeys("up", "k")),
		Down:    key.NewBinding(key.WithKeys("down", "j")),
	}
}

// Update handles key events
func (p *DetailPanel) Update(msg tea.Msg) tea.Cmd {
	km := DefaultDetailKeyMap()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, km.PrevTab):
			if p.tab > 0 {
				p.tab--
			}
		case key.Matches(msg, km.NextTab):
			if p.tab < TabSkillMD {
				p.tab++
				if p.tab == TabSkillMD {
					p.viewport.SetContent(p.skillMD)
					p.viewport.GotoTop()
				}
			}
		case key.Matches(msg, km.Up), key.Matches(msg, km.Down):
			switch p.tab {
			case TabInfo:
				var cmd tea.Cmd
				p.infoViewport, cmd = p.infoViewport.Update(msg)
				return cmd
			case TabSkillMD:
				var cmd tea.Cmd
				p.viewport, cmd = p.viewport.Update(msg)
				return cmd
			}
		}
	}
	return nil
}

// View renders the detail panel
func (p *DetailPanel) View() string {
	if p.skill == nil {
		return p.styles.Muted.Render("Select a skill to view details")
	}

	var b strings.Builder

	// Tabs
	b.WriteString(p.renderTabs())
	b.WriteString("\n\n")

	// Content based on tab
	switch p.tab {
	case TabInfo:
		b.WriteString(p.infoViewport.View())
	case TabSkillMD:
		b.WriteString(p.renderSkillMD())
	}

	return b.String()
}

func (p *DetailPanel) renderTabs() string {
	tabs := []string{"Info", "SKILL.md"}
	var rendered []string

	for i, tab := range tabs {
		if Tab(i) == p.tab {
			rendered = append(rendered, p.styles.TabActive.Render(tab))
		} else {
			rendered = append(rendered, p.styles.TabInactive.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (p *DetailPanel) renderInfo() string {
	var b strings.Builder

	// Title + status badge on same line
	title := p.skill.Name
	if p.skill.Source.Tag != "" {
		title += "@" + p.skill.Source.Tag
	}
	b.WriteString(p.styles.Title.Render(title))

	isUntracked := p.localInfo != nil && p.installed == nil
	if p.localInfo != nil {
		b.WriteString("  ")
		if p.localInfo.IsModified {
			b.WriteString(p.styles.BadgeModified.Render("● MODIFIED"))
		} else if p.isOutdated {
			b.WriteString(p.styles.BadgeOutdated.Render("↑ UPDATE AVAILABLE"))
		} else {
			b.WriteString(p.styles.Badge.Render("● INSTALLED"))
		}
		if p.installed != nil {
			b.WriteString(p.styles.Muted.Render(" " + truncate(p.installed.Commit, 7)))
		} else {
			b.WriteString(p.styles.Muted.Render(" (untracked)"))
		}
	} else {
		b.WriteString("  ")
		b.WriteString(p.styles.Muted.Render("○ Not installed"))
	}
	b.WriteString("\n")

	// Show hint when both modified AND outdated
	if p.localInfo != nil && p.localInfo.IsModified && p.isOutdated {
		b.WriteString(p.styles.BadgeOutdated.Render("  ↑ Update available"))
		b.WriteString(p.styles.Muted.Render(" (commit or discard local changes first)"))
		b.WriteString("\n")
	}

	if isUntracked {
		// Untracked skill: show local path, not registry source info
		b.WriteString(p.styles.Label.Render("Location"))
		loc := p.localInfo.Path
		if len(loc) > p.width-14 {
			loc = loc[:p.width-17] + "..."
		}
		b.WriteString(p.styles.Value.Render(loc))
		b.WriteString("\n")
		b.WriteString(p.styles.Muted.Render("Not managed by lazyas. Use 'i' to install from registry."))
		b.WriteString("\n")
	} else {
		// Author
		if p.skill.Author != "" {
			b.WriteString(p.styles.Label.Render("Author"))
			b.WriteString(p.styles.Value.Render(p.skill.Author))
			b.WriteString("\n")
		}

		// Repository
		b.WriteString(p.styles.Label.Render("Repository"))
		repo := p.skill.Source.Repo
		if len(repo) > p.width-14 {
			repo = repo[:p.width-17] + "..."
		}
		b.WriteString(p.styles.Value.Render(repo))
		b.WriteString("\n")

		// Path (if present)
		if p.skill.Source.Path != "" {
			b.WriteString(p.styles.Label.Render("Path"))
			b.WriteString(p.styles.Value.Render(p.skill.Source.Path))
			b.WriteString("\n")
		}

		// Version
		b.WriteString(p.styles.Label.Render("Version"))
		version := p.skill.Source.Tag
		if version == "" {
			version = "latest"
		}
		b.WriteString(p.styles.Value.Render(version))
		b.WriteString("\n")
	}

	// Tags
	if len(p.skill.Tags) > 0 {
		b.WriteString(p.styles.Label.Render("Tags"))
		for _, tag := range p.skill.Tags {
			b.WriteString(p.styles.Tag.Render(tag))
		}
		b.WriteString("\n")
	}

	// Description (last, since it can be multi-line)
	if p.skill.Description != "" {
		b.WriteString("\n")
		b.WriteString(wordWrap(p.skill.Description, p.width-4))
	}

	return b.String()
}

func (p *DetailPanel) renderSkillMD() string {
	if p.skillMD == "" {
		if p.localInfo == nil {
			return p.styles.Muted.Render("Install skill to view SKILL.md")
		}
		return p.styles.Muted.Render("SKILL.md not found")
	}

	return p.viewport.View()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}

	var result strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for i, word := range words {
		if i > 0 {
			if lineLen+1+len(word) > width {
				result.WriteString("\n")
				lineLen = 0
			} else {
				result.WriteString(" ")
				lineLen++
			}
		}
		result.WriteString(word)
		lineLen += len(word)
	}

	return result.String()
}
