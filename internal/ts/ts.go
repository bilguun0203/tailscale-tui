package ts

import (
	"context"

	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
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
