package tui

import (
	"fmt"
	"os"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)



type model struct{}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m model) View() string {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	return style.Render("kube-debugger TUI\nPress 'q' to quit.")
}

func StartTUI() {
	p := tea.NewProgram(model{})
	if err := p.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running TUI:", err)
		os.Exit(1)
	}
}
