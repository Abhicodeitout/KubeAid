// Package history persists health-score snapshots to ~/.kube-debugger/history.json
// so users can track trends over time with `kube-debugger history`.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry is a single health-score snapshot.
type Entry struct {
	App         string    `json:"app"`
	Namespace   string    `json:"namespace"`
	HealthScore int       `json:"health_score"`
	RecordedAt  time.Time `json:"recorded_at"`
}

func dataFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kube-debugger", "history.json")
}

// DataFilePath returns the path to the history file (exported for CLI use).
func DataFilePath() string {
	return dataFile()
}

func load() ([]Entry, error) {
	f := dataFile()
	data, err := os.ReadFile(f)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func save(entries []Entry) error {
	f := dataFile()
	if err := os.MkdirAll(filepath.Dir(f), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f, data, 0o600)
}

// Record appends a new health-score entry and keeps only the last 200 entries per app.
func Record(app, namespace string, score int) {
	entries, _ := load()
	entries = append(entries, Entry{
		App:         app,
		Namespace:   namespace,
		HealthScore: score,
		RecordedAt:  time.Now().UTC(),
	})
	entries = capEntriesPerApp(entries, 200)
	_ = save(entries)
}

// Clear removes history entries for a single app (and optional namespace).
// If app is empty, it clears all history.
func Clear(app, namespace string) (int, error) {
	entries, err := load()
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, nil
	}

	app = strings.TrimSpace(app)
	namespace = strings.TrimSpace(namespace)

	if app == "" {
		if err := os.Remove(dataFile()); err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		return len(entries), nil
	}

	kept := make([]Entry, 0, len(entries))
	removed := 0
	for _, e := range entries {
		matchesApp := e.App == app
		matchesNS := namespace == "" || e.Namespace == namespace
		if matchesApp && matchesNS {
			removed++
			continue
		}
		kept = append(kept, e)
	}

	if removed == 0 {
		return 0, nil
	}

	if len(kept) == 0 {
		if err := os.Remove(dataFile()); err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		return removed, nil
	}

	return removed, save(kept)
}

func capEntriesPerApp(entries []Entry, perApp int) []Entry {
	if perApp <= 0 || len(entries) == 0 {
		return entries
	}

	counts := make(map[string]int)
	keptReversed := make([]Entry, 0, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if counts[e.App] >= perApp {
			continue
		}
		counts[e.App]++
		keptReversed = append(keptReversed, e)
	}

	for i, j := 0, len(keptReversed)-1; i < j; i, j = i+1, j-1 {
		keptReversed[i], keptReversed[j] = keptReversed[j], keptReversed[i]
	}
	return keptReversed
}

// GetHistory returns all entries for a given app+namespace, oldest first.
func GetHistory(app, namespace string) ([]Entry, error) {
	all, err := load()
	if err != nil {
		return nil, err
	}
	var result []Entry
	for _, e := range all {
		if e.App == app && (namespace == "" || e.Namespace == namespace) {
			result = append(result, e)
		}
	}
	return result, nil
}

// RenderHistory returns a formatted string of health score history.
func RenderHistory(app, namespace string) string {
	entries, err := GetHistory(app, namespace)
	if err != nil {
		return fmt.Sprintf("❌ Failed to load history: %v\n", err)
	}
	if len(entries) == 0 {
		return fmt.Sprintf("No history found for app '%s'. Run `analyze` first.\n", app)
	}

	out := fmt.Sprintf("Health score history for: %s (namespace: %s)\n", app, namespace)
	out += fmt.Sprintf("%-30s  %s\n", "Recorded At", "Score")
	out += fmt.Sprintf("%-30s  %s\n", "------------------------------", "------")
	for _, e := range entries {
		bar := scoreBar(e.HealthScore)
		out += fmt.Sprintf("%-30s  %3d/100  %s\n", e.RecordedAt.Format("2006-01-02 15:04:05 UTC"), e.HealthScore, bar)
	}
	return out
}

func scoreBar(score int) string {
	filled := score / 10
	return "[" + repeat("█", filled) + repeat("░", 10-filled) + "]"
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
