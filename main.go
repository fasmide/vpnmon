package main

import (
	"github.com/fasmide/vpnmon/gui"
	"github.com/fasmide/vpnmon/vpn"
)

func main() {

	m, err := vpn.NewManagement()
	if err != nil {
		panic(err)
	}
	vpnUpdates := make(chan vpn.Status)
	go m.UpdateLoop(vpnUpdates)

	pingUpdates := make(chan vpn.PingResponse)
	pinger := vpn.NewPinger(pingUpdates)
	pinger.StartLoop()

	deMux := make(chan interface{})

	g := gui.NewGUI()

	go func() {

		for {
			select {
			case update := <-vpnUpdates:
				// send this update to GUI
				deMux <- update

				// Look for new clients we need added to fastping
				for _, client := range update.ClientList {
					pinger.Add(client.VirtualAddress)
				}
			case update := <-pingUpdates:
				deMux <- update
			}
		}

	}()
	g.Loop(deMux)

}
