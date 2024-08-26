package ts

import (
	"net/netip"

	"tailscale.com/ipn/ipnstate"
)

type StatusDataMsg *ipnstate.Status
type StatusErrorMsg error
type ConnectMsg bool
type ToggleConnectionMsg bool
type PingMsg netip.Addr

type ActionType int

const (
	ConnectAction ActionType = iota
	OfferExitNode
	PingAction
)

func (f ActionType) String() string {
	return [...]string{
		"TSConnect",
		"TSPing",
	}[f]
}
