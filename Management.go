package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"
)

type Management struct {
	conn     net.Conn
	reader   *bufio.Reader
	myStatus *Status
}

func NewManagement() (*Management, error) {
	conn, err := net.Dial("tcp", "127.0.0.1:7505")
	if err != nil {
		// handle error
		return nil, err
	}

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString(byte('\n'))

	if err != nil {
		return nil, err
	}

	log.Printf("Openvpn says: %s", response)

	return &Management{conn: conn, reader: reader, myStatus: NewStatus()}, nil
}

func (m *Management) readLoadStats() error {

	m.conn.Write([]byte("load-stats\n"))
	response, err := m.reader.ReadString(byte('\n'))

	if err != nil {
		return err
	}

	m.myStatus.LoadStats.Update(
		strings.Trim(response, "\r\n"),
	)

	return nil
}

func (m *Management) readStatus() error {
	m.conn.Write([]byte("status\n"))

	var (
		status []string
		line   string
		err    error
	)

	status = make([]string, 0, 20)

	for {
		line, err = m.reader.ReadString(byte('\n'))

		if err != nil {
			return err
		}

		status = append(
			status, strings.Trim(line, "\r\n"),
		)

		// Figure out if this is the last line of status report
		if line == "END\r\n" {
			break
		}
	}

	m.myStatus.Update(status)

	return nil
}

func (m *Management) UpdateLoop(updates chan Status) {
	var err error

	for {

		err = m.readLoadStats()
		if err != nil {
			panic(err)
		}

		err = m.readStatus()
		if err != nil {
			panic(err)
		}

		updates <- *m.myStatus
		time.Sleep(time.Second)
	}
}
