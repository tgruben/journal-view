package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"journal-view/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func parseDate(args []string) (time.Time, error) {
	if len(args) == 0 {
		return time.Now().Truncate(24 * time.Hour), nil
	}

	input := args[0]

	// Try various date formats
	formats := []string{
		"2006-01-02",
		"20060102",
		"01/02/2006",
		"2006/01/02",
		"Jan 2, 2006",
		"January 2, 2006",
		"2006-1-2",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return t.Truncate(24 * time.Hour), nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %q (try YYYY-MM-DD)", input)
}

func fetchEntry(token string, date time.Time) (string, string, error) {
	path := ResolveJournalPath(date)
	client := &DropboxClient{Token: token}

	content, err := client.Download(path)
	if err != nil {
		return "", path, err
	}

	if content == "" {
		return "", path, fmt.Errorf("no journal entry found for %s", date.Format("2006-01-02"))
	}

	return content, path, nil
}

func main() {
	raw := flag.Bool("raw", false, "dump markdown content to stdout, bypassing the TUI")
	flag.Parse()

	date, err := parseDate(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "journal-view: %v\n", err)
		os.Exit(1)
	}

	configPath := defaultConfigPath()
	cfg, err := loadConfig(configPath)
	if err != nil {
		ui.ExitWithMessage(fmt.Errorf("loading config: %w", err))
	}

	token, err := resolveToken(cfg)
	if err != nil {
		ui.ExitWithMessage(err)
	}

	content, _, err := fetchEntry(token, date)
	if err != nil {
		ui.ExitWithMessage(err)
	}

	if *raw {
		fmt.Print(content)
		return
	}

	fetchFn := func() (string, error) {
		return content, nil
	}

	m := ui.NewModel(date, fetchFn)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		ui.ExitWithMessage(fmt.Errorf("running TUI: %w", err))
	}
}
