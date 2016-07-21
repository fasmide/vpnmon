package gui

import (
	"fmt"
	"sort"
	"time"

	"github.com/fasmide/vpnmon/vpn"
	"github.com/gizak/termui"
)

type GUI struct {
	lastStatus     vpn.Status
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
		"[%13s] %-19s %15s %10s %10s %20s %16s",
		"Common Name",
		"Real Address",
		"Virtual Address",
		"BytesIn",
		"BytesOut",
		"Since",
		"Ping",
	)

	g.clientList = clientList

	// traffic widget
	lc2 := termui.NewLineChart()
	lc2.BorderLabel = "Traffic"
	// lc2.Data =
	lc2.Height = 10
	lc2.X = 0
	lc2.Y = 12
	lc2.AxesColor = termui.ColorWhite
	lc2.LineColor = termui.ColorCyan | termui.AttrBold

	// Generic info
	p := termui.NewPar(":PRESS q TO QUIT DEMO")
	p.Height = 10
	p.TextFgColor = termui.ColorWhite
	p.BorderLabel = "Open Vpn"
	p.BorderFg = termui.ColorCyan

	// build
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(6, 0, p),
			termui.NewCol(6, 0, lc2),
		),
		termui.NewRow(
			termui.NewCol(12, 0, clientList),
		),
	)

}

func (g *GUI) acceptEvents(events chan interface{}) {
	for e := range events {
		if event, ok := e.(vpn.Status); ok {
			termui.SendCustomEvt("/vpnupdate", event)
			continue
		}

		if pingUpdate, ok := e.(vpn.PingResponse); ok {
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

		g.lastStatus = e.Data.(vpn.Status)

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
		event := e.Data.(vpn.PingResponse)

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

func (g *GUI) renderClientLine(client *vpn.Client) string {
	ping := "N/A"
	if rtt, ok := g.lastPing[client.VirtualAddress]; ok {
		ping = rtt.String()
	}
	return fmt.Sprintf("[%13s] %19s %15s %10d %10d %20s %16s",
		client.CommonName,
		client.RealAddress,
		client.VirtualAddress,
		client.BytesReceived,
		client.BytesSent,
		time.Now().Sub(client.ConnectedSince),
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
