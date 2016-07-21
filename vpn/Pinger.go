package main

import (
	"net"
	"time"

	fastping "github.com/tatsushid/go-fastping"
)

type PingResponse struct {
	Rtt time.Duration
	Ip  *net.IPAddr
}

type Pinger struct {
	pinger    *fastping.Pinger
	pingerIps map[string]bool
	results   chan PingResponse
}

func NewPinger(channel chan PingResponse) *Pinger {
	fp := fastping.NewPinger()

	fp.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		channel <- PingResponse{Rtt: rtt, Ip: addr}
	}

	fp.OnIdle = func() {
		go fp.Run()
	}

	return &Pinger{
		pinger:    fp,
		pingerIps: make(map[string]bool),
		results:   channel,
	}
}

func (p *Pinger) StartLoop() {
	go p.pinger.Run()
}

func (p *Pinger) Add(addr string) {
	// I'm guessing this should have some locking on it
	if _, found := p.pingerIps[addr]; !found {
		p.pinger.AddIP(addr)
		p.pingerIps[addr] = true
	}
}
