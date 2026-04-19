package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordCapsEntriesPerApp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	for i := 0; i < 210; i++ {
		Record("app-a", "default", 80)
	}
	for i := 0; i < 30; i++ {
		Record("app-b", "default", 90)
	}

	entries, err := load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	countA := 0
	countB := 0
	for _, e := range entries {
		switch e.App {
		case "app-a":
			countA++
		case "app-b":
			countB++
		}
	}

	if countA != 200 {
		t.Fatalf("expected 200 entries for app-a, got %d", countA)
	}
	if countB != 30 {
		t.Fatalf("expected 30 entries for app-b, got %d", countB)
	}
}

func TestClearAppOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	Record("app-a", "default", 70)
	Record("app-a", "prod", 75)
	Record("app-b", "default", 90)

	removed, err := Clear("app-a", "")
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if removed != 2 {
		t.Fatalf("expected 2 removed entries, got %d", removed)
	}

	entries, err := load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(entries) != 1 || entries[0].App != "app-b" {
		t.Fatalf("expected only app-b entry to remain, got %#v", entries)
	}
}

func TestClearAllRemovesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	Record("app-a", "default", 70)
	Record("app-b", "default", 90)

	removed, err := Clear("", "")
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if removed != 2 {
		t.Fatalf("expected 2 removed entries, got %d", removed)
	}

	if _, err := os.Stat(filepath.Join(home, ".kube-debugger", "history.json")); !os.IsNotExist(err) {
		t.Fatalf("expected history file to be removed, got err=%v", err)
	}
}
