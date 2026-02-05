package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"lazyas/internal/config"
	"lazyas/internal/git"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
	"lazyas/internal/symlink"
	"lazyas/internal/tui/layout"
	"lazyas/internal/tui/panels"
)

// Mode represents the application mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirm
	ModeLoading
	ModeAddRepo
	ModeBackendSetup
	ModeStarterKit
	ModeUpdateResult
	ModeError
)

// ConfirmAction represents the action to confirm
type ConfirmAction int

const (
	ConfirmInstall ConfirmAction = iota
	ConfirmRemove
	ConfirmRemoveRepo
	ConfirmOverwrite
)

// App is the main TUI application model
type App struct {
	cfg      *config.Config
	registry *registry.Registry
	manifest *manifest.Manager

	// Layout
	layout *layout.PanelLayout

	// Panels
	skills *panels.SkillsPanel
	detail *panels.DetailPanel

	// Mode
	mode          Mode
	confirmAction ConfirmAction
	confirmSkill  *registry.SkillEntry
	confirmRepo   string // Repo name for removal confirmation
	confirmSel    int    // 0 = yes, 1 = no

	// Loading
	loadingMsg string
	spinnerIdx int

	// Add repo dialog
	addRepoName  textinput.Model
	addRepoURL   textinput.Model
	addRepoFocus int // 0 = name, 1 = url

	// Backend setup
	backendStatuses  []symlink.LinkStatus
	backendSelection []bool // Checkboxes for backend setup
	backendCursor    int    // Cursor in backend setup modal

	// Starter kit
	starterKitSelection []bool
	starterKitCursor    int

	// Update results
	updateResult *updateDoneMsg

	// Error modal
	errorTitle  string
	errorDetail string

	// Backend status for header
	linkedBackends int
	totalBackends  int

	// State
	message string
	err     error
	width   int
	height  int
	ready   bool

	// Styles
	styles AppStyles
}

// AppStyles holds application-wide styles
type AppStyles struct {
	Title        lipgloss.Style
	StatusBar    lipgloss.Style
	HelpKey      lipgloss.Style
	HelpText     lipgloss.Style
	ActivePanel  lipgloss.Style
	Panel        lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
	ConfirmBox   lipgloss.Style
	Button       lipgloss.Style
	ButtonActive lipgloss.Style
	Muted        lipgloss.Style
}

func defaultAppStyles() AppStyles {
	return AppStyles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		HelpKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")),
		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		ActivePanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true),
		ConfirmBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2),
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2),
		ButtonActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Bold(true).
			Padding(0, 2),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
	}
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	nameInput := textinput.New()
	nameInput.Placeholder = "repo-name"
	nameInput.CharLimit = 50

	urlInput := textinput.New()
	urlInput.Placeholder = "https://github.com/org/skills-repo"
	urlInput.CharLimit = 200

	return &App{
		cfg:         cfg,
		registry:    registry.NewRegistry(cfg),
		manifest:    manifest.NewManager(cfg),
		layout:      layout.NewPanelLayout(),
		mode:        ModeLoading,
		loadingMsg:  "Fetching skill index...",
		styles:      defaultAppStyles(),
		addRepoName: nameInput,
		addRepoURL:  urlInput,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.fetchIndex,
		tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
	)
}

// Messages
type (
	indexFetchedMsg  struct{}
	indexErrorMsg    struct{ err error }
	installDoneMsg   struct{ skill string }
	installErrMsg    struct{ err error }
	removeDoneMsg    struct{ skill string }
	removeErrMsg     struct{ err error }
	repoAddedMsg     struct{ name string }
	repoAddErrMsg    struct{ err error }
	repoRemovedMsg   struct{ name string }
	repoRemoveErrMsg struct{ err error }
	syncDoneMsg      struct{ skillCount int }
	syncErrMsg       struct{ err error }
	updateDoneMsg    struct {
		updated int
		skipped int
		failed  int
		results []updateSkillResult
	}
	updateErrMsg       struct{ err error }
	backendLinkDoneMsg struct{ linked int }
	backendLinkErrMsg  struct{ err error }
	starterKitDoneMsg  struct{ count int }
	starterKitErrMsg   struct{ err error }
	tickMsg            struct{}
)

type updateSkillResult struct {
	name   string
	status string // "updated", "skipped", "failed", "up-to-date"
}

func (a *App) fetchIndex() tea.Msg {
	return a.doFetchIndex(false)
}

func (a *App) fetchIndexForced() tea.Msg {
	return a.doFetchIndex(true)
}

func (a *App) doFetchIndex(force bool) tea.Msg {
	if err := a.manifest.Load(); err != nil {
		return indexErrorMsg{err}
	}

	if err := a.registry.Fetch(force); err != nil {
		return indexErrorMsg{err}
	}

	return indexFetchedMsg{}
}

func (a *App) initPanels() {
	// Scan for local skills
	localSkills := a.manifest.ScanLocalSkills()
	installed := make(map[string]bool)
	modified := make(map[string]bool)
	localOnly := make(map[string]bool)
	manifestInstalled := a.manifest.ListInstalled()
	for name, local := range localSkills {
		installed[name] = true
		if local.IsModified {
			modified[name] = true
		}
		if _, tracked := manifestInstalled[name]; !tracked {
			localOnly[name] = true
		}
	}

	// Merge registry skills with local-only skills
	skills := mergeSkills(a.registry.ListSkills(), localSkills)

	// Create panels
	a.skills = panels.NewSkillsPanel(skills, installed, modified)
	a.skills.SetLocalOnly(localOnly)
	a.skills.SetFocused(true)
	a.skills.SetSize(a.layout.LeftContentWidth(), a.layout.ContentHeight())

	a.detail = panels.NewDetailPanel()
	a.detail.SetFocused(false)
	a.detail.SetSize(a.layout.RightContentWidth(), a.layout.ContentHeight())

	// Update detail panel with selected skill
	a.updateDetailPanel()
}

func mergeSkills(registrySkills []registry.SkillEntry, localSkills map[string]manifest.LocalSkill) []registry.SkillEntry {
	seen := make(map[string]bool)
	result := make([]registry.SkillEntry, 0, len(registrySkills)+len(localSkills))

	for _, skill := range registrySkills {
		result = append(result, skill)
		seen[skill.Name] = true
	}

	for name, local := range localSkills {
		if !seen[name] {
			result = append(result, registry.SkillEntry{
				Name:        name,
				Description: local.Description,
				Source: registry.SkillSource{
					Repo: local.Path,
				},
			})
		}
	}

	return result
}

