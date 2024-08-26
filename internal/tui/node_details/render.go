package nodedetails

import (
	"fmt"
	"strings"
	"time"

	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/ipn/ipnstate"
	tsKey "tailscale.com/types/key"
)

func NodeDetailRender(tsStatus *ipnstate.Status, nodeID tsKey.NodePublic, customTitle string) string {
	title := constants.PrimaryTitleStyle.Render("Node info")
	if customTitle != "" {
		title = customTitle
	}
	status := constants.SecondaryTextStyle.Render("Status: ")
	hostname := constants.SecondaryTextStyle.Render("Host: ")
	userInfo := "??? <???>"
	ips := constants.SecondaryTextStyle.Render("IPs: ")
	relay := constants.SecondaryTextStyle.Render("Relay: ")
	offersExitNode := "no"
	exitNode := constants.SecondaryTextStyle.Render("Exit node: ")
	asExitNode := ""
	keyExpiry := constants.SecondaryTextStyle.Render("Key expiry: ")
	currentDevice := false
	if tsStatus != nil {
		node, ok := tsStatus.Peer[nodeID]
		if !ok && tsStatus.Self.PublicKey == nodeID {
			node = tsStatus.Self
			currentDevice = true
			ok = true
		}
		if ok {
			if user, ok := tsStatus.User[node.UserID]; ok {
				userInfo = fmt.Sprintf("%s <%s>", user.DisplayName, user.LoginName)
			} else {
				userInfo = fmt.Sprintf("??? <%d>", node.UserID)
			}
			if node.ExitNodeOption {
				offersExitNode = constants.WarningTextStyle.Render("yes")
			}
			var ipList []string
			for _, ip := range node.TailscaleIPs {
				ipList = append(ipList, ip.String())
			}
			if node.Online {
				status += constants.SuccessTextStyle.Render("Online")
			} else {
				status += constants.DangerTextStyle.Render("Offline")
			}
			if node.KeyExpiry == nil {
				keyExpiry += "Disabled"
			} else {
				if node.Expired {
					keyExpiry += constants.DangerTextStyle.Render("Expired ")
				} else {
					keyExpiry += "Active "
				}
				keyExpiry += constants.DimmedTextStyle.Render("(" + node.KeyExpiry.Local().Format(time.RFC3339) + ")")
			}
			ipList = append(ipList, node.DNSName)
			ips += strings.Join(ipList, " | ")
			hostname += node.HostName + " (" + node.OS + ")"
			if currentDevice {
				hostname += " " + constants.DimmedTextStyle.Render("*This device*")
			}
			relay += node.Relay
			exitNode += constants.DimmedTextStyle.Render("offers: ") + offersExitNode
			if node.ExitNode {
				asExitNode = constants.WarningTextStyle.Render("~ This node is currently being used as an exit node.")
			}
			if currentDevice && tsStatus.ExitNodeStatus != nil {
				for _, peer := range tsStatus.Peer {
					if peer.ID == tsStatus.ExitNodeStatus.ID {
						exitNode += constants.DimmedTextStyle.Render(" / using: ") + constants.WarningTextStyle.Render(peer.HostName)
						break
					}
				}
			}
		}
	}
	body := lipgloss.JoinVertical(lipgloss.Left, userInfo+"\n", hostname, status, ips, relay, keyExpiry, exitNode, asExitNode)
	return constants.HeaderStyle.Render(fmt.Sprintf("%s\n\n%s", title, body))
}
