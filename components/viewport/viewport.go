package viewport

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const lineContinuationIndicator = "..."

// New returns a new model with the given width and height as well as default values.
func New(width, height int) (m Model) {
	m.width = width
	m.height = height
	m.setInitialValues()
	return m
}

// Model is the Bubble Tea model for this viewport element.
type Model struct {
	width         int
	height        int
	contentHeight int // excludes header height, should always be internal
	keyMap        viewportKeyMap

	// Currently, causes flickering if enabled.
	mouseWheelEnabled bool

	// yOffset is the vertical scroll position of the text.
	yOffset int

	// xOffset is the horizontal scroll position of the text.
	xOffset int

	// CursorRow is the row index of the cursor.
	CursorRow int

	// styleViewport applies a lipgloss style to the viewport. Realistically, it's most
	// useful for setting borders, margins and padding.
	styleViewport  lipgloss.Style
	styleCursorRow lipgloss.Style

	initialized   bool
	header        []string
	lines         []string
	maxLineLength int
}

func (m Model) ContentEmpty() bool {
	return len(m.header) == 0 && len(m.lines) == 0
}

func (m *Model) setInitialValues() {
	m.contentHeight = m.height - len(m.header)
	m.keyMap = getKeyMap()
	m.mouseWheelEnabled = false
	m.styleViewport = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	m.styleCursorRow = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	m.initialized = true
}

// Init exists to satisfy the tea.Model interface for composability purposes.
func (m Model) Init() tea.Cmd {
	return nil
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// fixCursorRow adjusts the cursor to be in a visible location if it is outside the visible content
func (m *Model) fixCursorRow() {
	if m.CursorRow > m.lastVisibleLineIdx() {
		m.SetCursorRow(m.lastVisibleLineIdx())
	}
}

// SetHeight sets the pager's height, including header.
func (m *Model) SetHeight(h int) {
	m.height = h
	m.contentHeight = h - len(m.header)
	m.fixCursorRow()
}

// SetWidth sets the pager's width.
func (m *Model) SetWidth(w int) {
	m.width = w
}

// SetHeader sets the pager's header content.
func (m *Model) SetHeader(header string) {
	m.header = strings.Split(normalizeLineEndings(header), "\n")
	m.contentHeight = m.height - len(m.header)
}

// SetContent sets the pager's text content.
func (m *Model) SetContent(s string) {
	lines := strings.Split(normalizeLineEndings(s), "\n")
	maxLineLength := 0
	for _, line := range lines {
		if lineLength := len(line); lineLength > maxLineLength {
			maxLineLength = lineLength
		}
	}
	m.lines = lines
	m.maxLineLength = maxLineLength
	m.fixCursorRow()
}

// SetLoading clears the header and content and displays the loading message
func (m *Model) SetLoading(s string) {
	m.SetContent("")
	m.SetHeader(s)
}

// maxLinesIdx returns the maximum index of the model's lines
func (m *Model) maxLinesIdx() int {
	return len(m.lines) - 1
}

// lastVisibleLineIdx returns the maximum visible line index
func (m Model) lastVisibleLineIdx() int {
	return min(m.maxLinesIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model) maxYOffset() int {
	if m.maxLinesIdx() < m.contentHeight {
		return 0
	}
	return m.maxLinesIdx() - m.contentHeight + 1
}

// maxCursorRow returns the maximum CursorRow
func (m Model) maxCursorRow() int {
	return len(m.lines) - 1
}

// setYOffset sets the yOffset with bounds.
func (m *Model) setYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.yOffset = maxYOffset
	} else {
		m.yOffset = max(0, n)
	}
}

// SetCursorRow sets the CursorRow with bounds. Adjusts yOffset as necessary.
func (m *Model) SetCursorRow(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelection := m.maxCursorRow(); n > maxSelection {
		m.CursorRow = maxSelection
	} else {
		m.CursorRow = max(0, n)
	}

	if lastVisibleLineIdx := m.lastVisibleLineIdx(); m.CursorRow > lastVisibleLineIdx {
		m.viewDown(m.CursorRow - lastVisibleLineIdx)
	} else if m.CursorRow < m.yOffset {
		m.viewUp(m.yOffset - m.CursorRow)
	}
}

