package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hermanu/logz/internal/parser"
)

// FilterState manages all filter state
type FilterState struct {
	Level      string
	Services   []string
	ServiceOn  map[string]bool
	Window     string
}

func NewFilterState() *FilterState {
	return &FilterState{
		ServiceOn: make(map[string]bool),
		Window:    "1h",
	}
}

// LevelCount represents counts for each level
type LevelCount struct {
	Name   string
	Count  int
	Active bool
}

// Sidebar renders the filter sidebar
type Sidebar struct {
	Filter      *FilterState
	LevelCounts []LevelCount
	Width       int
	Selected    int
	Focused     bool
}

func NewSidebar(w int) *Sidebar {
	counts := []LevelCount{
		{Name: "FATAL", Count: 0, Active: false},
		{Name: "ERROR", Count: 0, Active: true},
		{Name: "WARN", Count: 0, Active: false},
		{Name: "INFO", Count: 0, Active: false},
		{Name: "DEBUG", Count: 0, Active: false},
		{Name: "TRACE", Count: 0, Active: false},
	}
	return &Sidebar{
		Filter:      NewFilterState(),
		LevelCounts: counts,
		Width:       w,
	}
}

func (s *Sidebar) Render() string {
	var lines []string

	lines = append(lines, s.renderSection("LEVELS", s.renderLevels()))
	if len(s.Filter.Services) > 0 {
		lines = append(lines, "")
		lines = append(lines, s.renderSection("SERVICES", s.renderServices()))
	}
	lines = append(lines, "")
	lines = append(lines, s.renderSection("WINDOWS", s.renderWindows()))

	for len(lines) < 18 {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (s *Sidebar) renderSection(title, content string) string {
	return lipgloss.NewStyle().Foreground(Muted).Bold(true).Render(title)
}

func (s *Sidebar) renderLevels() string {
	var lines []string
	idx := 0
	for _, lc := range s.LevelCounts {
		prefix := " "
		if idx == s.Selected && s.Focused {
			prefix = ">"
		}
		symbol := "□"
		if lc.Active {
			symbol = "■"
		}
		levelStyle := lipgloss.NewStyle().Foreground(LevelColor(lc.Name))
		if lc.Active {
			levelStyle = levelStyle.Bold(true)
		}
		countStr := formatCount(lc.Count)
		lines = append(lines, fmt.Sprintf("%s%s %s %s", prefix, symbol, levelStyle.Render(lc.Name), lipgloss.NewStyle().Foreground(Muted).Render(countStr)))
		idx++
	}
	return strings.Join(lines, "\n")
}

func (s *Sidebar) renderServices() string {
	var lines []string
	idx := 0
	for _, svc := range s.Filter.Services {
		prefix := " "
		active := s.Filter.ServiceOn[svc]
		symbol := "□"
		if active {
			symbol = "■"
		}
		svcStyle := lipgloss.NewStyle()
		if active {
			svcStyle = svcStyle.Foreground(Foreground)
		} else {
			svcStyle = svcStyle.Foreground(Muted)
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", prefix, symbol, svcStyle.Render(svc)))
		idx++
	}
	return strings.Join(lines, "\n")
}

func (s *Sidebar) renderWindows() string {
	windows := []string{"5m", "15m", "1h", "24h"}
	var lines []string
	for _, w := range windows {
		prefix := " "
		if w == s.Filter.Window {
			prefix = ">"
		}
		symbol := "○"
		if w == s.Filter.Window {
			symbol = "●"
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", prefix, symbol, w))
	}
	return strings.Join(lines, "\n")
}

// LogList displays scrollable log entries
type LogList struct {
	Entries  []LogEntry
	Selected int
	Scroll   int
	Width    int
	Height   int
	Focused  bool
}

type LogEntry struct {
	Timestamp time.Time
	Level     parser.Level
	Message   string
	Fields    map[string]string
}

func NewLogList(w, h int) *LogList {
	return &LogList{
		Width:    w,
		Height:   h,
		Selected: 0,
		Scroll:   0,
	}
}

func (l *LogList) SetEntries(entries []parser.Entry, _ *FilterState) {
	l.Entries = make([]LogEntry, len(entries))
	for i, e := range entries {
		l.Entries[i] = LogEntry{
			Timestamp: e.Timestamp,
			Level:     e.Level,
			Message:   e.Message,
			Fields:    e.Fields,
		}
	}
	if l.Selected >= len(l.Entries) && len(l.Entries) > 0 {
		l.Selected = len(l.Entries) - 1
	}
	l.ensureScroll()
}

func (l *LogList) ensureScroll() {
	visible := l.Height - 2
	if l.Selected < l.Scroll {
		l.Scroll = l.Selected
	} else if l.Selected >= l.Scroll+visible {
		l.Scroll = l.Selected - visible + 1
	}
}

func (l *LogList) MoveUp() {
	if l.Selected > 0 {
		l.Selected--
		l.ensureScroll()
	}
}

func (l *LogList) MoveDown() {
	if l.Selected < len(l.Entries)-1 {
		l.Selected++
		l.ensureScroll()
	}
}

func (l *LogList) Render() string {
	var lines []string
	visible := l.Height - 2
	start := l.Scroll
	end := l.Scroll + visible
	if end > len(l.Entries) {
		end = len(l.Entries)
	}
	if start < 0 {
		start = 0
	}

	for i := start; i < end; i++ {
		e := l.Entries[i]
		selected := i == l.Selected
		lines = append(lines, renderLogLine(e, selected, l.Width-2))
	}

	for len(lines) < visible {
		lines = append(lines, strings.Repeat(" ", l.Width-2))
	}
	return strings.Join(lines, "\n")
}

func renderLogLine(e LogEntry, selected bool, width int) string {
	ts := ""
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.Format("15:04:05.000")
	}
	level := strings.ToUpper(e.Level.String())
	msg := e.Message
	if len(msg) > width-20 {
		msg = msg[:width-23] + "..."
	}

	levelCol := LevelColor(level)
	selChar := " "
	if selected {
		selChar = ">"
	}

	if selected {
		return fmt.Sprintf("%s%s %s %s",
			selChar,
			lipgloss.NewStyle().Foreground(Muted).Render(ts),
			lipgloss.NewStyle().Foreground(levelCol).Bold(true).Render(padLeft(level, 5)),
			lipgloss.NewStyle().Foreground(Foreground).Render(msg),
		)
	}
	return fmt.Sprintf("%s%s %s %s",
		selChar,
		lipgloss.NewStyle().Foreground(Muted).Render(ts),
		lipgloss.NewStyle().Foreground(levelCol).Render(padLeft(level, 5)),
		lipgloss.NewStyle().Foreground(Foreground).Render(msg),
	)
}

func padLeft(s string, w int) string {
	l := len([]rune(s))
	if l >= w {
		return s[:w]
	}
	return strings.Repeat(" ", w-l) + s
}

// DetailPane shows full entry details
type DetailPane struct {
	Entry   *LogEntry
	Width   int
	Height  int
	Focused bool
}

func NewDetailPane(w, h int) *DetailPane {
	return &DetailPane{Width: w, Height: h}
}

func (d *DetailPane) SetEntry(e *LogEntry) {
	if e != nil {
		d.Entry = e
	}
}

func (d *DetailPane) Render() string {
	if d.Entry == nil {
		lines := []string{}
		for i := 0; i < d.Height-2; i++ {
			lines = append(lines, lipgloss.NewStyle().Foreground(Muted).Render("Select an entry"))
		}
		return strings.Join(lines, "\n")
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Foreground(Muted).Bold(true).Render("TIMESTAMP"))
	if !d.Entry.Timestamp.IsZero() {
		lines = append(lines, d.Entry.Timestamp.Format("2006-01-02T15:04:05.000Z"))
	}
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(Muted).Bold(true).Render("LEVEL"))
	lvl := strings.ToUpper(d.Entry.Level.String())
	lines = append(lines, lipgloss.NewStyle().Foreground(LevelColor(lvl)).Bold(true).Render(lvl))
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(Muted).Bold(true).Render("MESSAGE"))
	lines = append(lines, d.Entry.Message)

	if len(d.Entry.Fields) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(Muted).Bold(true).Render("METADATA"))
		for k, v := range d.Entry.Fields {
			lines = append(lines, fmt.Sprintf("%s: %s", k, v))
		}
	}

	for len(lines) < d.Height-2 {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func formatCount(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}