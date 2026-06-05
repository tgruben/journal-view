package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateLoading = iota
	StateSuccess
	StateError
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			MarginTop(1)

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)
)

// FetchFn fetches a journal entry for a given date.
type FetchFn func(date time.Time) (string, error)

type Model struct {
	state    int
	content  string
	err      error
	spinner  spinner.Model
	date     time.Time
	width    int
	height   int
	viewport viewport.Model
	fetchFn  FetchFn
	renderer *glamour.TermRenderer
}

// LoadedMsg carries fetched journal content.
type LoadedMsg struct {
	Content string
	Err     error
}

func NewModel(date time.Time, fetch FetchFn) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	return Model{
		state:    StateLoading,
		date:     date,
		fetchFn:  fetch,
		spinner:  s,
		viewport: vp,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		if m.state == StateSuccess && m.content != "" {
			m.updateViewportContent()
		}
		return m, nil

	case LoadedMsg:
		if msg.Err != nil {
			m.state = StateError
			m.err = msg.Err
		} else {
			m.state = StateSuccess
			m.content = msg.Content
			m.updateViewportContent()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "left", "h":
			return m.prevDate()
		case "right", "l":
			return m.nextDate()
		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	var s string

	switch m.state {
	case StateLoading:
		s = fmt.Sprintf("\n  %s Fetching journal entry for %s...\n\n  Press q to quit\n",
			m.spinner.View(),
			m.date.Format("2006-01-02"),
		)

	case StateSuccess:
		dateStr := m.date.Format("Monday, January 2, 2006")
		header := fmt.Sprintf("  %s — %s\n\n",
			titleStyle.Render("Journal Entry"),
			dateStyle.Render(dateStr),
		)

		content := m.viewport.View()

		footer := fmt.Sprintf("\n  %s",
			helpStyle.Render("←/→ h/l: navigate  ↑/↓ j/k: scroll  g/G: top/bottom  mouse wheel: scroll  q: quit"),
		)

		s = header + content + footer

	case StateError:
		s = fmt.Sprintf("\n  %s\n\n  Error fetching journal entry:\n  %s\n\n  Press q to quit\n",
			titleStyle.Render("Error"),
			errorStyle.Render(m.err.Error()),
		)
	}

	return s
}

func (m *Model) updateViewportContent() {
	width := m.width
	if width == 0 {
		width = 80
	}
	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4),
	)
	rendered, err := m.renderer.Render(m.content)
	if err != nil {
		rendered = m.content
	}
	m.viewport.SetContent(rendered)
}

func (m Model) prevDate() (tea.Model, tea.Cmd) {
	if m.fetchFn == nil {
		return m, nil
	}
	m.date = m.date.AddDate(0, 0, -1)
	m.state = StateLoading
	m.viewport.GotoTop()
	m.viewport.SetContent("")
	return m, tea.Batch(m.spinner.Tick, m.fetchCmd())
}

func (m Model) nextDate() (tea.Model, tea.Cmd) {
	if m.fetchFn == nil {
		return m, nil
	}
	m.date = m.date.AddDate(0, 0, 1)
	m.state = StateLoading
	m.viewport.GotoTop()
	m.viewport.SetContent("")
	return m, tea.Batch(m.spinner.Tick, m.fetchCmd())
}

func (m Model) fetchCmd() tea.Cmd {
	if m.fetchFn == nil {
		return nil
	}
	fetchDate := m.date
	fetchFn := m.fetchFn
	return func() tea.Msg {
		content, err := fetchFn(fetchDate)
		return LoadedMsg{Content: content, Err: err}
	}
}

// ExitWithMessage prints an error message and exits without launching the TUI.
func ExitWithMessage(err error) {
	fmt.Fprintf(os.Stderr, "journal-view: %v\n", err)
	os.Exit(1)
}
