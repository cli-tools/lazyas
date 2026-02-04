package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// Panel represents the currently focused panel
type Panel int

const (
	PanelLeft Panel = iota
	PanelRight
)

// PanelLayout manages a two-panel layout with focus tracking
type PanelLayout struct {
	focus      Panel
	leftWidth  int
	rightWidth int
	height     int
	totalWidth int
	splitRatio float64 // Ratio for left panel (0.0-1.0)
}

// NewPanelLayout creates a new panel layout with default 30/70 split
func NewPanelLayout() *PanelLayout {
	return &PanelLayout{
		focus:      PanelLeft,
		splitRatio: 0.30,
	}
}

// SetSize updates the layout dimensions
func (p *PanelLayout) SetSize(width, height int) {
	p.totalWidth = width
	p.height = height
	p.leftWidth = int(float64(width) * p.splitRatio)
	p.rightWidth = width - p.leftWidth - 1 // -1 for separator
}

// Focus returns the currently focused panel
func (p *PanelLayout) Focus() Panel {
	return p.focus
}

// FocusLeft sets focus to the left panel
func (p *PanelLayout) FocusLeft() {
	p.focus = PanelLeft
}

// FocusRight sets focus to the right panel
func (p *PanelLayout) FocusRight() {
	p.focus = PanelRight
}

// ToggleFocus switches focus between panels
func (p *PanelLayout) ToggleFocus() {
	if p.focus == PanelLeft {
		p.focus = PanelRight
	} else {
		p.focus = PanelLeft
	}
}

// LeftWidth returns the width allocated to the left panel
func (p *PanelLayout) LeftWidth() int {
	return p.leftWidth
}

// RightWidth returns the width allocated to the right panel
func (p *PanelLayout) RightWidth() int {
	return p.rightWidth
}

// Height returns the panel height
func (p *PanelLayout) Height() int {
	return p.height
}

// ContentHeight returns the height available for content (minus borders)
func (p *PanelLayout) ContentHeight() int {
	return p.height - 2 // -2 for top and bottom border
}

// LeftContentWidth returns width available for left panel content (minus borders)
func (p *PanelLayout) LeftContentWidth() int {
	return p.leftWidth - 2 // -2 for left and right border
}

// RightContentWidth returns width available for right panel content (minus borders)
func (p *PanelLayout) RightContentWidth() int {
	return p.rightWidth - 2 // -2 for left and right border
}

// PanelStyles holds styles for panel borders
type PanelStyles struct {
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style
}

// DefaultPanelStyles returns the default panel styles
func DefaultPanelStyles() PanelStyles {
	return PanelStyles{
		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")), // Purple
		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")), // Dark gray
	}
}
