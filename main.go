package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/benharmonics/dummyserver/router"
)

var (
	host         = flag.String("host", "localhost", "The exposed host name")
	port         = flag.Int("port", 9090, "The port on which the server listens for incoming HTTP requests")
	printHeaders = flag.Bool("headers", false, "If set, prints the headers for each received request")
	ignoreParse  = flag.Bool("noparse", false, "If set, the server will not attempt to parse the request body")
	logToFile    = flag.Bool("log", false, "If set, log to the log file")
	logFile      = flag.String("logfile", "dummyserver.log", "Set the log file")
)

func main() {
	flag.Parse()
	router.PrintHeaders = *printHeaders
	router.IgnoreParse = *ignoreParse
	if *logToFile {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		defer func(f *os.File) {
			if err := f.Close(); err != nil {
				log.Fatalln(err)
			}
		}(f)
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	}
	http.HandleFunc("/", router.Router)
	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Println("[*] Listening on", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
