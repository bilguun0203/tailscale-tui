package main

import (
	"context"
	"fmt"
	"log"

	"tailscale.com/client/tailscale"
)

func main() {
	ctx := context.Background()
	status, err := tailscale.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Me:\n - %s : %s\n", status.Self.HostName, status.Self.TailscaleIPs)
	fmt.Println("Nodes:")
	for _, v := range status.Peer {
		state := "OFF"
		if v.Online {
			state = "ON"
		}
		fmt.Printf(" - %s (%s) : %s <%s>\n", v.HostName, state, status.User[v.UserID].DisplayName, status.User[v.UserID].LoginName)
		for _, ip := range v.TailscaleIPs {
			fmt.Printf("    - %s\n", ip)
		}
	}
}
