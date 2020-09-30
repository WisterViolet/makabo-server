package wsservice

import (
	"fmt"
	"sync"

	"github.com/WisterViolet/makabo-server/wsmodule/app"
	"github.com/gorilla/websocket"
)

type Conn interface {
	Run(readCh chan []byte, closeCh chan bool)
	Write([]byte)
	Close()
}

func NewConn(ws *websocket.Conn) app.Conn {
	return &conn{
		ws: ws,
	}
}

type conn struct {
	ws      *websocket.Conn
	wg      *sync.WaitGroup
	writeCh chan []byte
}

func (c *conn) Run(readCh chan []byte, closeCh chan bool) {
	c.wg = &sync.WaitGroup{}
	c.writeCh = make(chan []byte, 5)

	errCh := make(chan error)
	c.wg.Add(1)
	go c.waitRead(readCh, errCh)

	c.wg.Add(1)
	go c.waitWrite()
	for {
		select {
		case <-errCh:
			close(c.writeCh)
			c.wg.Wait()

			close(closeCh)
			return
		}
	}
}

func (c *conn) Write(data []byte) {
	c.writeCh <- data
}

func (c *conn) Close() {
	c.ws.Close()
}

func (c *conn) waitWrite() {
	defer c.wg.Done()

	fmt.Println("Begin waitWrite Goroutine")
	for data := range c.writeCh {
		if err := c.ws.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("Error", err)
			break
		}
	}
	c.Close()
	fmt.Println("End waitWrite Gorouutine")
}

func (c *conn) waitRead(readCh chan []byte, errCh chan error) {
	defer c.wg.Done()

	fmt.Println("Begin waitRead Goroutine")
	for {
		_, readData, err := c.ws.ReadMessage()
		if err != nil {
			errCh <- err
			break
		}
		readCh <- readData
	}
	c.Close()
	fmt.Println("End waitRead Goroutine")
}
