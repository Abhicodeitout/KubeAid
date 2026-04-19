package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-debugger/pkg/analyzer"
	"kube-debugger/pkg/kubernetes"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Padding(0, 1)
	styleTab      = lipgloss.NewStyle().Padding(0, 2)
	styleActiveTab = lipgloss.NewStyle().Padding(0, 2).Bold(true).Foreground(lipgloss.Color("205")).Underline(true)
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleRed      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleHelp     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	styleBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	styleLog      = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

// ── tabs ─────────────────────────────────────────────────────────────────────

type tabView int

const (
	tabPods    tabView = iota
	tabLogs
	tabContexts
	tabCount
)

var tabNames = []string{"Pods", "Logs", "Contexts"}

// ── messages ─────────────────────────────────────────────────────────────────

type podsLoadedMsg struct{ pods []podEntry }
type logsLoadedMsg struct{ logs string }
type contextsLoadedMsg struct{ contexts []string }
type errMsg struct{ err error }

type podEntry struct {
	name      string
	namespace string
	status    string
	restarts  int32
}

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	activeTab    tabView
	pods         []podEntry
	podCursor    int
	logs         string
	contexts     []string
	contextCursor int
	namespace    string
	loading      bool
	err          error
	width        int
	height       int
}

func initialModel(namespace string) model {
	if namespace == "" {
		namespace = "default"
	}
	return model{namespace: namespace, loading: true}
}

// ── init / commands ───────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(loadPods(m.namespace), loadContexts())
}

func loadPods(namespace string) tea.Cmd {
	return func() tea.Msg {
		cs, err := kubernetes.GetKubeClient()
		if err != nil {
			return errMsg{err}
		}
		list, err := cs.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errMsg{err}
		}
		var pods []podEntry
		for _, p := range list.Items {
			status := string(p.Status.Phase)
			if len(p.Status.ContainerStatuses) > 0 && p.Status.ContainerStatuses[0].State.Waiting != nil {
				status = p.Status.ContainerStatuses[0].State.Waiting.Reason
			}
			var rc int32
			if len(p.Status.ContainerStatuses) > 0 {
				rc = p.Status.ContainerStatuses[0].RestartCount
			}
			pods = append(pods, podEntry{
				name:      p.Name,
				namespace: namespace,
				status:    status,
				restarts:  rc,
			})
		}
		return podsLoadedMsg{pods}
	}
}

func loadContexts() tea.Cmd {
	return func() tea.Msg {
		ctxs, err := kubernetes.ListKubeContexts()
		if err != nil {
			return contextsLoadedMsg{} // tolerate missing kubeconfig
		}
		return contextsLoadedMsg{ctxs}
	}
}

func loadLogs(namespace, podName string) tea.Cmd {
	return func() tea.Msg {
		cs, err := kubernetes.GetKubeClient()
		if err != nil {
			return logsLoadedMsg{"error: " + err.Error()}
		}
		logs, err := kubernetes.GetPodLogs(cs, namespace, podName)
		if err != nil {
			return logsLoadedMsg{"error: " + err.Error()}
		}
		return logsLoadedMsg{logs}
	}
}

