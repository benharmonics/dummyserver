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

	response struct {
		ID      uint64 `json:"id"`
		Message string `json:"message"`
	}
)

var (
	idGenerator               = requestIDGenerator{}
	printHeaders, ignoreParse bool
)

func NewServer(headersOn, noparse bool) Server {
	printHeaders, ignoreParse = headersOn, noparse
	s := Server{http.NewServeMux()}
	s.HandleFunc("/", router)
	return s
}

func router(w http.ResponseWriter, r *http.Request) {
	requestId := idGenerator.next()
	log.Printf("[*] %s \"%s %s %s\" (request %d)\n", r.RemoteAddr, r.Method, r.URL, r.Proto, requestId)
	if printHeaders {
		headerBytes, err := json.MarshalIndent(r.Header, "", "\t")
		must(err) // can't error?
		log.Printf("[*] Request header (request %d):\n%v\n", requestId, string(headerBytes))
	}
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Access-Control-Allow-Origin", "*")
		// w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		return
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(response{requestId, "Ok"})
		return
	case http.MethodPost:
		if ignoreParse {
			return
		}
		next := decodeBody(requestId)
		next(w, r)
	default:
		log.Printf("[!] Unsupported HTTP method %s\n", r.Method)
	}
}

func decodeBody(requestId uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err == nil {
			b, err := json.MarshalIndent(data, "", "\t")
			must(err)
			log.Printf("[*] JSON request data (request %d):\n%s\n", requestId, string(b))
			_ = json.NewEncoder(w).Encode(response{requestId, "Ok"})
			return
		}
		log.Println("[!] Failed to parse request as application/json - attempting to parse as multipart/form-data")
		if err := r.ParseMultipartForm(1_000_000); err == nil {
			b, err := json.MarshalIndent(r.MultipartForm, "", "\t")
			must(err)
			log.Printf("[*] Multipart form data (request %d):\n%s\n", requestId, string(b))
			_ = json.NewEncoder(w).Encode(response{requestId, "Ok"})
			return
		}
		log.Printf("[x] Failed to parse request as either application/json or multipart/form-data (request %d)\n", requestId)
		http.Error(w, "failed to parse request", http.StatusBadRequest)
	}
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
