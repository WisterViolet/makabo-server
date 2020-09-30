package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/WisterViolet/makabo-server/wsmodule"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Response")
	})

	wm := wsmodule.NewWsModule()
	wm.Setup()
	http.HandleFunc("/ws", wm.HandleFunction)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