func (a *App) updateDetailPanel() {
	if a.skills == nil || a.detail == nil {
		return
	}

	skill := a.skills.Selected()
	if skill == nil {
		a.detail.SetSkill(nil, nil, nil, "")
		return
	}

	var installed *manifest.InstalledSkill
	if info, ok := a.manifest.GetInstalled(skill.Name); ok {
		installed = &info
	}

	localSkills := a.manifest.ScanLocalSkills()
	var local *manifest.LocalSkill
	if l, ok := localSkills[skill.Name]; ok {
		local = &l
	}

	a.detail.SetSkill(skill, installed, local, a.cfg.SkillsDir)
}

// checkBackendStatus updates the backend status for the header display
func (a *App) checkBackendStatus() {
	statuses := symlink.CheckBackendLinks(a.cfg.Backends, a.cfg.SkillsDir)
	a.totalBackends = 0
	a.linkedBackends = 0
	for _, s := range statuses {
		if s.Linked || s.Available {
			a.totalBackends++
		}
		if s.Linked {
			a.linkedBackends++
		}
	}
	a.backendStatuses = statuses
}

// Update handles all application events
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.layout.SetSize(msg.Width, msg.Height-4) // Reserve space for header and status bar
		if a.skills != nil {
			a.skills.SetSize(a.layout.LeftContentWidth(), a.layout.ContentHeight())
		}
		if a.detail != nil {
			a.detail.SetSize(a.layout.RightContentWidth(), a.layout.ContentHeight())
		}
		a.ready = true
		return a, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		switch a.mode {
		case ModeNormal:
			return a.updateNormal(msg)
		case ModeConfirm:
			return a.updateConfirm(msg)
		case ModeAddRepo:
			return a.updateAddRepo(msg)
		case ModeBackendSetup:
			return a.updateBackendSetup(msg)
		case ModeStarterKit:
			return a.updateStarterKit(msg)
		case ModeUpdateResult:
			return a.updateUpdateResult(msg)
		case ModeError:
			return a.updateError(msg)
		}

	case indexFetchedMsg:
		a.initPanels()
		a.checkBackendStatus()
		// Replace stale "refreshing..." message with completion summary
		if a.message != "" {
			a.message = a.styles.Success.Render(fmt.Sprintf("Done. %d skill(s) available.", len(a.registry.ListSkills())))
		}
		// Show backend setup modal if there are new available backends
		if symlink.HasNewBackends(a.backendStatuses, a.cfg.DismissedBackends) {
			a.mode = ModeBackendSetup
			a.initBackendSetup()
		} else if len(a.cfg.Repos) == 0 && !a.cfg.StarterKitDismissed {
			a.initStarterKit()
			a.mode = ModeStarterKit
		} else {
			a.mode = ModeNormal
		}
		return a, nil

	case indexErrorMsg:
		a.err = msg.err
		a.mode = ModeNormal
		// Initialize panels with local skills only
		a.initPanels()
		a.checkBackendStatus()
		return a, nil

	case installDoneMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Installed %s", msg.skill))
		a.refreshPanels()
		a.mode = ModeNormal
		return a, nil

	case installErrMsg:
		a.errorTitle = "Install Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case removeDoneMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Removed %s", msg.skill))
		a.refreshPanels()
		a.mode = ModeNormal
		return a, nil

	case removeErrMsg:
		a.errorTitle = "Remove Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case repoAddedMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Added repository '%s' - refreshing...", msg.name))
		a.err = nil
		// Refresh registry with new repo (force to bypass cache)
		a.registry = registry.NewRegistry(a.cfg)
		a.loadingMsg = "Fetching skill index..."
		a.mode = ModeLoading
		return a, tea.Batch(
			a.fetchIndexForced,
			tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
		)

	case repoAddErrMsg:
		a.errorTitle = "Add Repository Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case repoRemovedMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Removed repository '%s' - refreshing...", msg.name))
		a.err = nil
		// Refresh registry without removed repo (force to bypass cache)
		a.registry = registry.NewRegistry(a.cfg)
		a.loadingMsg = "Fetching skill index..."
		a.mode = ModeLoading
		return a, tea.Batch(
			a.fetchIndexForced,
			tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
		)

	case repoRemoveErrMsg:
		a.errorTitle = "Remove Repository Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case syncDoneMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Synced. %d skill(s) available.", msg.skillCount))
		a.refreshPanels()
		a.filterSkills()
		a.mode = ModeNormal
		return a, nil

	case syncErrMsg:
		a.errorTitle = "Sync Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case updateDoneMsg:
		a.updateResult = &msg
		a.refreshPanels()
		a.mode = ModeUpdateResult
		return a, nil

	case updateErrMsg:
		a.errorTitle = "Update Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case backendLinkDoneMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Linked %d backend(s)", msg.linked))
		a.checkBackendStatus()
		// Undismiss newly linked backends
		for _, s := range a.backendStatuses {
			if s.Linked {
				a.cfg.UndismissBackend(s.Backend.Name)
			}
		}
		a.cfg.Save()
		// Chain: show starter kit if no repos configured yet
		if len(a.cfg.Repos) == 0 && !a.cfg.StarterKitDismissed {
			a.initStarterKit()
			a.mode = ModeStarterKit
		} else {
			a.mode = ModeNormal
		}
		return a, nil

	case backendLinkErrMsg:
		a.errorTitle = "Backend Link Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case starterKitDoneMsg:
		a.message = a.styles.Success.Render(fmt.Sprintf("Added %d repository(ies) - refreshing...", msg.count))
		a.err = nil
		// Refresh registry with new repos (force to bypass cache)
		a.registry = registry.NewRegistry(a.cfg)
		a.loadingMsg = "Fetching skill index..."
		a.mode = ModeLoading
		return a, tea.Batch(
			a.fetchIndexForced,
			tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
		)

	case starterKitErrMsg:
		a.errorTitle = "Starter Kit Failed"
		a.errorDetail = msg.err.Error()
		a.mode = ModeError
		return a, nil

	case tickMsg:
		if a.mode == ModeLoading {
			a.spinnerIdx = (a.spinnerIdx + 1) % 4
			return a, tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} })
		}
	}

	return a, nil
}

