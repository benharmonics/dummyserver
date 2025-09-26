package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/benharmonics/dummyserver/server"
)

func main() {
	port := flag.Int("port", 9000, "The port on which the server listens for incoming HTTP requests")
	headers := flag.Bool("headers", false, "If true, prints the headers for each received request")
	noparse := flag.Bool("noparse", false, "If true, the server will not attempt to parse the request body")
	flag.Parse()
	srv := server.NewServer(*headers, *noparse)
	addr := fmt.Sprintf("localhost:%d", *port)
	log.Println("[*] Listening on", addr)
	panic(http.ListenAndServe(addr, srv))
}
