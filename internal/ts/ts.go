package ts

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
)

var lc tailscale.LocalClient

func GetStatus() (*ipnstate.Status, error) {
	return lc.Status(context.Background())
}

// Connect/Disconnect Tailscale network.
// Equivalent to `tailscale up` (status=true) `tailscale down` (status=false) commands
func SetTSStatus(status bool) {
	lc.EditPrefs(context.Background(), &ipn.MaskedPrefs{
		Prefs: ipn.Prefs{
			WantRunning: status,
		},
		WantRunningSet: true,
	})
}

func Ping(ip netip.Addr) (*ipnstate.PingResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	pr, err := lc.Ping(ctx, ip, tailcfg.PingDisco)
	cancel()
	return pr, err
}

func PingResultString(pr *ipnstate.PingResult) (string, error) {
	if pr == nil {
		return "", nil
	}
	var message string
	if pr.Err != "" {
		if pr.IsLocalIP {
			message = "local ip"
			return message, nil
		}
		return message, errors.New(pr.Err)
	}
	latency := time.Duration(pr.LatencySeconds * float64(time.Second)).Round(time.Millisecond)
	via := pr.Endpoint
	if pr.DERPRegionID != 0 {
		via = fmt.Sprintf("DERP(%s)", pr.DERPRegionCode)
	}
	if via == "" {
		via = string(tailcfg.PingDisco)
	}
	extra := ""
	if pr.PeerAPIPort != 0 {
		extra = fmt.Sprintf(", %d", pr.PeerAPIPort)
	}
	message = fmt.Sprintf("pong from %s (%s%s) via %v in %v", pr.NodeName, pr.NodeIP, extra, via, latency)
	return message, nil
}