func (a *App) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "q":
		if a.skills != nil && a.skills.IsSearching() {
			// Let the panel handle it
		} else {
			return a, tea.Quit
		}

	case "h", "left":
		if a.skills != nil && !a.skills.IsSearching() {
			a.layout.FocusLeft()
			a.skills.SetFocused(true)
			a.detail.SetFocused(false)
			return a, nil
		}

	case "l", "right":
		if a.skills != nil && !a.skills.IsSearching() {
			a.layout.FocusRight()
			a.skills.SetFocused(false)
			a.detail.SetFocused(true)
			return a, nil
		}

	case "i":
		if a.skills != nil && !a.skills.IsSearching() {
			if skill := a.skills.Selected(); skill != nil {
				onDisk := a.manifest.IsInstalled(skill.Name)
				if !onDisk {
					// Not on disk: install directly
					a.confirmSkill = skill
					a.loadingMsg = fmt.Sprintf("Installing %s...", skill.Name)
					a.mode = ModeLoading
					return a, tea.Batch(
						a.installSkill(skill),
						tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
					)
				}
				// Already on disk (tracked or untracked): confirm overwrite
				a.confirmAction = ConfirmOverwrite
				a.confirmSkill = skill
				a.confirmSel = 0
				a.mode = ModeConfirm
				return a, nil
			}
		}

	case "r":
		if a.skills != nil && !a.skills.IsSearching() {
			// Check if cursor is on a repo header
			if header := a.skills.SelectedHeader(); header != nil && header.RepoURL != "" {
				// Find matching repo in config
				for _, repo := range a.cfg.Repos {
					if repo.URL == header.RepoURL {
						a.confirmAction = ConfirmRemoveRepo
						a.confirmRepo = repo.Name
						a.confirmSel = 0
						a.mode = ModeConfirm
						return a, nil
					}
				}
			} else if skill := a.skills.Selected(); skill != nil {
				if a.manifest.IsInstalled(skill.Name) {
					a.confirmAction = ConfirmRemove
					a.confirmSkill = skill
					a.confirmSel = 0
					a.mode = ModeConfirm
					return a, nil
				}
			}
		}

	case "c":
		if a.skills != nil && !a.skills.IsSearching() {
			a.skills.ClearSearch()
			a.filterSkills()
			return a, nil
		}

	case "A":
		if a.skills != nil && !a.skills.IsSearching() {
			a.addRepoName.Reset()
			a.addRepoURL.Reset()
			a.addRepoFocus = 0
			a.addRepoName.Focus()
			a.mode = ModeAddRepo
			return a, textinput.Blink
		}

	case "b":
		if a.skills != nil && !a.skills.IsSearching() {
			a.checkBackendStatus()
			a.initBackendSetup()
			a.mode = ModeBackendSetup
			return a, nil
		}

	case "U":
		if a.skills != nil && !a.skills.IsSearching() {
			a.loadingMsg = "Updating skills..."
			a.mode = ModeLoading
			return a, tea.Batch(
				a.updateAllSkills(),
				tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
			)
		}

	case "S":
		if a.skills != nil && !a.skills.IsSearching() {
			a.loadingMsg = "Syncing repositories..."
			a.mode = ModeLoading
			return a, tea.Batch(
				a.syncRepos(),
				tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
			)
		}

	case "K":
		if a.skills != nil && !a.skills.IsSearching() {
			a.initStarterKit()
			a.mode = ModeStarterKit
			return a, nil
		}
	}

	// Route to focused panel
	var cmd tea.Cmd
	if a.layout.Focus() == layout.PanelLeft && a.skills != nil {
		prevSelected := a.skills.Selected()
		cmd = a.skills.Update(msg)

		// If search query changed, filter skills
		if key == "enter" && a.skills.GetQuery() != "" {
			a.filterSkills()
		}

		// Update detail if selection changed
		if a.skills.Selected() != prevSelected {
			a.updateDetailPanel()
		}
	} else if a.detail != nil {
		cmd = a.detail.Update(msg)
	}

	return a, cmd
}

func (a *App) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		a.confirmSel = 0
	case "right", "l":
		a.confirmSel = 1
	case "y", "Y":
		a.confirmSel = 0
		return a.executeConfirm()
	case "n", "N", "esc", "q":
		a.mode = ModeNormal
		return a, nil
	case "enter":
		return a.executeConfirm()
	}
	return a, nil
}

func (a *App) updateAddRepo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = ModeNormal
		return a, nil

	case "tab", "down":
		if a.addRepoFocus == 0 {
			a.addRepoFocus = 1
			a.addRepoName.Blur()
			a.addRepoURL.Focus()
		} else {
			a.addRepoFocus = 0
			a.addRepoURL.Blur()
			a.addRepoName.Focus()
		}
		return a, textinput.Blink

	case "shift+tab", "up":
		if a.addRepoFocus == 1 {
			a.addRepoFocus = 0
			a.addRepoURL.Blur()
			a.addRepoName.Focus()
		} else {
			a.addRepoFocus = 1
			a.addRepoName.Blur()
			a.addRepoURL.Focus()
		}
		return a, textinput.Blink

	case "enter":
		name := strings.TrimSpace(a.addRepoName.Value())
		url := strings.TrimSpace(a.addRepoURL.Value())

		if name == "" || url == "" {
			a.message = a.styles.Error.Render("Name and URL are required")
			return a, nil
		}

		// Add repo in background
		return a, func() tea.Msg {
			if err := a.cfg.AddRepo(name, url); err != nil {
				return repoAddErrMsg{err}
			}
			return repoAddedMsg{name}
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if a.addRepoFocus == 0 {
		a.addRepoName, cmd = a.addRepoName.Update(msg)
	} else {
		a.addRepoURL, cmd = a.addRepoURL.Update(msg)
	}
	return a, cmd
}

// Backend setup modal handling
func (a *App) initBackendSetup() {
	a.backendSelection = make([]bool, len(a.backendStatuses))
	// Pre-select available+unlinked backends
	for i, s := range a.backendStatuses {
		a.backendSelection[i] = s.Available && !s.Linked && s.Error == nil
	}
	a.backendCursor = 0
}

