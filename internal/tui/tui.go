package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/damianoneill/claude-desktop-config/internal/config"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")) // bright blue

	styleSubtle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	styleEnabled = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true) // bright green

	styleDisabled = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // dim

	stylePending = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true) // yellow — staged but not yet saved

	styleCursorBar = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Bold(true)

	styleURL = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")) // blue

	styleStatusOK = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	styleStatusErr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	styleStatusInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12"))

	styleKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)

	styleDivider = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))
)

// ── Model ─────────────────────────────────────────────────────────────────────

// serverRow is one entry in the list.
type serverRow struct {
	name    string
	server  config.MCPServer
	enabled bool // current effective state (original + pending)
}

// Model is the Bubble Tea model for the TUI.
type Model struct {
	sourceFile  string
	src         *config.SourceConfig
	keepBackups int

	rows    []serverRow
	cursor  int
	pending map[string]bool // name → new enabled state (staged, not yet saved)

	status    string // status bar message
	statusErr bool   // true = render status in red

	width  int
	height int

	quitting bool
}

// New creates a Model loaded from sourceFile.
func New(sourceFile string, keepBackups int) (*Model, error) {
	src, err := config.Load(sourceFile)
	if err != nil {
		return nil, err
	}
	m := &Model{
		sourceFile:  sourceFile,
		src:         src,
		keepBackups: keepBackups,
		pending:     make(map[string]bool),
	}
	m.buildRows()
	return m, nil
}

// buildRows (re)builds the sorted row slice from src + pending.
func (m *Model) buildRows() {
	names := make([]string, 0, len(m.src.MCPServers))
	for n := range m.src.MCPServers {
		names = append(names, n)
	}
	sort.Strings(names)

	m.rows = make([]serverRow, 0, len(names))
	for _, name := range names {
		srv := m.src.MCPServers[name]
		enabled := config.IsEnabled(srv)
		if v, ok := m.pending[name]; ok {
			enabled = v
		}
		m.rows = append(m.rows, serverRow{name: name, server: srv, enabled: enabled})
	}
	m.showComment()
}

// showComment sets the status bar to the _comment of the currently selected row,
// or clears it if the row has no comment.
func (m *Model) showComment() {
	if len(m.rows) == 0 {
		return
	}
	comment := m.rows[m.cursor].server.Comment
	if comment != "" {
		m.status = comment
		m.statusErr = false
	} else {
		m.status = ""
	}
}

// pendingCount returns the number of staged (unsaved) changes.
func (m *Model) pendingCount() int { return len(m.pending) }

// enabledCount returns the current effective enabled count (orig + pending).
func (m *Model) enabledCount() int {
	n := 0
	for _, r := range m.rows {
		if r.enabled {
			n++
		}
	}
	return n
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return nil
}

// ── Messages ──────────────────────────────────────────────────────────────────

type statusMsg struct {
	text string
	err  bool
}

type savedMsg struct{}
type appliedMsg struct{ destPath string }

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case statusMsg:
		m.status = msg.text
		m.statusErr = msg.err

	case savedMsg:
		// pending flushed to disk — clear pending map and reload
		m.pending = make(map[string]bool)
		src, err := config.Load(m.sourceFile)
		if err != nil {
			m.status = "reload error: " + err.Error()
			m.statusErr = true
		} else {
			m.src = src
		}
		m.buildRows()

	case appliedMsg:
		m.status = fmt.Sprintf("✓ Written to %s — restart Claude Desktop", msg.destPath)
		m.statusErr = false

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			if m.pendingCount() > 0 {
				// discard pending and quit
				m.pending = make(map[string]bool)
				m.buildRows()
			}
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.showComment()

		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
			m.showComment()

		case " ": // toggle
			if len(m.rows) == 0 {
				break
			}
			row := m.rows[m.cursor]
			original := config.IsEnabled(row.server)
			newState := !row.enabled

			if newState == original {
				// back to original — remove from pending
				delete(m.pending, row.name)
			} else {
				m.pending[row.name] = newState
			}
			m.buildRows()

		case "s": // save pending to source file
			if m.pendingCount() == 0 {
				m.status = "no pending changes to save"
				m.statusErr = false
				break
			}
			return m, m.cmdSave()

		case "a": // save + apply to Claude Desktop
			return m, m.cmdApply()

		case "d": // dry-run preview
			return m, m.cmdDryRun()

		case "?", "h":
			m.status = "space toggle  s save  a apply  d dry-run  q quit"
			m.statusErr = false
		}
	}

	return m, nil
}

// ── Commands ──────────────────────────────────────────────────────────────────

func (m *Model) cmdSave() tea.Cmd {
	return func() tea.Msg {
		if err := m.flushPending(); err != nil {
			return statusMsg{text: "save failed: " + err.Error(), err: true}
		}
		return savedMsg{}
	}
}

func (m *Model) cmdApply() tea.Cmd {
	return func() tea.Msg {
		// Flush pending first
		if m.pendingCount() > 0 {
			if err := m.flushPending(); err != nil {
				return statusMsg{text: "save failed: " + err.Error(), err: true}
			}
		}

		src, err := config.Load(m.sourceFile)
		if err != nil {
			return statusMsg{text: "load failed: " + err.Error(), err: true}
		}

		dest := config.Filter(src)

		destPath, err := config.DestPath()
		if err != nil {
			return statusMsg{text: "dest path: " + err.Error(), err: true}
		}

		merged, err := config.MergeConfig(destPath, dest)
		if err != nil {
			return statusMsg{text: "merge failed: " + err.Error(), err: true}
		}
		if err := config.WriteConfig(destPath, merged, m.keepBackups); err != nil {
			return statusMsg{text: "write failed: " + err.Error(), err: true}
		}

		return appliedMsg{destPath: destPath}
	}
}

