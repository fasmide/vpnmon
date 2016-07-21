package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Status struct {
	ClientList map[string]*Client
	Updated    time.Time
	LoadStats  LoadStats
}

type LoadStats struct {
	BytesSent     uint64
	BytesReceived uint64
	NClients      uint64
}

type Client struct {
	CommonName     string
	RealAddress    string
	BytesSent      uint64
	BytesReceived  uint64
	ConnectedSince time.Time
	VirtualAddress string
	Updated        time.Time
}

// ref Mon Jan 2 15:04:05 -0700 MST 2006
// example: Mon Jun 20 14:00:59 2016
const timeConversion = "Mon Jan 2 15:04:05 2006"

func (l *LoadStats) Update(data string) {
	// this is what we are dealing with:
	// SUCCESS: nclients=2,bytesin=2356510,bytesout=2437636

	token := strings.Index(data, ",")
	nClients := data[strings.Index(data, "nclients=")+9 : token]

	// remove part..
	data = data[token+1:]

	token = strings.Index(data, ",")
	bytesIn := data[strings.Index(data, "bytesin=")+8 : token]

	data = data[token+1:]

	// we dont have a token for the last bit here
	bytesOut := data[strings.Index(data, "bytesout=")+9:]

	// conversion
	bytesOutInteger, err := strconv.ParseUint(bytesOut, 10, 64)

	if err != nil {
		panic(err)
	}

	bytesInInteger, err := strconv.ParseUint(bytesIn, 10, 64)

	if err != nil {
		panic(err)
	}

	nClientsInteger, err := strconv.ParseUint(nClients, 10, 64)

	if err != nil {
		panic(err)
	}

	// we made it!
	l.NClients = nClientsInteger
	l.BytesReceived = bytesInInteger
	l.BytesSent = bytesOutInteger

}

func NewStatus() *Status {
	return &Status{ClientList: make(map[string]*Client)}
}

// this method tries to parse openvpn-status, this is what we are dealing with:
// OpenVPN CLIENT LIST
// Updated,Tue Jun 21 15:51:38 2016
// Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
// kaelderspand,62.107.84.188:37000,1379573,755373,Mon Jun 20 14:49:40 2016
// raspberrypi3,62.107.84.188:39757,45477,88146,Tue Jun 21 14:35:41 2016
// ROUTING TABLE
// Virtual Address,Common Name,Real Address,Last Ref
// 10.8.88.6,kaelderspand,62.107.84.188:37000,Tue Jun 21 15:51:38 2016
// 10.8.88.10,raspberrypi3,62.107.84.188:39757,Tue Jun 21 14:39:16 2016
// GLOBAL STATS
// Max bcast/mcast queue length,2
// END
func (o *Status) Update(data []string) error {

	// this is going to be soooo hackish
	lookForUpdated := func(l []string) (bool, error) {
		if l[0] == "Updated" {

			t, err := time.Parse(timeConversion, l[1])

			if err != nil {
				return false, err
			}

			o.Updated = t
			return true, nil
		}
		return false, nil
	}

	lookForClients := func(l []string) (bool, error) {
		if l[0] == "ROUTING TABLE" {
			// thats it, there is no more for us
			return true, nil
		}

		if l[0] == "Common Name" {
			// this is the header we are looking for
			return false, nil
		}

		bytesIn, err := strconv.ParseUint(l[2], 10, 64)
		if err != nil {
			return false, err
		}

		bytesOut, err := strconv.ParseUint(l[3], 10, 64)
		if err != nil {
			return false, err
		}

		connectedSince, err := time.Parse(timeConversion, l[4])
		if err != nil {
			return false, err
		}

		if val, ok := o.ClientList[l[0]]; ok {
			// update existing val
			val.BytesReceived = bytesIn
			val.BytesSent = bytesOut
			val.Updated = time.Now()
			// These properly havnt changed anyways...
			// val.ConnectedSince = connectedSince
			// val.CommonName = l[0]
			// val.RealAddress = l[1]

		} else {
			// we have a new client
			o.ClientList[l[0]] = &Client{
				BytesReceived:  bytesIn,
				BytesSent:      bytesOut,
				ConnectedSince: connectedSince,
				CommonName:     l[0],
				RealAddress:    l[1],
				Updated:        time.Now(),
			}
		}
		// there could be more...
		return false, nil
	}

	lookForRoutingTable := func(l []string) (bool, error) {
		// our very first line
		if l[0] == "Virtual Address" {
			return false, nil
		}
		// This is our end line
		if l[0] == "GLOBAL STATS" {
			return true, nil
		}

		if val, ok := o.ClientList[l[1]]; ok {
			// update with virtual address
			val.VirtualAddress = l[0]
		} else {
			// there was no such client, but there should be as this parser runs after lookForClients
			return false, errors.New("A client appered in the routing table that was no client... huh?")
		}

		return false, nil

	}

	parserOrder := []func([]string) (bool, error){lookForUpdated, lookForClients, lookForRoutingTable}

	var parserIndex = 0

	for _, line := range data {

		l := strings.Split(line, ",")
		ok, err := parserOrder[parserIndex](l)

		if err != nil {
			panic(err)
		}

		if ok {
			// Move to the next parser...
			parserIndex++
			if len(parserOrder) <= parserIndex {
				break
			}
		}

	}

	// holy .... we made it!

	return nil

}
