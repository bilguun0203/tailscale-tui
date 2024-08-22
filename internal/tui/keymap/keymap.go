package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	CopyIpv4      key.Binding
	CopyIpv6      key.Binding
	CopyDNSName   key.Binding
	Refresh       key.Binding
	Enter         key.Binding
	Back          key.Binding
	Quit          key.Binding
	ForceQuit     key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
	TSUp          key.Binding
	TSDown        key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CopyIpv4, k.Back, k.Quit, k.ShowFullHelp}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CopyIpv4, k.CopyIpv6, k.CopyDNSName},
		{k.Back, k.Quit, k.CloseFullHelp},
	}
}

func NewKeyMap() KeyMap {
	return KeyMap{
		CopyIpv4: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy ipv4"),
		),
		CopyIpv6: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("Y", "copy ipv6"),
		),
		CopyDNSName: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy domain"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "details"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		TSUp: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "up"),
		),
		TSDown: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "down"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
