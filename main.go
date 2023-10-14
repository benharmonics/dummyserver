package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/benharmonics/dummyserver/server"
)

var port = flag.Int("port", 9000, "The port on which the server listens for incoming HTTP requests")

func main() {
	flag.Parse()
	srv := server.NewServer()
	addr := fmt.Sprintf("localhost:%d", *port)
	log.Println("[*] Listening on", addr)
	panic(http.ListenAndServe(addr, srv))
}
