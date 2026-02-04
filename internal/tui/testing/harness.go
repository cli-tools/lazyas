package testing

import (
	tea "github.com/charmbracelet/bubbletea"
)

// TestHarness provides a framework for testing Bubble Tea models
type TestHarness struct {
	model tea.Model
}

// NewTestHarness creates a new test harness wrapping a Bubble Tea model
func NewTestHarness(model tea.Model) *TestHarness {
	return &TestHarness{model: model}
}

// Model returns the current model state
func (h *TestHarness) Model() tea.Model {
	return h.model
}

// SendKey sends a single key message and returns the resulting command
func (h *TestHarness) SendKey(key string) tea.Cmd {
	msg := KeyMsg(key)
	var cmd tea.Cmd
	h.model, cmd = h.model.Update(msg)
	return cmd
}

// SendKeys sends multiple key messages in sequence
func (h *TestHarness) SendKeys(keys ...string) []tea.Cmd {
	cmds := make([]tea.Cmd, len(keys))
	for i, key := range keys {
		cmds[i] = h.SendKey(key)
	}
	return cmds
}

// SendMsg sends any tea.Msg to the model
func (h *TestHarness) SendMsg(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	h.model, cmd = h.model.Update(msg)
	return cmd
}

// SendWindowSize sends a window size message
func (h *TestHarness) SendWindowSize(width, height int) tea.Cmd {
	return h.SendMsg(tea.WindowSizeMsg{Width: width, Height: height})
}

// ExecuteCmd executes a command and sends the resulting message to the model
func (h *TestHarness) ExecuteCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg != nil {
		h.model, _ = h.model.Update(msg)
	}
	return msg
}

// ExecuteCmds executes multiple commands in sequence
func (h *TestHarness) ExecuteCmds(cmds []tea.Cmd) []tea.Msg {
	msgs := make([]tea.Msg, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			msg := h.ExecuteCmd(cmd)
			if msg != nil {
				msgs = append(msgs, msg)
			}
		}
	}
	return msgs
}

// View returns the current view of the model
func (h *TestHarness) View() string {
	return h.model.View()
}

// Init initializes the model and returns the init command
func (h *TestHarness) Init() tea.Cmd {
	return h.model.Init()
}

// KeyMsg converts a key string to a tea.KeyMsg
// Supports common key names: "enter", "esc", "tab", "backspace", "up", "down", "left", "right"
// Single characters like "j", "k", "q" are treated as literal keys
// Ctrl combinations like "ctrl+c" are supported
func KeyMsg(key string) tea.KeyMsg {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+z":
		return tea.KeyMsg{Type: tea.KeyCtrlZ}
	default:
		// Treat as a literal rune
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		// Unknown key, return as runes
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}
