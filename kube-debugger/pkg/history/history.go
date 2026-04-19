// Package history persists health-score snapshots to ~/.kube-debugger/history.json
// so users can track trends over time with `kube-debugger history`.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	// cap at 1000 total
	if len(entries) > 1000 {
		entries = entries[len(entries)-1000:]
	}
	_ = save(entries)
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
