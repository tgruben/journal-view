package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
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
)

// FetchFn is a function that fetches a journal entry and returns its content.
type FetchFn func() (string, error)

type Model struct {
	state   int
	content string
	err     error
	spinner spinner.Model
	date    time.Time
	width   int
	fetchFn FetchFn
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

	return Model{
		state:   StateLoading,
		date:    date,
		fetchFn: fetch,
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			content, err := m.fetchFn()
			return LoadedMsg{Content: content, Err: err}
		},
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case LoadedMsg:
		if msg.Err != nil {
			m.state = StateError
			m.err = msg.Err
		} else {
			m.state = StateSuccess
			m.content = msg.Content

			// Initialize renderer with word wrap
			width := m.width
			if width == 0 {
				width = 80
			}
			m.renderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(width-4),
			)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	var s string

	switch m.state {
	case StateLoading:
		s = fmt.Sprintf("\n %s Fetching journal entry for %s...\n\n Press q or esc to quit\n",
			m.spinner.View(),
			m.date.Format("2006-01-02"),
		)

	case StateSuccess:
		// Render the markdown
		rendered, err := m.renderer.Render(m.content)
		if err != nil {
			// Fallback to plain text if rendering fails
			rendered = m.content
		}

		dateStr := m.date.Format("Monday, January 2, 2006")
		s = fmt.Sprintf("\n %s — %s\n\n%s\n",
			titleStyle.Render("Journal Entry"),
			dateStyle.Render(dateStr),
			rendered,
		)

	case StateError:
		s = fmt.Sprintf("\n %s\n\n Error fetching journal entry:\n %s\n\n Press q or esc to quit\n",
			titleStyle.Render("Error"),
			errorStyle.Render(m.err.Error()),
		)
	}

	return s + "\n"
}

// ExitWithMessage prints an error message and exits without launching the TUI.
func ExitWithMessage(err error) {
	fmt.Fprintf(os.Stderr, "journal-view: %v\n", err)
	os.Exit(1)
}