func (a *App) updateBackendSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Dismiss all available+unlinked backends so modal doesn't re-appear
		for _, s := range a.backendStatuses {
			if s.Available && !s.Linked && s.Error == nil {
				a.cfg.DismissBackend(s.Backend.Name)
			}
		}
		a.cfg.Save()
		// Chain: show starter kit if no repos configured yet
		if len(a.cfg.Repos) == 0 && !a.cfg.StarterKitDismissed {
			a.initStarterKit()
			a.mode = ModeStarterKit
		} else {
			a.mode = ModeNormal
		}
		return a, nil

	case "j", "down":
		if a.backendCursor < len(a.backendStatuses)-1 {
			a.backendCursor++
		}
		return a, nil

	case "k", "up":
		if a.backendCursor > 0 {
			a.backendCursor--
		}
		return a, nil

	case " ", "x":
		// Toggle selection (only for available+unlinked backends)
		if a.backendCursor < len(a.backendStatuses) {
			s := a.backendStatuses[a.backendCursor]
			if s.Available && !s.Linked {
				a.backendSelection[a.backendCursor] = !a.backendSelection[a.backendCursor]
			}
		}
		return a, nil

	case "enter":
		// Link selected backends
		var toLink []symlink.LinkStatus
		for i, sel := range a.backendSelection {
			if sel && !a.backendStatuses[i].Linked {
				toLink = append(toLink, a.backendStatuses[i])
			}
		}

		if len(toLink) == 0 {
			// Chain: show starter kit if no repos configured yet
			if len(a.cfg.Repos) == 0 && !a.cfg.StarterKitDismissed {
				a.initStarterKit()
				a.mode = ModeStarterKit
			} else {
				a.mode = ModeNormal
			}
			return a, nil
		}

		a.loadingMsg = "Linking backends..."
		a.mode = ModeLoading
		return a, tea.Batch(
			a.linkBackends(toLink),
			tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
		)
	}

	return a, nil
}

// Update result modal handling
func (a *App) updateUpdateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		a.mode = ModeNormal
		a.updateResult = nil
		return a, nil
	}
	return a, nil
}

// Error modal handling
func (a *App) updateError(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		a.mode = ModeNormal
		a.errorTitle = ""
		a.errorDetail = ""
		return a, nil
	}
	return a, nil
}

func (a *App) executeConfirm() (tea.Model, tea.Cmd) {
	if a.confirmSel == 1 {
		a.mode = ModeNormal
		return a, nil
	}

	switch a.confirmAction {
	case ConfirmInstall:
		a.loadingMsg = fmt.Sprintf("Installing %s...", a.confirmSkill.Name)
		a.mode = ModeLoading
		return a, a.installSkill(a.confirmSkill)
	case ConfirmRemove:
		a.loadingMsg = fmt.Sprintf("Removing %s...", a.confirmSkill.Name)
		a.mode = ModeLoading
		return a, a.removeSkill(a.confirmSkill)
	case ConfirmRemoveRepo:
		repoName := a.confirmRepo
		a.loadingMsg = "Removing repository..."
		a.mode = ModeLoading
		return a, a.removeRepo(repoName)
	case ConfirmOverwrite:
		a.loadingMsg = fmt.Sprintf("Installing %s...", a.confirmSkill.Name)
		a.mode = ModeLoading
		return a, tea.Batch(
			a.overwriteAndInstall(a.confirmSkill),
			tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
		)
	}
	return a, nil
}

func (a *App) filterSkills() {
	if a.skills == nil {
		return
	}

	localSkills := a.manifest.ScanLocalSkills()
	query := a.skills.GetQuery()

	var skills []registry.SkillEntry
	if query == "" {
		skills = mergeSkills(a.registry.ListSkills(), localSkills)
	} else {
		skills = mergeSkills(a.registry.SearchSkills(query), localSkills)
	}
	a.skills.SetSkills(skills)
	a.updateDetailPanel()
}

func (a *App) refreshPanels() {
	localSkills := a.manifest.ScanLocalSkills()
	installed := make(map[string]bool)
	modified := make(map[string]bool)
	localOnly := make(map[string]bool)
	manifestInstalled := a.manifest.ListInstalled()
	for name, local := range localSkills {
		installed[name] = true
		if local.IsModified {
			modified[name] = true
		}
		if _, tracked := manifestInstalled[name]; !tracked {
			localOnly[name] = true
		}
	}
	a.skills.SetInstalled(installed)
	a.skills.SetModified(modified)
	a.skills.SetLocalOnly(localOnly)
	a.updateDetailPanel()
}

func (a *App) installSkill(skill *registry.SkillEntry) tea.Cmd {
	return func() tea.Msg {
		targetDir := a.manifest.GetSkillPath(skill.Name)

		result, err := git.Clone(git.CloneOptions{
			Repo:      skill.Source.Repo,
			Path:      skill.Source.Path,
			Tag:       skill.Source.Tag,
			TargetDir: targetDir,
		})
		if err != nil {
			return installErrMsg{err}
		}

		if err := git.ValidateSkill(targetDir); err != nil {
			os.RemoveAll(targetDir)
			return installErrMsg{err}
		}

		if err := a.manifest.AddSkill(
			skill.Name,
			skill.Source.Tag,
			result.Commit,
			skill.Source.Repo,
			skill.Source.Path,
		); err != nil {
			return installErrMsg{err}
		}

		return installDoneMsg{skill.Name}
	}
}

func (a *App) overwriteAndInstall(skill *registry.SkillEntry) tea.Cmd {
	return func() tea.Msg {
		targetDir := a.manifest.GetSkillPath(skill.Name)
		// Remove the existing local copy
		os.RemoveAll(targetDir)
		// Install from registry
		result, err := git.Clone(git.CloneOptions{
			Repo:      skill.Source.Repo,
			Path:      skill.Source.Path,
			Tag:       skill.Source.Tag,
			TargetDir: targetDir,
		})
		if err != nil {
			return installErrMsg{err}
		}
		if err := git.ValidateSkill(targetDir); err != nil {
			os.RemoveAll(targetDir)
			return installErrMsg{err}
		}
		if err := a.manifest.AddSkill(
			skill.Name,
			skill.Source.Tag,
			result.Commit,
			skill.Source.Repo,
			skill.Source.Path,
		); err != nil {
			return installErrMsg{err}
		}
		return installDoneMsg{skill.Name}
	}
}

func (a *App) removeSkill(skill *registry.SkillEntry) tea.Cmd {
	return func() tea.Msg {
		skillDir := a.manifest.GetSkillPath(skill.Name)

		if err := os.RemoveAll(skillDir); err != nil {
			return removeErrMsg{err}
		}

		if err := a.manifest.RemoveSkill(skill.Name); err != nil {
			return removeErrMsg{err}
		}

		return removeDoneMsg{skill.Name}
	}
}

func (a *App) syncRepos() tea.Cmd {
	return func() tea.Msg {
		if err := a.registry.Fetch(true); err != nil {
			return syncErrMsg{err}
		}
		return syncDoneMsg{len(a.registry.ListSkills())}
	}
}

