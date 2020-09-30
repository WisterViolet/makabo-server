package app

type Packet struct {
	ID   uint16      `json:"id"`
	Body interface{} `json:"body"`
}

type PacketBody struct {
	Msg string `json:"msg"`
}
