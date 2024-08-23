package statusbar

import (
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	barStyle    lipgloss.Style
	prefix      string
	prefixStyle lipgloss.Style
	msg         string
	msgStyle    lipgloss.Style
	suffix      string
	suffixStyle lipgloss.Style
	w           int
	h           int
}

func (m *Model) UpdatePrefix(v string) {
	m.prefix = v
}

func (m *Model) UpdateMessage(v string) {
	m.msg = v
}

func (m *Model) UpdateSuffix(v string) {
	m.suffix = v
}

func (m *Model) UpdatePrefixStyle(style lipgloss.Style) {
	m.prefixStyle = style
}

func (m *Model) UpdateMessageStyle(style lipgloss.Style) {
	m.msgStyle = style
}

func (m *Model) UpdateSuffixStyle(style lipgloss.Style) {
	m.suffixStyle = style
}

func (m Model) Init() tea.Cmd {
	return nil
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.barStyle = m.barStyle.Width(m.w)
	}
	return m, nil
}

func (m Model) View() string {
	prefixView := m.prefixStyle.Render(m.prefix)
	suffixView := m.suffixStyle.Render(m.suffix)
	msgW := m.w - lipgloss.Width(prefixView) - lipgloss.Width(suffixView)
	msgView := m.msgStyle.Padding(0, 1).Width(msgW).Render(m.msg)
	return m.barStyle.Render(prefixView + msgView + suffixView)
}

func New() Model {
	barStyle := lipgloss.NewStyle().Background(constants.ColorNormalInv).Margin(1, 0).Height(1)
	return Model{
		barStyle:    barStyle,
		prefix:      "TAILSCALE-TUI",
		prefixStyle: constants.PrimaryTitleStyle,
		msg:         "",
		msgStyle:    lipgloss.NewStyle().Background(barStyle.GetBackground()).Foreground(constants.ColorBW),
		suffix:      "",
		suffixStyle: constants.SecondaryTitleStyle,
	}
}