func (a *App) removeRepo(name string) tea.Cmd {
	return func() tea.Msg {
		if err := a.cfg.RemoveRepo(name); err != nil {
			return repoRemoveErrMsg{err}
		}
		return repoRemovedMsg{name}
	}
}

func (a *App) updateAllSkills() tea.Cmd {
	return func() tea.Msg {
		// Get installed skills
		installed := a.manifest.ListInstalled()
		if len(installed) == 0 {
			return updateDoneMsg{0, 0, 0, nil}
		}

		// Force refresh registry first
		a.registry.Fetch(true)

		var updated, skipped, failed int
		var results []updateSkillResult

		for name, info := range installed {
			skillPath := a.manifest.GetSkillPath(name)

			// Check for modifications
			modified, _ := git.IsModified(skillPath)
			if modified {
				results = append(results, updateSkillResult{name, "skipped"})
				skipped++
				continue
			}

			// Check if update available
			skill := a.registry.GetSkill(name)

			// Determine target version
			targetTag := ""
			if skill != nil {
				targetTag = skill.Source.Tag
			}

			// Check if it's a sparse checkout
			gitDir := skillPath + "/.git"
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				if skill == nil {
					results = append(results, updateSkillResult{name, "skipped"})
					skipped++
					continue
				}

				// Remove and re-clone for sparse checkouts
				os.RemoveAll(skillPath)
				result, err := git.Clone(git.CloneOptions{
					Repo:      skill.Source.Repo,
					Path:      skill.Source.Path,
					Tag:       targetTag,
					TargetDir: skillPath,
				})
				if err != nil {
					results = append(results, updateSkillResult{name, "failed"})
					failed++
					continue
				}

				a.manifest.AddSkill(name, targetTag, result.Commit, skill.Source.Repo, skill.Source.Path)
				results = append(results, updateSkillResult{name, "updated"})
				updated++
			} else {
				// Regular git update
				result, err := git.Update(skillPath, targetTag)
				if err != nil {
					results = append(results, updateSkillResult{name, "failed"})
					failed++
					continue
				}

				if result.Commit != info.Commit {
					sourceRepo := info.SourceRepo
					sourcePath := info.SourcePath
					if skill != nil {
						sourceRepo = skill.Source.Repo
						sourcePath = skill.Source.Path
					}
					a.manifest.AddSkill(name, targetTag, result.Commit, sourceRepo, sourcePath)
					results = append(results, updateSkillResult{name, "updated"})
					updated++
				} else {
					results = append(results, updateSkillResult{name, "up-to-date"})
					skipped++
				}
			}
		}

		return updateDoneMsg{updated, skipped, failed, results}
	}
}

func (a *App) linkBackends(toLink []symlink.LinkStatus) tea.Cmd {
	return func() tea.Msg {
		linked := 0
		for _, s := range toLink {
			if s.HasFiles && !s.IsSymlink {
				// Migrate existing directory
				if err := symlink.MigrateExistingDir(s.Backend, a.cfg.SkillsDir); err != nil {
					return backendLinkErrMsg{fmt.Errorf("failed to migrate %s: %w", s.Backend.Name, err)}
				}
			} else {
				// Create symlink
				if err := symlink.CreateLink(s.Backend, a.cfg.SkillsDir); err != nil {
					return backendLinkErrMsg{fmt.Errorf("failed to link %s: %w", s.Backend.Name, err)}
				}
			}
			linked++
		}
		return backendLinkDoneMsg{linked}
	}
}

// View renders the application
func (a *App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Title with backend status
	b.WriteString(a.styles.Title.Render("lazyas"))
	b.WriteString("  ")
	b.WriteString(a.styles.StatusBar.Render("Lazy Agent Skills"))

	// Backend status in header
	if a.totalBackends > 0 {
		b.WriteString("  ")
		backendInfo := a.renderBackendStatusHeader()
		b.WriteString(backendInfo)
	}
	b.WriteString("\n\n")

	switch a.mode {
	case ModeLoading:
		if a.skills != nil && a.detail != nil {
			b.WriteString(a.overlayModal(a.renderPanels(), a.renderLoadingContent()))
		} else {
			b.WriteString(a.renderLoading())
		}
	case ModeNormal:
		b.WriteString(a.renderPanels())
	case ModeConfirm:
		b.WriteString(a.overlayModal(a.renderPanels(), a.renderConfirmContent()))
	case ModeAddRepo:
		b.WriteString(a.overlayModal(a.renderPanels(), a.renderAddRepoContent()))
	case ModeBackendSetup:
		b.WriteString(a.overlayModal(a.renderPanels(), a.renderBackendSetupContent()))
	case ModeStarterKit:
		b.WriteString(a.overlayModal(a.renderPanels(), a.renderStarterKitContent()))
	case ModeUpdateResult:
		b.WriteString(a.overlayModal(a.renderPanels(), a.renderUpdateResultContent()))
	case ModeError:
		b.WriteString(a.overlayModal(a.renderPanels(), a.renderErrorContent()))
	}

	// Error or message (always reserve the line to prevent layout jumps)
	b.WriteString("\n")
	if a.err != nil {
		b.WriteString(a.styles.Error.Render(fmt.Sprintf("Error: %v", a.err)))
	} else if a.message != "" {
		b.WriteString(a.message)
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(a.renderStatusBar())

	return b.String()
}

func (a *App) renderBackendStatusHeader() string {
	var parts []string
	for _, s := range a.backendStatuses {
		if s.Linked {
			parts = append(parts, a.styles.Success.Render(s.Backend.Name+" ✓"))
		} else if s.Available {
			parts = append(parts, a.styles.Muted.Render(s.Backend.Name+" ○"))
		}
		// Skip backends that are neither linked nor available
	}
	return strings.Join(parts, " ")
}

func (a *App) renderLoading() string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸"}
	return fmt.Sprintf("%s %s", spinners[a.spinnerIdx%len(spinners)], a.loadingMsg)
}

func (a *App) renderLoadingContent() string {
	modalBg := lipgloss.Color("#1a1a2e")
	contentWidth := 40
	if len(a.loadingMsg)+6 > contentWidth {
		contentWidth = len(a.loadingMsg) + 6
	}

	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	spinners := []string{"⠋", "⠙", "⠹", "⠸"}
	spinner := spinners[a.spinnerIdx%len(spinners)]
	line := fmt.Sprintf("  %s %s", spinner, a.loadingMsg)

	return lipgloss.JoinVertical(lipgloss.Left,
		lineBg.Render(""),
		lineBg.Render(line),
		lineBg.Render(""),
	)
}

