package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout2 manages the three-pane layout
type Layout2 struct {
	Width       int
	Height      int
	FocusedPane int
}

func NewLayoutFromDimensions(w, h int) *Layout2 {
	return &Layout2{
		Width:       w,
		Height:     h,
		FocusedPane: PaneList,
	}
}

func (l *Layout2) CalcSizes() (sidebarW, listW, detailW int) {
	sidebarW = 18
	detailW = 28
	listW = l.Width - sidebarW - detailW - 6
	if listW < 15 {
		listW = 15
	}
	return sidebarW, listW, detailW
}

func (l *Layout2) ListAreaSize() (w, h int) {
	_, listW, _ := l.CalcSizes()
	h = l.Height - 4
	return listW - 2, h
}

func (l *Layout2) DetailAreaSize() (w, h int) {
	_, _, dw := l.CalcSizes()
	h = l.Height - 4
	return dw - 2, h
}

func (l *Layout2) Frame(header, sidebar, list, detail, status, hints string) string {
	lines := []string{}

	borderTop := "╭" + strings.Repeat("─", l.Width-2) + "╮"
	lines = append(lines, borderTop)

	lines = append(lines, "│"+pad(header, l.Width-2)+"│")

	divider := "├" + strings.Repeat("─", l.Width-2) + "┤"
	lines = append(lines, divider)

	sidebarLines := splitLines(sidebar)
	listLines := splitLines(list)
	detailLines := splitLines(detail)

	sidebarW, _, _ := l.CalcSizes()
	maxLines := maxInt(maxInt(len(sidebarLines), len(listLines)), len(detailLines))
	for i := 0; i < maxLines; i++ {
		sl := getOrEmpty(sidebarLines, i, sidebarW)
		cl := getOrEmpty(listLines, i, l.Width-sidebarW-6)
		dl := getOrEmpty(detailLines, i, 28)
		lines = append(lines, "│"+sl+cl+dl+"│")
	}

	statusDiv := "├" + strings.Repeat("─", l.Width-2) + "┤"
	lines = append(lines, statusDiv)

	lines = append(lines, "│"+pad(status, l.Width-2)+"│")
	lines = append(lines, "│"+pad(hints, l.Width-2)+"│")

	borderBottom := "╰" + strings.Repeat("─", l.Width-2) + "╯"
	lines = append(lines, borderBottom)

	return strings.Join(lines, "\n")
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func getOrEmpty(s []string, i, w int) string {
	if i < len(s) {
		return pad(s[i], w)
	}
	return pad("", w)
}

func pad(s string, w int) string {
	runes := []rune(s)
	if len(runes) >= w {
		return string(runes[:w-1]) + "…"
	}
	return s + strings.Repeat(" ", w-len(runes))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ = lipgloss.NewStyle