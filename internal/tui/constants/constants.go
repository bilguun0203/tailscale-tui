package constants

import "github.com/charmbracelet/lipgloss"

var ColorBW = lipgloss.AdaptiveColor{Light: "#000", Dark: "#FFF"}
var ColorNormal = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}
var ColorNormalInv = lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#1A1A1A"}
var ColorDanger = lipgloss.AdaptiveColor{Light: "#ff005f", Dark: "#ff005f"}
var ColorSuccess = lipgloss.AdaptiveColor{Light: "#00835a", Dark: "#00ffaf"}
var ColorWarning = lipgloss.AdaptiveColor{Light: "#e9a000", Dark: "ffaf00"}
var ColorDimmed = lipgloss.AdaptiveColor{Light: "#7c737c", Dark: "#777777"}
var ColorMuted = lipgloss.AdaptiveColor{Light: "#6e5e6e", Dark: "#4D4D4D"}
var ColorPrimary = lipgloss.AdaptiveColor{Light: "#008448", Dark: "#20F394"}
var ColorSecondary = lipgloss.AdaptiveColor{Light: "#007c88", Dark: "#00E5FA"}

var PrimaryTitleStyle = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color(ColorPrimary.Dark)).Foreground(lipgloss.Color("#000000"))
var SecondaryTitleStyle = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color(ColorSecondary.Light)).Foreground(lipgloss.Color("#000000"))
var WarningTitleStyle = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color(ColorWarning.Light)).Foreground(lipgloss.Color("#000000"))
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