func (a *App) renderPanels() string {
	if a.skills == nil || a.detail == nil {
		return ""
	}

	// Left panel
	leftStyle := a.styles.Panel
	if a.layout.Focus() == layout.PanelLeft {
		leftStyle = a.styles.ActivePanel
	}
	leftContent := a.skills.View()
	leftPanel := leftStyle.
		Width(a.layout.LeftWidth() - 2).
		Height(a.layout.ContentHeight()).
		Render(leftContent)

	// Right panel
	rightStyle := a.styles.Panel
	if a.layout.Focus() == layout.PanelRight {
		rightStyle = a.styles.ActivePanel
	}
	rightContent := a.detail.View()
	rightPanel := rightStyle.
		Width(a.layout.RightWidth() - 2).
		Height(a.layout.ContentHeight()).
		Render(rightContent)

	// Join panels horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)
}

func (a *App) overlayModal(background, modalContent string) string {
	// Create modal box with solid background
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(1, 2)

	modal := modalStyle.Render(modalContent)

	// Get dimensions
	contentHeight := a.height - 5 // Account for header and status bar

	// Calculate modal dimensions
	modalWidth := lipgloss.Width(modal)
	modalHeight := lipgloss.Height(modal)

	// Calculate centered position
	startX := (a.width - modalWidth) / 2
	startY := (contentHeight - modalHeight) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Split into lines
	bgLines := strings.Split(background, "\n")
	fgLines := strings.Split(modal, "\n")

	// Ensure background has enough lines
	for len(bgLines) < contentHeight {
		bgLines = append(bgLines, "")
	}

	// Overlay modal onto background using ANSI-aware manipulation
	for i, fgLine := range fgLines {
		bgIdx := startY + i
		if bgIdx >= 0 && bgIdx < len(bgLines) {
			bgLines[bgIdx] = overlayLine(bgLines[bgIdx], fgLine, startX, a.width)
		}
	}

	return strings.Join(bgLines, "\n")
}

// overlayLine overlays a foreground string onto a background string at the given position
func overlayLine(bg, fg string, startX, totalWidth int) string {
	// Get the visible prefix from background (before modal starts)
	prefix := ansi.Truncate(bg, startX, "")

	// Pad prefix if background line is too short
	prefixWidth := ansi.StringWidth(prefix)
	if prefixWidth < startX {
		prefix += strings.Repeat(" ", startX-prefixWidth)
	}

	// Get foreground width and calculate where suffix should start
	fgWidth := ansi.StringWidth(fg)
	suffixStart := startX + fgWidth

	// Pad the area to the right of the modal with spaces.
	// Attempting to recover the background content here is fragile because
	// ANSI codes and multi-byte border characters get corrupted by truncation.
	var suffix string
	if suffixStart < totalWidth {
		suffix = strings.Repeat(" ", totalWidth-suffixStart)
	}

	// Reset ANSI at transitions to prevent color bleed
	reset := "\033[0m"
	return prefix + reset + fg + reset + suffix
}

func (a *App) renderConfirmContent() string {
	var title, message string
	switch a.confirmAction {
	case ConfirmInstall:
		title = "Install Skill"
		message = fmt.Sprintf("Install %s?", a.confirmSkill.Name)
	case ConfirmRemove:
		title = "Remove Skill"
		message = fmt.Sprintf("Remove %s?", a.confirmSkill.Name)
	case ConfirmRemoveRepo:
		title = "Remove Repository"
		message = fmt.Sprintf("Remove repo '%s'?", a.confirmRepo)
	case ConfirmOverwrite:
		title = "Install from Registry"
		message = fmt.Sprintf("Replace local %s with registry version?", a.confirmSkill.Name)
	}

	// Modal background color for consistent styling
	modalBg := lipgloss.Color("#1a1a2e")

	yesBtn := a.styles.Button.Background(modalBg).Render(" Yes ")
	noBtn := a.styles.Button.Background(modalBg).Render(" No ")
	if a.confirmSel == 0 {
		yesBtn = a.styles.ButtonActive.Render(" Yes ")
		noBtn = a.styles.Button.Background(modalBg).Render(" No ")
	} else {
		yesBtn = a.styles.Button.Background(modalBg).Render(" Yes ")
		noBtn = a.styles.ButtonActive.Render(" No ")
	}

	// Calculate content width for consistent background
	contentWidth := 30
	if len(message) > contentWidth {
		contentWidth = len(message) + 4
	}

	// Style for consistent background on all lines
	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	titleStyled := a.styles.Title.Background(modalBg).Width(contentWidth).Render(title)
	messageStyled := lineBg.Render(message)
	emptyLine := lineBg.Render("")
	spacer := lipgloss.NewStyle().Background(modalBg).Render("  ")
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yesBtn, spacer, noBtn)
	buttonsStyled := lineBg.Render(buttons)

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyled,
		emptyLine,
		messageStyled,
		emptyLine,
		buttonsStyled,
	)
}

func (a *App) renderAddRepoContent() string {
	// Set input widths
	a.addRepoName.Width = 50
	a.addRepoURL.Width = 50

	// Modal background color for consistent styling
	modalBg := lipgloss.Color("#1a1a2e")
	contentWidth := 70

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Background(modalBg).
		Width(8)

	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	var nameIndicator, urlIndicator string
	if a.addRepoFocus == 0 {
		nameIndicator = a.styles.Title.Background(modalBg).Render("> ")
		urlIndicator = lipgloss.NewStyle().Background(modalBg).Render("  ")
	} else {
		nameIndicator = lipgloss.NewStyle().Background(modalBg).Render("  ")
		urlIndicator = a.styles.Title.Background(modalBg).Render("> ")
	}

	titleStyled := a.styles.Title.Background(modalBg).Width(contentWidth).Render("Add Repository")
	descStyled := lineBg.Render("Add a skills repository to fetch skills from.")
	emptyLine := lineBg.Render("")
	nameRow := lineBg.Render(lipgloss.JoinHorizontal(lipgloss.Top, nameIndicator, labelStyle.Render("Name"), a.addRepoName.View()))
	urlRow := lineBg.Render(lipgloss.JoinHorizontal(lipgloss.Top, urlIndicator, labelStyle.Render("URL"), a.addRepoURL.View()))
	helpStyled := a.styles.Muted.Background(modalBg).Width(contentWidth).Render("tab: next    enter: add    esc: cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyled,
		emptyLine,
		descStyled,
		emptyLine,
		nameRow,
		emptyLine,
		urlRow,
		emptyLine,
		helpStyled,
	)
}

