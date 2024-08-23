package ts

import "tailscale.com/ipn/ipnstate"

type StatusDataMsg *ipnstate.Status

type StatusErrorMsg error
