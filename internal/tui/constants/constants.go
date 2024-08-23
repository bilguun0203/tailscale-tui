package constants

import "github.com/charmbracelet/lipgloss"

var ColorNormal = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}
var ColorDanger = lipgloss.AdaptiveColor{Light: "197", Dark: "197"}
var ColorSuccess = lipgloss.AdaptiveColor{Light: "034", Dark: "049"}
var ColorWarning = lipgloss.AdaptiveColor{Light: "214", Dark: "214"}
var ColorDimmed = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}
var ColorMuted = lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}
var ColorPrimary = lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"}
var ColorSecondary = lipgloss.AdaptiveColor{Light: "#05EAFF", Dark: "#00E5FA"}

var PrimaryTitleStyle = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color(ColorPrimary.Dark)).Foreground(lipgloss.Color("#000000"))
var SecondaryTitleStyle = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color(ColorSecondary.Light)).Foreground(lipgloss.Color("#000000"))
var NormalTextStyle = lipgloss.NewStyle().Foreground(ColorNormal)
var DangerTextStyle = lipgloss.NewStyle().Foreground(ColorDanger)
var SuccessTextStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
var WarningTextStyle = lipgloss.NewStyle().Foreground(ColorWarning)
var DimmedTextStyle = lipgloss.NewStyle().Foreground(ColorDimmed)
var MutedTextStyle = lipgloss.NewStyle().Foreground(ColorMuted)
var PrimaryTextStyle = lipgloss.NewStyle().Foreground(ColorPrimary)
var SecondaryTextStyle = lipgloss.NewStyle().Foreground(ColorSecondary)
var LayoutStyle = lipgloss.NewStyle()
var SpinnerStyle = lipgloss.NewStyle().Foreground(ColorPrimary)
var HeaderStyle = lipgloss.NewStyle().Margin(1, 2)