// visibleLines retrieves the visible lines based on the yOffset
func (m Model) visibleLines() []string {
	start := m.yOffset
	end := start + m.contentHeight
	if end > m.maxLinesIdx() {
		return m.lines[start:]
	}
	return m.lines[start:end]
}

// cursorRowDown moves the CursorRow down by the given number of lines.
func (m *Model) cursorRowDown(n int) {
	m.SetCursorRow(m.CursorRow + n)
}

// cursorRowUp moves the CursorRow up by the given number of lines.
func (m *Model) cursorRowUp(n int) {
	m.SetCursorRow(m.CursorRow - n)
}

// viewDown moves the view down by the given number of lines.
func (m *Model) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

// viewUp moves the view up by the given number of lines.
func (m *Model) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

// viewLeft moves the view left the given number of columns.
func (m *Model) viewLeft(n int) {
	m.xOffset = max(0, m.xOffset-n)
}

// viewRight moves the view right the given number of columns.
func (m *Model) viewRight(n int) {
	m.xOffset = min(m.maxLineLength-m.width, m.xOffset+n)
}

// Update handles standard message-based viewport updates.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.initialized {
		m.setInitialValues()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.cursorRowUp(1)

		case key.Matches(msg, m.keyMap.Down):
			m.cursorRowDown(1)

		case key.Matches(msg, m.keyMap.Left):
			m.viewLeft(m.width / 4)

		case key.Matches(msg, m.keyMap.Right):
			m.viewRight(m.width / 4)

		case key.Matches(msg, m.keyMap.HalfPageUp):
			m.viewUp(m.contentHeight / 2)
			m.cursorRowUp(m.contentHeight / 2)

		case key.Matches(msg, m.keyMap.HalfPageDown):
			m.viewDown(m.contentHeight / 2)
			m.cursorRowDown(m.contentHeight / 2)

		case key.Matches(msg, m.keyMap.PageUp):
			m.viewUp(m.contentHeight)
			m.cursorRowUp(m.contentHeight)

		case key.Matches(msg, m.keyMap.PageDown):
			m.viewDown(m.contentHeight)
			m.cursorRowDown(m.contentHeight)
		}

		//dev.Debug(fmt.Sprintf("selection %d, yoffset %d, height %d, contentHeight %d, len(m.lines) %d", m.CursorRow, m.yOffset, m.height, m.contentHeight, len(m.lines)))

	case tea.MouseMsg:
		if !m.mouseWheelEnabled {
			break
		}
		switch msg.Type {
		case tea.MouseWheelUp:
			m.cursorRowUp(1)

		case tea.MouseWheelDown:
			m.cursorRowDown(1)
		}
	}

	// could return non-nil cmd in the future
	return m, nil
}

// View returns the string representing the viewport.
func (m Model) View() string {
	visibleLines := m.visibleLines()

	// TODO LEO: deal with headers that are wider than viewport width
	viewLines := strings.Join(m.header, "\n") + "\n"
	for idx, line := range visibleLines {
		if lineLength := len(line); lineLength > m.width {
			line = line[m.xOffset : m.xOffset+m.width]
			if m.xOffset < lineLength-m.width {
				line = line[:len(line)-len(lineContinuationIndicator)] + lineContinuationIndicator
			}
			if m.xOffset > 0 {
				line = lineContinuationIndicator + line[len(lineContinuationIndicator):]
			}
		}
		if m.yOffset+idx == m.CursorRow {
			// render selected row
			viewLines += m.styleCursorRow.
				Width(m.width).
				Render(line)
		} else {
			viewLines += line
		}
		viewLines += "\n"
	}
	trimmedViewLines := strings.TrimSpace(viewLines)
	renderedViewLines := m.styleViewport.
		Width(m.width).
		Height(m.height).
		Render(trimmedViewLines)
	return renderedViewLines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
