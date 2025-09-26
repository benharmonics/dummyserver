package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type (
	Server struct{ *http.ServeMux }

	requestIDGenerator struct {
		current uint64
		mu      sync.Mutex
	}
)

const multipartFormMaxMemory = 1 << 21 // 2mb

var idGenerator = requestIDGenerator{}

func NewServer() Server {
	s := Server{http.NewServeMux()}
	s.HandleFunc("/", handler)
	return s
}

func handler(w http.ResponseWriter, r *http.Request) {
	headerBytes, err := json.MarshalIndent(r.Header, "", "\t")
	must(err)
	requestId := idGenerator.next()
	log.Printf("[*] %s \"%s %s %s\" (request %d)\n", r.RemoteAddr, r.Method, r.URL, r.Proto, requestId)
	log.Printf("[*] Request header (request %d):\n%v\n", requestId, string(headerBytes))
	var data interface{}
	if err = json.NewDecoder(r.Body).Decode(&data); err == nil {
		b, err := json.MarshalIndent(data, "", "\t")
		must(err)
		log.Printf("[*] JSON request data (request %d):\n%s\n", requestId, string(b))
		return
	}
	// Failed to parse request as application/json - attempting to parse as multipart/form-data
	if err = r.ParseMultipartForm(multipartFormMaxMemory); err == nil {
		b, err := json.MarshalIndent(r.MultipartForm, "", "\t")
		must(err)
		log.Printf("[*] Multipart form data (request %d):\n%s\n", requestId, string(b))
		return
	}
	log.Printf("[x] Failed to parse request: unsupported Content-Type %s (request %d)\n", r.Header.Get("Content-Type"), requestId)
	http.Error(w, "failed to parse as either JSON or multipart form", http.StatusNotImplemented)
}

func (gen *requestIDGenerator) next() uint64 {
	gen.mu.Lock()
	defer gen.mu.Unlock()
	ret := gen.current
	gen.current++
	return ret
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
