package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hermanu/logz/internal/config"
	"github.com/hermanu/logz/internal/filter"
	"github.com/hermanu/logz/internal/parser"
	"github.com/hermanu/logz/internal/ui"
	"github.com/spf13/cobra"
)

func interactiveRun(_ *cobra.Command, args []string) error {
	sources := args
	if len(sources) == 0 || (len(sources) == 1 && sources[0] == "-") {
		sources = []string{"-"}
	}

	m := interactiveModel{
		sources: sources,
		layout:  ui.NewLayoutFromDimensions(100, 30),
		sidebar: ui.NewSidebar(18),
		loglist:  ui.NewLogList(50, 24),
		detail:  ui.NewDetailPane(30, 24),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

type interactiveModel struct {
	sources []string
	width   int
	height  int
	ready   bool
	err     error

	layout  *ui.Layout2
	sidebar *ui.Sidebar
	loglist *ui.LogList
	detail  *ui.DetailPane

	entries []parser.Entry
	stats   logStats

	showHelp bool
}

type logStats struct {
	total   int
	matches int
	byLevel map[parser.Level]int
}

func (m interactiveModel) Init() tea.Cmd {
	return func() tea.Msg {
		entries, stats, err := loadLogs(m.sources, m.sidebar.Filter)
		if err != nil {
			return loadError{err: err}
		}
		return logsLoaded{entries: entries, stats: stats}
	}
}

func loadLogs(sources []string, f *ui.FilterState) ([]parser.Entry, logStats, error) {
	stats := logStats{
		byLevel: make(map[parser.Level]int),
	}

	filt := filter.Filter{Invert: false}

	// Level filter
	if f.Level != "" {
		if lvl, ok := parser.ParseLevel(f.Level); ok {
			filt.MinLevel = lvl
		}
	}

	// Window filter
	windowDur := parseWindow(f.Window)
	if windowDur > 0 {
		filt.Since = time.Now().Add(-windowDur)
	}

	// Service filter
	if len(f.Services) > 0 {
		filt.FieldEquals = make(map[string]string)
		for _, svc := range f.Services {
			if f.ServiceOn[svc] {
				filt.FieldEquals["service"] = svc
				break
			}
		}
	}

	entries := []parser.Entry{}
	serviceSet := make(map[string]bool)

	for _, src := range sources {
		r := os.Stdin
		if src != "-" {
			var err error
			r, err = os.Open(src)
			if err != nil {
				continue
			}
			defer r.Close()
		}

		cfg, _ := config.Load()
		opts, _ := cfg.ParserOptions()
		p := parser.NewJSONParser(opts)

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			entry, err := p.Parse(line)
			if err != nil {
				continue
			}

			// Auto-detect service
			if svc, ok := entry.Fields["service"]; ok {
				if !serviceSet[svc] {
					serviceSet[svc] = true
					f.Services = append(f.Services, svc)
					f.ServiceOn[svc] = false
				}
			}

			if filt.Allow(entry) {
				entries = append(entries, entry)
			}
			stats.byLevel[entry.Level]++
			stats.total++
		}
	}

	stats.matches = len(entries)

	// Cap at 100k
	if len(entries) > 100000 {
		entries = entries[len(entries)-100000:]
	}

	return entries, stats, nil
}

func parseWindow(w string) time.Duration {
	switch w {
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "24h":
		return 24 * time.Hour
	default:
		return 0
	}
}

type loadError struct{ err error }

func (e loadError) Error() string { return e.err.Error() }

type logsLoaded struct {
	entries []parser.Entry
	stats   logStats
}

func (m interactiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout = ui.NewLayoutFromDimensions(msg.Width, msg.Height)
		sw, lw, dw := m.layout.CalcSizes()
		m.sidebar = ui.NewSidebar(sw)
		m.loglist = ui.NewLogList(lw, m.height-4)
		m.detail = ui.NewDetailPane(dw, m.height-4)
		m.ready = true
		return m, nil

	case loadError:
		m.err = msg.err
		return m, nil

	case logsLoaded:
		m.entries = msg.entries
		m.stats = msg.stats

		// Update level counts
		levels := []ui.LevelCount{
			{Name: "FATAL", Count: msg.stats.byLevel[parser.LevelFatal]},
			{Name: "ERROR", Count: msg.stats.byLevel[parser.LevelError]},
			{Name: "WARN", Count: msg.stats.byLevel[parser.LevelWarn]},
			{Name: "INFO", Count: msg.stats.byLevel[parser.LevelInfo]},
			{Name: "DEBUG", Count: msg.stats.byLevel[parser.LevelDebug]},
		}
		for i := range levels {
			levels[i].Active = m.sidebar.Filter.Level == levels[i].Name
		}
		m.sidebar.LevelCounts = levels

		m.loglist.SetEntries(msg.entries, m.sidebar.Filter)
		if m.loglist.Selected < len(m.loglist.Entries) {
			m.detail.SetEntry(&m.loglist.Entries[m.loglist.Selected])
		}
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
		case "left":
			m.layout.FocusedPane = (m.layout.FocusedPane + 2) % 3
		case "right":
			m.layout.FocusedPane = (m.layout.FocusedPane + 1) % 3
		case "up":
			if m.layout.FocusedPane == 1 {
				m.loglist.MoveUp()
				if m.loglist.Selected < len(m.loglist.Entries) {
					m.detail.SetEntry(&m.loglist.Entries[m.loglist.Selected])
				}
			}
		case "down":
			if m.layout.FocusedPane == 1 {
				m.loglist.MoveDown()
				if m.loglist.Selected < len(m.loglist.Entries) {
					m.detail.SetEntry(&m.loglist.Entries[m.loglist.Selected])
				}
			}
		case "/":
			m.sidebar.Filter.Level = "ERROR"
			return m, m.reload()
		}
	}
	return m, nil
}

