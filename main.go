package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/benharmonics/dummyserver/router"
)

var (
	port         = flag.Int("port", 9090, "The port on which the server listens for incoming HTTP requests")
	printHeaders = flag.Bool("headers", false, "If set, prints the headers for each received request")
	ignoreParse  = flag.Bool("noparse", false, "If set, the server will not attempt to parse the request body")
)

func main() {
	flag.Parse()
	router.PrintHeaders = *printHeaders
	router.IgnoreParse = *ignoreParse
	s := http.NewServeMux()
	s.HandleFunc("/", router.Router)
	addr := fmt.Sprintf("localhost:%d", *port)
	log.Println("[*] Listening on", addr)
	panic(http.ListenAndServe(addr, s))
}
