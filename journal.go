package main

import (
	"fmt"
	"time"
)

// ResolveJournalPath returns the Dropbox path for a given date's journal entry.
func ResolveJournalPath(date time.Time) string {
	return fmt.Sprintf("/Notes/Journal/%s/%s/Note%s.md",
		date.Format("2006"),
		date.Format("01"),
		date.Format("20060102"),
	)
}