func (a *App) renderBackendSetupContent() string {
	modalBg := lipgloss.Color("#1a1a2e")
	contentWidth := 50

	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	titleStyled := a.styles.Title.Background(modalBg).Width(contentWidth).Render("Backend Setup")
	emptyLine := lineBg.Render("")
	descStyled := lineBg.Render("lazyas manages skills in ~/.lazyas/skills/")
	desc2Styled := lineBg.Render("Select backends to link:")

	var lines []string
	lines = append(lines, titleStyled, emptyLine, descStyled, desc2Styled, emptyLine)

	for i, s := range a.backendStatuses {
		expandedPath, _ := config.ExpandPath(s.Backend.Path)
		selected := i == a.backendCursor

		var line string
		var suffix string

		if s.Linked {
			line = fmt.Sprintf("  [ ] %s (%s)", s.Backend.Name, expandedPath)
			suffix = " ✓ linked"
		} else if s.Error != nil {
			line = fmt.Sprintf("  [ ] %s (%s)", s.Backend.Name, expandedPath)
			suffix = " ✗ error"
		} else if !s.Available {
			// Unavailable backend - not installed on this system
			line = fmt.Sprintf("  [ ] %s", s.Backend.Name)
			suffix = " not installed"
		} else {
			checkbox := " "
			if a.backendSelection[i] {
				checkbox = "x"
			}
			line = fmt.Sprintf("  [%s] %s (%s)", checkbox, s.Backend.Name, expandedPath)
			if s.HasFiles {
				suffix = " (has files)"
			}
		}

		if selected && !s.Available && !s.Linked && s.Error == nil {
			// Dim highlight for unavailable backends
			dimCursorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Foreground(lipgloss.Color("#6B7280")).
				Width(contentWidth)
			lines = append(lines, dimCursorStyle.Render(line+suffix))
		} else if selected {
			// Render entire line uniformly with cursor highlight
			cursorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Width(contentWidth).
				Bold(true)
			lines = append(lines, cursorStyle.Render(line+suffix))
		} else if !s.Available && !s.Linked && s.Error == nil {
			// Muted/gray for unavailable backends
			mutedLine := a.styles.Muted.Render(line + suffix)
			lines = append(lines, lineBg.Render(mutedLine))
		} else {
			// Render label + styled suffix separately on modal background
			var styledLine string
			if s.Linked {
				styledLine = line + a.styles.Success.Render(suffix)
			} else if s.Error != nil {
				styledLine = line + a.styles.Error.Render(suffix)
			} else {
				styledLine = line + suffix
			}
			lines = append(lines, lineBg.Render(styledLine))
		}
	}

	lines = append(lines, emptyLine)
	helpStyled := a.styles.Muted.Background(modalBg).Width(contentWidth).Render("space: toggle  enter: link  esc: skip")
	lines = append(lines, helpStyled)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (a *App) renderUpdateResultContent() string {
	if a.updateResult == nil {
		return ""
	}

	modalBg := lipgloss.Color("#1a1a2e")
	contentWidth := 45

	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	titleStyled := a.styles.Title.Background(modalBg).Width(contentWidth).Render("Update Skills")
	emptyLine := lineBg.Render("")

	var lines []string
	lines = append(lines, titleStyled, emptyLine)

	for _, r := range a.updateResult.results {
		var statusIcon string
		switch r.status {
		case "updated":
			statusIcon = a.styles.Success.Background(modalBg).Render("✓ updated")
		case "up-to-date":
			statusIcon = a.styles.Muted.Background(modalBg).Render("  up to date")
		case "skipped":
			statusIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Background(modalBg).Render("⚠ local changes")
		case "failed":
			statusIcon = a.styles.Error.Background(modalBg).Render("✗ failed")
		}
		line := fmt.Sprintf("  %-20s %s", r.name, statusIcon)
		lines = append(lines, lineBg.Render(line))
	}

	lines = append(lines, emptyLine)

	summary := fmt.Sprintf("Updated: %d  Skipped: %d  Failed: %d",
		a.updateResult.updated, a.updateResult.skipped, a.updateResult.failed)
	lines = append(lines, lineBg.Render(summary))
	lines = append(lines, emptyLine)

	helpStyled := a.styles.Muted.Background(modalBg).Width(contentWidth).Render("enter/esc: close")
	lines = append(lines, helpStyled)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (a *App) renderErrorContent() string {
	modalBg := lipgloss.Color("#1a1a2e")
	contentWidth := 60

	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	titleStyled := a.styles.Error.Background(modalBg).Width(contentWidth).Render(a.errorTitle)
	emptyLine := lineBg.Render("")

	var lines []string
	lines = append(lines, titleStyled, emptyLine)

	// Split error detail into lines, truncate to fit modal width
	for _, detailLine := range strings.Split(a.errorDetail, "\n") {
		detailLine = strings.TrimRight(detailLine, " \t\r")
		if len(detailLine) > contentWidth-2 {
			detailLine = detailLine[:contentWidth-5] + "..."
		}
		if detailLine == "" {
			continue
		}
		styled := a.styles.Muted.Background(modalBg).Render("  " + detailLine)
		lines = append(lines, lineBg.Render(styled))
	}

	lines = append(lines, emptyLine)
	helpStyled := a.styles.Muted.Background(modalBg).Width(contentWidth).Render("enter/esc: close")
	lines = append(lines, helpStyled)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Starter kit modal handling
func (a *App) initStarterKit() {
	// Remove starter kit repos from config that yielded no skills
	a.pruneDeadStarterKitRepos()

	a.starterKitSelection = make([]bool, len(config.StarterKitRepos))
	// Pre-select repos not already in config
	for i, repo := range config.StarterKitRepos {
		a.starterKitSelection[i] = !a.hasRepo(repo.Name, repo.URL)
	}
	a.starterKitCursor = 0
}

// pruneDeadStarterKitRepos removes starter kit repos from config that have
// no skills in the registry (i.e. their fetch failed).
func (a *App) pruneDeadStarterKitRepos() {
	if a.registry == nil {
		return
	}

	// Build set of repo URLs that actually produced skills
	activeURLs := make(map[string]bool)
	for _, s := range a.registry.ListSkills() {
		if s.Source.Repo != "" {
			activeURLs[s.Source.Repo] = true
		}
	}

	// Check which starter kit repos are dead
	starterKitURLs := make(map[string]bool)
	for _, sk := range config.StarterKitRepos {
		starterKitURLs[sk.URL] = true
	}

	var valid []config.Repo
	changed := false
	for _, r := range a.cfg.Repos {
		if starterKitURLs[r.URL] && !activeURLs[r.URL] {
			// Starter kit repo with no skills — drop it
			changed = true
			continue
		}
		valid = append(valid, r)
	}

	if changed {
		a.cfg.Repos = valid
		a.cfg.Save()
	}
}

// hasRepo checks if a repo is already configured (by name or URL)
func (a *App) hasRepo(name, url string) bool {
	for _, r := range a.cfg.Repos {
		if r.Name == name || r.URL == url {
			return true
		}
	}
	return false
}

func (a *App) updateStarterKit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		a.cfg.StarterKitDismissed = true
		a.cfg.Save()
		a.mode = ModeNormal
		return a, nil

	case "j", "down":
		if a.starterKitCursor < len(config.StarterKitRepos)-1 {
			a.starterKitCursor++
		}
		return a, nil

	case "k", "up":
		if a.starterKitCursor > 0 {
			a.starterKitCursor--
		}
		return a, nil

	case " ", "x":
		if a.starterKitCursor < len(config.StarterKitRepos) {
			repo := config.StarterKitRepos[a.starterKitCursor]
			if !a.hasRepo(repo.Name, repo.URL) {
				a.starterKitSelection[a.starterKitCursor] = !a.starterKitSelection[a.starterKitCursor]
			}
		}
		return a, nil

	case "enter":
		var selected []config.Repo
		for i, sel := range a.starterKitSelection {
			if sel {
				selected = append(selected, config.StarterKitRepos[i])
			}
		}

		a.cfg.StarterKitDismissed = true
		a.cfg.Save()

		if len(selected) == 0 {
			a.mode = ModeNormal
			return a, nil
		}

		a.loadingMsg = "Adding repositories..."
		a.mode = ModeLoading
		return a, tea.Batch(
			a.addStarterKitRepos(selected),
			tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} }),
		)
	}

	return a, nil
}

