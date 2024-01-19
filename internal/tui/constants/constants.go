package constants

import "github.com/charmbracelet/lipgloss"

var TitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("#20F394")).Foreground(lipgloss.Color("#000000"))
var AltTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("#05EAFF")).Foreground(lipgloss.Color("#000000"))
var NormalTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"})
var DangerTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "197", Dark: "197"})
var SuccessTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "034", Dark: "049"})
var WarningTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "214", Dark: "214"})
var MutedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})
var AccentTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"})
var AltAccentTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#05EAFF", Dark: "#00E5FA"})
var LayoutStyle = lipgloss.NewStyle()
var SpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var HeaderStyle = lipgloss.NewStyle().Margin(1, 2)
