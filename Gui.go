package main

import (
	"fmt"

	"github.com/gizak/termui"
)

type GUI struct {

	// not sure what goes in here
	clientList *termui.List
}

func NewGUI() *GUI {

	return &GUI{}
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
	clientList.BorderLabel = fmt.Sprintf("[%13s] %-19s %15s %10s %10s", "Common Name", "Real Address", "Virtual Address", "BytesIn", "BytesOut")

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
		clients := make([]string, 0, 5)
		for _, client := range e.Data.(Status).ClientList {

			clients = append(clients, renderClientLine(client))
		}
		g.clientList.Items = clients
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

func renderClientLine(client *Client) string {
	return fmt.Sprintf("[%13s] %19s %15s %10d %10d",
		client.CommonName,
		client.RealAddress,
		client.VirtualAddress,
		client.BytesReceived,
		client.BytesSent,
	)
}
