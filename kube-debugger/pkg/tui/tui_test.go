package tui

import "testing"

func TestTUIModelView(t *testing.T) {
	m := model{}
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty TUI view")
	}
}
