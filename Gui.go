package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/gizak/termui"
)

type GUI struct {
	lastStatus     Status
	lastPing       map[string]time.Duration
	clientKeyOrder []string

	clientList *termui.List
}

func NewGUI() *GUI {

	return &GUI{lastPing: make(map[string]time.Duration)}
}

func (g *GUI) Init() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}

	clientList := termui.NewList()
	strs := []string{
		"[N/A] Finding clients....",
	}

	clientList.Items = strs
	clientList.ItemFgColor = termui.ColorYellow
	clientList.Height = termui.TermHeight()
	clientList.Border = true
	clientList.BorderLabel = fmt.Sprintf(
		"[%13s] %-19s %15s %10s %10s %15s",
		"Common Name",
		"Real Address",
		"Virtual Address",
		"BytesIn",
		"BytesOut",
		"Ping",
	)

	g.clientList = clientList

	// build
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(12, 0, clientList),
		),
	)

}

func (g *GUI) acceptEvents(events chan interface{}) {
	for e := range events {
		if event, ok := e.(Status); ok {
			termui.SendCustomEvt("/vpnupdate", event)
			continue
		}

		if pingUpdate, ok := e.(PingResponse); ok {
			termui.SendCustomEvt("/pingupdate", pingUpdate)
			continue
		}
		panic(fmt.Sprintf("I do not understand %T types", e))

	}
}

func (g *GUI) Loop(events chan interface{}) {
	defer termui.Close()
	g.Init()

	go g.acceptEvents(events)

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/vpnupdate", func(e termui.Event) {

		g.lastStatus = e.Data.(Status)

		// apply some sorting
		mk := make([]string, len(g.lastStatus.ClientList))
		i := 0
		for k, _ := range g.lastStatus.ClientList {
			mk[i] = k
			i++
		}
		sort.Strings(mk)
		g.clientKeyOrder = mk

		g.renderClientLines()
		termui.Render(termui.Body)
	})

	termui.Handle("/pingupdate", func(e termui.Event) {
		event := e.Data.(PingResponse)

		g.lastPing[event.Ip.String()] = event.Rtt
		g.renderClientLines()
		termui.Render(termui.Body)
	})

	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		g.clientList.Height = termui.TermHeight()
		termui.Render(termui.Body)
	})

	// calculate layout
	termui.Body.Align()
	termui.Render(termui.Body)

	termui.Loop()
}

func (g *GUI) renderClientLine(client *Client) string {
	ping := "N/A"
	if rtt, ok := g.lastPing[client.VirtualAddress]; ok {
		ping = rtt.String()
	}

	return fmt.Sprintf("[%13s] %19s %15s %10d %10d %15s",
		client.CommonName,
		client.RealAddress,
		client.VirtualAddress,
		client.BytesReceived,
		client.BytesSent,
		ping,
	)
}
func (g *GUI) renderClientLines() {
	clientStrings := make([]string, 0, 5)
	for _, clientKey := range g.clientKeyOrder {

		clientStrings = append(clientStrings, g.renderClientLine(g.lastStatus.ClientList[clientKey]))
	}

	g.clientList.Items = clientStrings

}
