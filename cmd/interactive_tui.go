package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hermanu/logz/internal/config"
	"github.com/hermanu/logz/internal/filter"
	"github.com/hermanu/logz/internal/parser"
	"github.com/spf13/cobra"
)

func interactiveRun(_ *cobra.Command, args []string) error {
	m := newInteractiveModel(args)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithReportFocus())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func newInteractiveModel(args []string) interactiveModel {
	m := interactiveModel{
		sources: args,
		filter: filterState{
			servicesOn: make(map[string]bool),
		},
	}
	if len(m.sources) == 0 || (len(m.sources) == 1 && m.sources[0] == "-") {
		m.sources = []string{"-"}
	}
	return m
}

type interactiveModel struct {
	width  int
	height int
	ready  bool
	sources []string

	entries []parser.Entry
	stats  logStats
	filter filterState

	selected int
	focus    pane

	err error
}

type filterState struct {
	level    string
	service string
	window  time.Duration

	services   []string
	servicesOn map[string]bool
}

type logStats struct {
	total   int
	matches int
	byLevel map[parser.Level]int
}

type pane int

const (
	paneSidebar pane = iota
	paneList
	paneDetail
)

func (m interactiveModel) Init() tea.Cmd {
	return func() tea.Msg {
		entries, stats, err := loadLogsSync(m.sources, m.filter)
		if err != nil {
			return loadError{err: err}
		}
		return logsLoaded{entries: entries, stats: stats}
	}
}

func loadLogsSync(sources []string, f filterState) ([]parser.Entry, logStats, error) {
	stats := logStats{
		byLevel: make(map[parser.Level]int),
	}

	filt := &filter.Filter{Invert: false}

	if f.level != "" {
		if lvl, ok := parser.ParseLevel(f.level); ok {
			filt.MinLevel = lvl
		}
	}
	if f.window > 0 {
		filt.Since = time.Now().Add(-f.window)
	}

	entries := []parser.Entry{}
	services := []string{}

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

		sniffer := bufio.NewReaderSize(r, 16*1024)
		var lines []string
		for {
			line, err := sniffer.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			lines = append(lines, line)
		}

		// Re-read for parsing
		if src != "-" {
			r.Close()
			r, _ = os.Open(src)
			defer r.Close()
		} else {
			r = os.Stdin
		}

		for _, line := range lines {
			entry, err := p.Parse(line)
			if err != nil {
				if errors.Is(err, parser.ErrSkip) {
					continue
				}
				continue
			}

			if entry.Fields != nil {
				if svc, ok := entry.Fields["service"]; ok {
					if !contains(services, svc) {
						services = append(services, svc)
					}
				}
			}

			if filt.Allow(entry) {
				entries = append(entries, entry)
			}
			stats.byLevel[entry.Level]++
			stats.total++
		}
		_ = services
	}

	stats.matches = len(entries)
	if len(entries) > 100000 {
		entries = entries[len(entries)-100000:]
	}

	return entries, stats, nil
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
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
		m.ready = true
		return m, nil

	case loadError:
		m.err = msg.err
		return m, nil

	case logsLoaded:
		m.entries = msg.entries
		m.stats = msg.stats
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left":
			m.focus = pane((int(m.focus) + 2) % 3)
		case "right":
			m.focus = pane((int(m.focus) + 1) % 3)
		case "up":
			if m.selected > 0 {
				m.selected--
			}
		case "down":
			if m.selected < len(m.entries)-1 {
				m.selected++
			}
		}
	}
	return m, nil
}

func (m interactiveModel) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	if !m.ready {
		return "Loading..."
	}

	var status string
	if m.stats.total > 0 {
		status = fmt.Sprintf("%s lines · %s matches", formatNum(m.stats.total), formatNum(m.stats.matches))
	} else {
		status = "No logs. Use: logz interactive <file>"
	}

	lines := []string{
		"╭────────────────────────────────────────┮",
		"│ › logz --interactive                 │",
		"├──────────┬───────────────────────────┤",
	}

	count := min(15, len(m.entries))
	for i := 0; i < count; i++ {
		e := m.entries[i]
		prefix := " "
		if i == m.selected {
			prefix = ">"
		}
		msg := e.Message
		if len(msg) > 35 {
			msg = msg[:32] + "..."
		}
		lines = append(lines, fmt.Sprintf("│%s %-5s │ %s", prefix, e.Level, msg))
	}

	lines = append(lines, "├──────────┴───────────────────────────┤")
	lines = append(lines, "│ "+status+" │")
	lines = append(lines, "╰────────────────────────────────────────┘")
	lines = append(lines, "")
	lines = append(lines, "Controls: ←/→ switch panes | ↑/↓ navigate | q quit")

	return strings.Join(lines, "\n")
}

func formatNum(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}