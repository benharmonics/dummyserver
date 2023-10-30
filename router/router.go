package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type response struct {
	ID      uint64 `json:"id"`
	Message string `json:"message"`
}

var (
	PrintHeaders, IgnoreParse bool

	// https://pkg.go.dev/net/http#Request.ParseMultipartForm
	multipartFormMaxMemory int64 = 50_000_000 // 50 mb stored in memory; additional form data stored on disk

	idGenerator = requestIDGenerator{}
	responseOK  = func(id uint64) response { return response{id, "OK"} }
)

func Router(w http.ResponseWriter, r *http.Request) {
	requestID := idGenerator.next()
	log.Printf("[*] %s \"%s %s %s\" (request %d)\n", r.RemoteAddr, r.Method, r.URL, r.Proto, requestID)
	if PrintHeaders {
		b, err := json.MarshalIndent(r.Header, "", "\t")
		must(err) // can't error?
		log.Printf("[*] Request header (request %d):\n%v\n", requestID, string(b))
	}
	switch r.Method {
	case http.MethodOptions: // CORS
		w.Header().Add("Access-Control-Allow-Origin", "*")
		// w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(responseOK(requestID))
	case http.MethodPost:
		if IgnoreParse {
			_ = json.NewEncoder(w).Encode(responseOK(requestID))
			return
		}
		next := decodeBody(requestID)
		next(w, r)
	default:
		log.Printf("[!] Unsupported HTTP method %s\n", r.Method)
		http.Error(w, fmt.Sprintf("unsupported method: %s", r.Method), http.StatusNotImplemented)
	}
}

func decodeBody(requestID uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			b, err := json.MarshalIndent(body, "", "\t")
			must(err)
			log.Printf("[*] JSON request data (request %d):\n%s\n", requestID, string(b))
			_ = json.NewEncoder(w).Encode(responseOK(requestID))
			return
		}
		log.Println("[!] Failed to parse request as application/json - attempting to parse as multipart/form-data")
		if err := r.ParseMultipartForm(multipartFormMaxMemory); err == nil {
			b, err := json.MarshalIndent(r.MultipartForm, "", "\t")
			must(err)
			log.Printf("[*] Multipart form data (request %d):\n%s\n", requestID, string(b))
			_ = json.NewEncoder(w).Encode(responseOK(requestID))
			return
		}
		log.Printf("[x] Failed to parse request as either application/json or multipart/form-data (request %d)\n", requestID)
		http.Error(w, "failed to parse request", http.StatusBadRequest)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
