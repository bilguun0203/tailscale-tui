package tui

import (
	"fmt"
	"time"

	"github.com/bilguun0203/tailscale-tui/internal/ts"
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	nodedetails "github.com/bilguun0203/tailscale-tui/internal/tui/node_details"
	nodelist "github.com/bilguun0203/tailscale-tui/internal/tui/node_list"
	statusbar "github.com/bilguun0203/tailscale-tui/internal/tui/status_bar"
	"github.com/bilguun0203/tailscale-tui/internal/tui/types"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/ipn/ipnstate"
	tsKey "tailscale.com/types/key"
)

type viewState int

const (
	viewStateList viewState = iota
	viewStateDetails
)

func (f viewState) String() string {
	return [...]string{
		"list",
		"details",
	}[f]
}

type Model struct {
	viewState      viewState
	tsStatus       *ipnstate.Status
	selectedNodeID tsKey.NodePublic
	isLoading      bool
	Err            error
	ExitMessage    string
	nodelist       nodelist.Model
	nodedetails    nodedetails.Model
	statusbar      statusbar.Model
	spinner        spinner.Model
	w, h           int
	statusH        int
	headerH        int
}

func (m Model) getTsStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := ts.GetStatus()
		if err != nil {
			return ts.StatusErrorMsg(err)
		}
		return ts.StatusDataMsg(status)
	}
}

func (m Model) headerView() string {
	if m.tsStatus == nil {
		return nodedetails.NodeDetailRender(nil, tsKey.NodePublic{}, constants.PrimaryTitleStyle.Render("Current Node"))
	}
	return nodedetails.NodeDetailRender(m.tsStatus, m.tsStatus.Self.PublicKey, constants.PrimaryTitleStyle.Render("Current Node"))
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		m.getTsStatus(),
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tmpCmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ts.StatusDataMsg:
		m.isLoading = false
		m.Err = nil
		m.tsStatus = msg
		m.viewState = viewStateList
		cmds = append(cmds, types.NewStatusMsg("Showing all network devices"))
	case ts.StatusErrorMsg:
		m.isLoading = false
		m.Err = msg
		return m, tea.Quit
	case ts.ToggleConnectionMsg:
		if m.tsStatus != nil {
			newStatus := !m.tsStatus.Self.Online
			m.isLoading = true
			if newStatus {
				cmds = append(cmds, types.NewStatusMsg("Connecting..."))
			} else {
				cmds = append(cmds, types.NewStatusMsg("Disconnecting..."))
			}
			ts.SetTSStatus(!m.tsStatus.Self.Online)
			cmds = append(cmds, func() tea.Msg { time.Sleep(2 * time.Second); return types.RefreshMsg(true) })
		}
	case nodedetails.BackMsg:
		m.viewState = viewStateList
		cmds = append(cmds, types.NewStatusMsg("Showing all network devices"))
		cmds = append(cmds, tea.ClearScreen)
	case types.RefreshMsg:
		m.isLoading = true
		cmds = append(cmds, m.getTsStatus())
		cmds = append(cmds, m.spinner.Tick)
	case types.StatusMsg:
		m.statusbar.UpdateMessage(string(msg))
	case types.ExitMsg:
		m.ExitMessage = string(msg)
		return m, tea.Quit
	case nodelist.NodeSelectedMsg:
		m.selectedNodeID = tsKey.NodePublic(msg)
		contentH := m.h - m.statusH
		m.nodedetails = nodedetails.New(m.tsStatus, m.selectedNodeID, m.w, contentH)
		m.viewState = viewStateDetails
		cmds = append(cmds, types.NewStatusMsg("Showing device details"))
		cmds = append(cmds, tea.ClearScreen)
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.headerH = lipgloss.Height(m.headerView())
		m.statusH = lipgloss.Height(m.statusbar.View())
		contentH := m.h - m.headerH - m.statusH
		m.nodelist.SetSize(m.w, contentH)
		m.nodedetails.SetSize(m.w, contentH)
	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, tmpCmd = m.spinner.Update(msg)
			cmds = append(cmds, tmpCmd)
		}
	}

	if m.isLoading {
		m.statusbar.UpdatePrefixStyle(constants.WarningTitleStyle)
	} else {
		m.statusbar.UpdatePrefixStyle(constants.PrimaryTitleStyle)
	}

	switch m.viewState {
	case viewStateDetails:
		m.nodedetails, tmpCmd = m.nodedetails.Update(msg)
		cmds = append(cmds, tmpCmd)
	case viewStateList:
		if m.isLoading {
			m.spinner, tmpCmd = m.spinner.Update(msg)
			cmds = append(cmds, tmpCmd)
		} else {
			m.nodelist, tmpCmd = m.nodelist.Update(msg)
			cmds = append(cmds, tmpCmd)
		}
	}
	m.statusbar, tmpCmd = m.statusbar.Update(msg)
	cmds = append(cmds, tmpCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.viewState {
	case viewStateDetails:
		if m.isLoading {
			m.statusbar.UpdateMessage(fmt.Sprintf("%s %s", m.spinner.View(), m.statusbar.Message()))
		}
		return lipgloss.JoinVertical(lipgloss.Left, m.nodedetails.View(), m.statusbar.View())
	case viewStateList:
		if m.isLoading {
			m.statusbar.UpdateMessage(fmt.Sprintf("%s Loading...", m.spinner.View()))
		}
		return lipgloss.JoinVertical(lipgloss.Left, m.headerView(), m.nodelist.View(), m.statusbar.View())
	default:
		return "*_*"
	}
}

func New() Model {
	m := Model{
		viewState: viewStateList,
		isLoading: true,
		spinner:   spinner.New(),
		statusbar: statusbar.New(),
	}
	m.spinner.Spinner = spinner.Line
	m.spinner.Style = constants.SpinnerStyle

	m.headerH = lipgloss.Height(m.headerView())
	m.statusH = lipgloss.Height(m.statusbar.View())
	contentH := m.h - m.headerH - m.statusH
	m.nodelist = nodelist.New(nil, m.w, contentH)
	m.nodedetails = nodedetails.New(m.tsStatus, tsKey.NodePublic{}, m.w, contentH)
	return m
}
