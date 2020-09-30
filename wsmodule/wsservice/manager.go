package wsservice

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type AcceptHandler func(Conn)
type CloseHandler func(Conn)

type Manager interface {
	HandleConnection(http.ResponseWriter, *http.Request)
	RegisterAcceptHandler(AcceptHandler)
	RegisterCloseHandler(CloseHandler)
}

type manager struct {
	managerAsync
	upgrader      websocket.Upgrader
	acceptHandler AcceptHandler
	closeHandler  CloseHandler
}

type managerAsync struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]Conn
}

func NewManager() Manager {
	m := &manager{
		upgrader: websocket.Upgrader{},
	}
	m.upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	m.conns = make(map[*websocket.Conn]Conn)
	return m
}

func (m *manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Lister Connection Error:", err)
		fmt.Fprintln(w, "Lister Connection Error:", err)
		return
	}
	defer m.closeConnection(ws)

	addr := ws.RemoteAddr().String()
	fmt.Println("NewConnetciton:", addr)

	c := NewConn(ws)
	m.mu.Lock()
	m.conns[ws] = c
	m.mu.Unlock()

	if m.acceptHandler != nil {
		m.acceptHandler(c)
	}
}

func (m *manager) RegisterAcceptHandler(handler AcceptHandler) {
	m.acceptHandler = handler
}

func (m *manager) RegisterCloseHandler(handler CloseHandler) {
	m.closeHandler = handler
}

func (m *manager) closeConnection(ws *websocket.Conn) {
	addr := ws.RemoteAddr().String()
	fmt.Println("CloseConnection:", addr)

	m.mu.Lock()
	c := m.conns[ws]
	delete(m.conns, ws)
	m.mu.Unlock()

	ws.Close()

	if m.closeHandler != nil {
		m.closeHandler(c)
	}
}
