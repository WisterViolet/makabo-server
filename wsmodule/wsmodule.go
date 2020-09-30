package wsmodule

import (
	"net/http"
	"sync"

	"github.com/WisterViolet/makabo-server/wsmodule/app"
	"github.com/WisterViolet/makabo-server/wsmodule/wsservice"
	"github.com/mitchellh/mapstructure"
)

type WsModule interface {
	Setup()
	HandleFunction(http.ResponseWriter, *http.Request)
	Write(uint16, string)
	Read() (uint16, string, error)
}

func NewWsModule() WsModule {
	wm := &wsModule{}
	wm.mg = wsservice.NewManager()
	wm.readPacketCh = make(chan app.Packet)
	wm.uconns = make(map[wsservice.Conn]app.User)
	return wm
}

type wsModule struct {
	asyncWsModule
	mg           wsservice.Manager
	readPacketCh chan app.Packet
}

type asyncWsModule struct {
	mu     sync.Mutex
	uconns map[wsservice.Conn]app.User
}

func (wm *wsModule) Setup() {
	wm.mg.RegisterAcceptHandler(wm.onAccept)
	wm.mg.RegisterCloseHandler(wm.onClose)
}

func (wm *wsModule) HandleFunction(w http.ResponseWriter, r *http.Request) {
	wm.mg.HandleConnection(w, r)
}

func (wm *wsModule) Write(id uint16, msg string) {
	packetBody := &app.PacketBody{
		Msg: msg,
	}
	for _, u := range wm.uconns {
		u.Write(id, packetBody)
	}
}

func (wm *wsModule) Read() (id uint16, msg string, err error) {
	msgPacket := &app.PacketBody{}
	packet := <-wm.readPacketCh
	if err := mapstructure.Decode(packet, msgPacket); err != nil {
		return 0xFFFF, "", err
	}
	return packet.ID, msgPacket.Msg, nil
}

func (wm *wsModule) onAccept(c wsservice.Conn) {
	u := app.NewUser(c)

	wm.mu.Lock()
	wm.uconns[c] = u
	wm.mu.Unlock()

	u.Run(wm.readPacketCh)
}

func (wm *wsModule) onClose(c wsservice.Conn) {
	wm.mu.Lock()
	delete(wm.uconns, c)
	wm.mu.Unlock()
}