func (a *App) addStarterKitRepos(repos []config.Repo) tea.Cmd {
	return func() tea.Msg {
		for _, r := range repos {
			if err := a.cfg.AddRepo(r.Name, r.URL); err != nil {
				return starterKitErrMsg{fmt.Errorf("failed to add %s: %w", r.Name, err)}
			}
		}
		return starterKitDoneMsg{len(repos)}
	}
}

func (a *App) renderStarterKitContent() string {
	modalBg := lipgloss.Color("#1a1a2e")
	contentWidth := 60

	lineBg := lipgloss.NewStyle().
		Background(modalBg).
		Width(contentWidth)

	titleStyled := a.styles.Title.Background(modalBg).Width(contentWidth).Render("Starter Kit Repositories")
	emptyLine := lineBg.Render("")
	descStyled := lineBg.Render("Add popular skill repositories to get started.")

	var lines []string
	lines = append(lines, titleStyled, emptyLine, descStyled, emptyLine)

	for i, repo := range config.StarterKitRepos {
		selected := i == a.starterKitCursor
		alreadyAdded := a.hasRepo(repo.Name, repo.URL)

		var line string
		var suffix string

		if alreadyAdded {
			line = fmt.Sprintf("  [ ] %s", repo.Name)
			suffix = "  ✓ added"
		} else {
			checkbox := " "
			if a.starterKitSelection[i] {
				checkbox = "x"
			}
			line = fmt.Sprintf("  [%s] %s", checkbox, repo.Name)
		}

		if selected && alreadyAdded {
			dimCursorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Foreground(lipgloss.Color("#6B7280")).
				Width(contentWidth)
			lines = append(lines, dimCursorStyle.Render(line+suffix))
		} else if selected {
			cursorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Width(contentWidth).
				Bold(true)
			lines = append(lines, cursorStyle.Render(line))
		} else if alreadyAdded {
			styledLine := a.styles.Muted.Render(line) + a.styles.Success.Render(suffix)
			lines = append(lines, lineBg.Render(styledLine))
		} else {
			lines = append(lines, lineBg.Render(line))
		}
	}

	lines = append(lines, emptyLine)
	helpStyled := a.styles.Muted.Background(modalBg).Width(contentWidth).Render("space: toggle  enter: add  esc: skip")
	lines = append(lines, helpStyled)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (a *App) renderStatusBar() string {
	var pairs []string

	if a.skills != nil && a.skills.IsSearching() {
		pairs = []string{
			"enter", "search",
			"esc", "cancel",
		}
	} else if a.mode == ModeConfirm {
		pairs = []string{
			"y", "yes",
			"n", "no",
			"←/→", "select",
			"enter", "confirm",
		}
	} else if a.mode == ModeAddRepo {
		pairs = []string{
			"tab", "next field",
			"enter", "add",
			"esc", "cancel",
		}
	} else if a.mode == ModeBackendSetup {
		pairs = []string{
			"j/k", "navigate",
			"space", "toggle",
			"enter", "link",
			"esc", "skip",
		}
	} else if a.mode == ModeStarterKit {
		pairs = []string{
			"j/k", "navigate",
			"space", "toggle",
			"enter", "add",
			"esc", "skip",
		}
	} else if a.mode == ModeUpdateResult || a.mode == ModeError {
		pairs = []string{
			"enter", "close",
			"esc", "close",
		}
	} else {
		pairs = []string{
			"j/k", "navigate",
			"h/l", "panels",
			"z", "fold",
			"i", "install",
			"r", "remove",
			"U", "update",
			"A", "add repo",
			"S", "sync",
			"b", "backends",
			"K", "starter kit",
			"/", "search",
			"q", "quit",
		}
	}

	var items []string
	for i := 0; i < len(pairs); i += 2 {
		items = append(items,
			a.styles.HelpKey.Render(pairs[i])+" "+a.styles.HelpText.Render(pairs[i+1]))
	}

	return a.styles.StatusBar.Render(strings.Join(items, "  "))
}

// Run starts the TUI application
func Run(cfg *config.Config) error {
	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	app := NewApp(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	// Check if the app stored an error (e.g., index fetch failure)
	if finalApp, ok := model.(*App); ok && finalApp.err != nil {
		return finalApp.err
	}
	return nil
}