// ── update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case podsLoadedMsg:
		m.pods = msg.pods
		m.loading = false

	case logsLoadedMsg:
		m.logs = msg.logs

	case contextsLoadedMsg:
		m.contexts = msg.contexts

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % tabCount

		case "shift+tab", "left":
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount

		case "up", "k":
			switch m.activeTab {
			case tabPods:
				if m.podCursor > 0 {
					m.podCursor--
				}
			case tabContexts:
				if m.contextCursor > 0 {
					m.contextCursor--
				}
			}

		case "down", "j":
			switch m.activeTab {
			case tabPods:
				if m.podCursor < len(m.pods)-1 {
					m.podCursor++
				}
			case tabContexts:
				if m.contextCursor < len(m.contexts)-1 {
					m.contextCursor++
				}
			}

		case "enter", "l":
			switch m.activeTab {
			case tabPods:
				if len(m.pods) > 0 {
					pod := m.pods[m.podCursor]
					m.activeTab = tabLogs
					m.logs = "Loading logs…"
					return m, loadLogs(pod.namespace, pod.name)
				}
			case tabContexts:
				if len(m.contexts) > 0 {
					ctx := m.contexts[m.contextCursor]
					_ = kubernetes.SwitchKubeContext(ctx)
					m.loading = true
					return m, loadPods(m.namespace)
				}
			}

		case "r":
			m.loading = true
			return m, loadPods(m.namespace)
		}
	}

	return m, nil
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	// Title bar
	b.WriteString(styleBorder.Render(styleTitle.Render("  KubeAid TUI  ·  ns: "+m.namespace)) + "\n")

	// Tab bar
	tabs := ""
	for i, name := range tabNames {
		if tabView(i) == m.activeTab {
			tabs += styleActiveTab.Render(name)
		} else {
			tabs += styleTab.Render(name)
		}
		if i < len(tabNames)-1 {
			tabs += styleDim.Render(" │ ")
		}
	}
	b.WriteString(tabs + "\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", 60)) + "\n")

	if m.err != nil {
		b.WriteString(styleRed.Render("❌  "+m.err.Error()) + "\n")
		b.WriteString(styleHelp.Render("q quit") + "\n")
		return b.String()
	}

	if m.loading {
		b.WriteString(styleDim.Render("  Loading…") + "\n")
		b.WriteString(styleHelp.Render("q quit") + "\n")
		return b.String()
	}

	switch m.activeTab {
	case tabPods:
		b.WriteString(renderPods(m))
	case tabLogs:
		b.WriteString(renderLogs(m))
	case tabContexts:
		b.WriteString(renderContexts(m))
	}

	// Help bar
	b.WriteString("\n" + styleHelp.Render("↑↓/jk navigate  enter/l select  tab next-tab  r refresh  q quit") + "\n")
	return b.String()
}

func renderPods(m model) string {
	if len(m.pods) == 0 {
		return styleDim.Render("  No pods found in namespace '"+m.namespace+"'") + "\n"
	}
	var b strings.Builder
	b.WriteString(styleDim.Render(fmt.Sprintf("  %-52s %-16s %s\n", "POD", "STATUS", "RESTARTS")))
	for i, p := range m.pods {
		line := fmt.Sprintf("  %-52s %-16s %d", p.name, p.status, p.restarts)
		if i == m.podCursor {
			b.WriteString(styleSelected.Render("> "+strings.TrimSpace(line)) + "\n")
		} else {
			b.WriteString(styleDim.Render(line) + "\n")
		}
	}
	b.WriteString("\n" + styleHelp.Render("  Press enter to view logs for selected pod") + "\n")
	return b.String()
}

func renderLogs(m model) string {
	if m.logs == "" {
		return styleDim.Render("  Select a pod in the Pods tab (press enter) to load its logs") + "\n"
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(m.logs), "\n") {
		ll := strings.ToLower(line)
		if strings.Contains(ll, "error") || strings.Contains(ll, "fatal") || strings.Contains(ll, "panic") {
			b.WriteString(styleRed.Render("  "+line) + "\n")
		} else {
			b.WriteString(styleLog.Render("  "+line) + "\n")
		}
	}
	return b.String()
}

func renderContexts(m model) string {
	if len(m.contexts) == 0 {
		return styleDim.Render("  No kubeconfig contexts found") + "\n"
	}
	var b strings.Builder
	b.WriteString(styleDim.Render("  Available kubeconfig contexts:\n\n"))
	for i, ctx := range m.contexts {
		if i == m.contextCursor {
			b.WriteString(styleSelected.Render("  > "+ctx) + "\n")
		} else {
			b.WriteString(styleDim.Render("    "+ctx) + "\n")
		}
	}
	b.WriteString("\n" + styleHelp.Render("  Press enter to switch context") + "\n")
	return b.String()
}

// ── entry point ───────────────────────────────────────────────────────────────

func StartTUI(namespace string) {
	// Try to load a quick analysis summary before launching
	_ = analyzer.AnalyzeApp // pre-warm import

	p := tea.NewProgram(initialModel(namespace), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running TUI:", err)
		os.Exit(1)
	}
}

