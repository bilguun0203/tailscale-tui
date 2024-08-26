package actionlist

import (
	"github.com/bilguun0203/tailscale-tui/internal/ts"
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionListItem struct {
	title, desc string
	value       ts.ActionType
}

func NewActionListItem(title string, desc string, value ts.ActionType) ActionListItem {
	return ActionListItem{title: title, desc: desc, value: value}
}
func (i ActionListItem) Title() string        { return i.title }
func (i ActionListItem) Description() string  { return i.desc }
func (i ActionListItem) Value() ts.ActionType { return i.value }
func (i ActionListItem) FilterValue() string  { return i.title + " " + i.desc }

type Model struct {
	list list.Model
}

func (m *Model) updateKeybindings() {
	m.list.KeyMap.NextPage.SetEnabled(false)
	m.list.KeyMap.PrevPage.SetEnabled(false)
	m.list.KeyMap.ShowFullHelp.SetEnabled(false)
	m.list.KeyMap.CloseFullHelp.SetEnabled(false)
	m.list.KeyMap.GoToStart.SetEnabled(false)
	m.list.KeyMap.GoToEnd.SetEnabled(false)
	m.list.KeyMap.Filter.SetEnabled(false)
	m.list.KeyMap.ClearFilter.SetEnabled(false)
	m.list.KeyMap.CancelWhileFiltering.SetEnabled(false)
	m.list.KeyMap.AcceptWhileFiltering.SetEnabled(false)
	m.list.KeyMap.Quit.SetEnabled(false)
	m.list.KeyMap.ForceQuit.SetEnabled(false)
}

func (m *Model) SetSize(w int, h int) {
	m.list.SetSize(w, h)
}

func (m Model) SelectedItem() ActionListItem {
	return m.list.SelectedItem().(ActionListItem)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.NewStyle().Width(m.list.Width()).Height(m.list.Height()).Render(m.list.View())
}

func New(items []ActionListItem, w int, h int) Model {
	d := list.NewDefaultDelegate()
	descStyle := lipgloss.NewStyle().Margin(0, 0, 0, 4)
	d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(constants.ColorNormal).Margin(0, 0, 0, 2)
	d.Styles.NormalDesc = descStyle.Foreground(constants.ColorDimmed)
	d.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(constants.ColorPrimary).Margin(0, 0, 0, 3)
	d.Styles.SelectedDesc = descStyle.Foreground(constants.ColorWarning).Margin(0, 0, 0, 5)
	d.Styles.DimmedTitle = constants.DimmedTextStyle.Margin(0, 0, 0, 2)
	d.Styles.DimmedDesc = descStyle.Foreground(constants.ColorMuted)

	lis := []list.Item{}
	for _, item := range items {
		lis = append(lis, item)
	}

	m := Model{
		list: list.New(lis, d, w, h),
	}
	m.list.Title = "Actions"
	m.list.Styles.Title = constants.PrimaryTitleStyle
	m.list.SetFilteringEnabled(false)
	m.list.SetShowFilter(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(false)
	m.list.SetShowStatusBar(false)
	m.updateKeybindings()
	return m
}
