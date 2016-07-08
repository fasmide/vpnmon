package main

func main() {

	m, err := NewManagement()
	if err != nil {
		panic(err)
	}
	vpnUpdates := make(chan Status)
	go m.UpdateLoop(vpnUpdates)

	pingUpdates := make(chan PingResponse)
	pinger := NewPinger(pingUpdates)
	pinger.StartLoop()

	deMux := make(chan interface{})

	g := NewGUI()

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
