package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/repos"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AF87"))

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type state int

const (
	stateList state = iota
	stateDetail
)

type item struct {
	cmd models.Command
}

func (i item) Title() string       { return i.cmd.Command }
func (i item) Description() string { return strings.Join(i.cmd.Arguments, " ") }
func (i item) FilterValue() string { return i.cmd.Command + " " + strings.Join(i.cmd.Arguments, " ") }

type itemDelegate struct {
	defaultDelegate list.DefaultDelegate
}

func newItemDelegate() itemDelegate {
	d := list.NewDefaultDelegate()
	return itemDelegate{defaultDelegate: d}
}

func (d itemDelegate) Height() int                               { return 2 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("> " + title)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("  " + title)
	}

	status := " ✓ "
	if !i.cmd.Status {
		status = " ✗ "
		status = statusStyle.Render(status)
	} else {
		status = successStyle.Render(status)
	}

	fmt.Fprintf(w, "%s%s\n", status, title)
	fmt.Fprintf(w, "     %s", dimStyle.Render(desc))
}

type Model struct {
	list     list.Model
	viewport viewport.Model
	repo     repos.RepositoryInterface
	choice   *models.Command
	state    state
	quitting bool
	width    int
	height   int
}

func NewModel(repo repos.RepositoryInterface) Model {
	commands, _ := repo.GetAllCommands()
	items := make([]list.Item, len(commands))
	for i, c := range commands {
		items[i] = item{cmd: c}
	}

	l := list.New(items, newItemDelegate(), 0, 0)
	l.Title = "Ambros Dashboard"
	l.SetFilteringEnabled(true)
	l.FilterInput.Placeholder = "Search commands..."
	
	// Enable fuzzy filtering
	l.KeyMap.Filter = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rerun")),
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
		}
	}

	return Model{
		list:  l,
		repo:  repo,
		state: stateList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)
		m.viewport = viewport.New(msg.Width-8, msg.Height-12)
		return m, nil

	case tea.KeyMsg:
		if m.state == stateDetail {
			switch msg.String() {
			case "esc", "backspace":
				m.state = stateList
				return m, nil
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &i.cmd
				m.state = stateDetail
				m.viewport.SetContent(fmt.Sprintf(
					"ID: %s\nStatus: %v\nTime: %s\n\nCommand:\n%s %s\n\nOutput:\n%s\n\nError:\n%s",
					m.choice.ID,
					m.choice.Status,
					m.choice.CreatedAt.Format("2006-01-02 15:04:05"),
					m.choice.Command,
					strings.Join(m.choice.Arguments, " "),
					m.choice.Output,
					m.choice.Error,
				))
			}
			return m, nil

		case "x":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				_ = m.repo.Delete(i.cmd.ID)
				commands, _ := m.repo.GetAllCommands()
				items := make([]list.Item, len(commands))
				for i, c := range commands {
					items[i] = item{cmd: c}
				}
				m.list.SetItems(items)
			}
			return m, nil

		case "r":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &i.cmd
				return m, tea.Quit
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.state == stateDetail {
		title := titleStyle.Render("Command Details")
		status := "✓ SUCCESS"
		if !m.choice.Status {
			status = "✗ FAILED"
			status = statusStyle.Render(status)
		} else {
			status = successStyle.Render(status)
		}

		header := fmt.Sprintf("%s %s\n\n", title, status)
		return docStyle.Render(header + detailStyle.Render(m.viewport.View()) + "\n\n(esc to back)")
	}

	return docStyle.Render(m.list.View())
}

func (m Model) GetChoice() *models.Command {
	return m.choice
}
