package app

import (
	"encoding/json"
	"fmt"
)

type User interface {
	Run(readPacketCh chan Packet)
	Write(msgid uint16, body interface{})
}

type user struct {
	conn   Conn
	readCh chan []byte
}

func NewUser(c Conn) User {
	return &user{
		conn:   c,
		readCh: make(chan []byte, 5),
	}
}

func (u *user) Run(readPacketCh chan Packet) {
	closeCh := make(chan bool)

	go u.conn.Run(u.readCh, closeCh)

	for {
		select {
		case data := <-u.readCh:
			u.handleData(data, readPacketCh)
		case <-closeCh:
			return
		default:
		}
	}
}

func (u *user) Write(msgid uint16, body interface{}) {
	packet := &Packet{
		ID:   msgid,
		Body: body,
	}
	bytes, err := json.Marshal(packet)
	if err != nil {
		u.conn.Close()
		return
	}
	u.conn.Write(bytes)
}

func (u *user) handleData(data []byte, readPacketCh chan Packet) {
	packet := &Packet{}
	if err := json.Unmarshal(data, packet); err != nil {
		fmt.Println("handle:", err)
		return
	}
	readPacketCh <- *packet
}