func (m *Model) cmdDryRun() tea.Cmd {
	return func() tea.Msg {
		// Build effective source (orig + pending) without writing to disk
		srccopy := cloneSource(m.src)
		for name, enabled := range m.pending {
			srv := srccopy.MCPServers[name]
			srv.Enabled = &enabled
			srccopy.MCPServers[name] = srv
		}
		dest := config.Filter(srccopy)
		names := sortedServerNames(dest.MCPServers)
		if len(names) == 0 {
			return statusMsg{text: "dry-run: no enabled servers", err: false}
		}
		return statusMsg{
			text: "dry-run enabled: " + strings.Join(names, ", "),
			err:  false,
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// flushPending writes pending toggles to the source file.
func (m *Model) flushPending() error {
	for name, enabled := range m.pending {
		srv := m.src.MCPServers[name]
		e := enabled
		srv.Enabled = &e
		m.src.MCPServers[name] = srv
	}
	return config.Save(m.sourceFile, m.src)
}

func cloneSource(src *config.SourceConfig) *config.SourceConfig {
	clone := &config.SourceConfig{
		Comment:    src.Comment,
		MCPServers: make(map[string]config.MCPServer, len(src.MCPServers)),
	}
	for k, v := range src.MCPServers {
		clone.MCPServers[k] = v
	}
	return clone
}

func sortedServerNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	total := len(m.rows)
	enabled := m.enabledCount()
	pending := m.pendingCount()

	// ── Header ────────────────────────────────────────────────────────────────
	header := styleTitle.Render("Claude Desktop MCP Servers")
	counts := styleSubtle.Render(fmt.Sprintf("%d enabled / %d total", enabled, total))
	if pending > 0 {
		counts += "  " + stylePending.Render(fmt.Sprintf("(%d unsaved)", pending))
	}
	b.WriteString(header + "  " + counts + "\n")
	b.WriteString(styleDivider.Render(strings.Repeat("─", max(m.width, 72))) + "\n")

	// ── Server list ───────────────────────────────────────────────────────────
	// Calculate column widths from actual data
	maxName := 4 // min "NAME"
	for _, r := range m.rows {
		if len(r.name) > maxName {
			maxName = len(r.name)
		}
	}

	// Column header
	header2 := fmt.Sprintf("  %-4s  %-*s  %s", "    ", maxName, "NAME", "URL")
	b.WriteString(styleSubtle.Render(header2) + "\n")

	for i, row := range m.rows {
		url := ""
		if len(row.server.Args) > 1 {
			url = row.server.Args[1]
		}

		_, isPending := m.pending[row.name]

		var indicator string
		var nameStyle, urlStyle lipgloss.Style
		switch {
		case isPending && row.enabled:
			indicator = stylePending.Render("●~")
			nameStyle = stylePending
			urlStyle = stylePending
		case isPending && !row.enabled:
			indicator = stylePending.Render("○~")
			nameStyle = stylePending
			urlStyle = stylePending
		case row.enabled:
			indicator = styleEnabled.Render("● ")
			nameStyle = styleEnabled
			urlStyle = styleURL
		default:
			indicator = styleDisabled.Render("○ ")
			nameStyle = styleDisabled
			urlStyle = styleDisabled
		}

		name := nameStyle.Render(fmt.Sprintf("%-*s", maxName, row.name))
		urlStr := urlStyle.Render(url)

		line := fmt.Sprintf("  %s  %s  %s", indicator, name, urlStr)

		if i == m.cursor {
			// pad to terminal width for full-width highlight
			padded := line + strings.Repeat(" ", max(0, m.width-lipgloss.Width(line)-1))
			b.WriteString(styleCursorBar.Render(padded) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	// ── Divider ───────────────────────────────────────────────────────────────
	b.WriteString(styleDivider.Render(strings.Repeat("─", max(m.width, 72))) + "\n")

	// ── Status bar ────────────────────────────────────────────────────────────
	if m.status != "" {
		var rendered string
		if m.statusErr {
			rendered = styleStatusErr.Render(m.status)
		} else if strings.HasPrefix(m.status, "✓") {
			rendered = styleStatusOK.Render(m.status)
		} else {
			rendered = styleStatusInfo.Render(m.status)
		}
		b.WriteString(rendered + "\n")
	}

	// ── Key hints ─────────────────────────────────────────────────────────────
	hints := []string{
		styleKey.Render("↑↓") + styleSubtle.Render("/") + styleKey.Render("jk") + styleSubtle.Render(" navigate"),
		styleKey.Render("space") + styleSubtle.Render(" toggle"),
		styleKey.Render("s") + styleSubtle.Render(" save"),
		styleKey.Render("a") + styleSubtle.Render(" apply"),
		styleKey.Render("d") + styleSubtle.Render(" dry-run"),
		styleKey.Render("q") + styleSubtle.Render(" quit"),
	}
	b.WriteString(styleSubtle.Render(strings.Join(hints, "  ")))

	return b.String()
}