func (m *interactiveModel) reload() tea.Cmd {
	return func() tea.Msg {
		entries, stats, err := loadLogs(m.sources, m.sidebar.Filter)
		if err != nil {
			return loadError{err: err}
		}
		return logsLoaded{entries: entries, stats: stats}
	}
}

func (m interactiveModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if !m.ready {
		return "Loading..."
	}

	sw, lw, dw := m.layout.CalcSizes()
	m.sidebar.Width = sw
	m.loglist.Width = lw
	m.detail.Width = dw

	sidebarView := m.sidebar.Render()
	listView := m.loglist.Render()
	detailView := m.detail.Render()

	status := fmt.Sprintf("%s lines · %s matches · %s window",
		formatNumber(m.stats.total),
		formatNumber(m.stats.matches),
		m.sidebar.Filter.Window)

	hints := "←/→ switch │ ↑/↓ navigate │ / filter │ ? help │ q quit"

	frame := m.layout.Frame(
		"logz --interactive",
		sidebarView,
		listView,
		detailView,
		status,
		hints,
	)

	if m.showHelp {
		return frame + "\n\n" + m.renderHelp()
	}

	return frame
}

func (m interactiveModel) renderHelp() string {
	help := []string{
		"╭────────────── HELP ──────────────╮",
		"│ Key       │ Action               │",
		"├──────────┼─────────────────────────┤",
		"│ ←/→      │ Switch pane           │",
		"│ ↑/↓      │ Navigate list        │",
		"│ Enter    │ View details        │",
		"│ /        │ Filter (ERROR)     │",
		"│ l        │ Toggle level       │",
		"│ s        │ Toggle service    │",
		"│ t        │ Cycle time window │",
		"│ ?        │ Toggle this help │",
		"│ q        │ Quit              │",
		"╰──────────┴─────────────────────╯",
	}
	return strings.Join(help, "\n")
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}