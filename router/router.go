package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
		b, _ := json.MarshalIndent(r.Header, "", "\t")
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
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			next := decodeJSON(requestID)
			next(w, r)
			return
		}
		if strings.HasPrefix(contentType, "multipart/form-data") {
			next := decodeFormData(requestID)
			next(w, r)
			return
		}
		log.Printf("[x] Unsupported Content-Type: %s (request %d)", contentType, requestID)
		http.Error(w, "failed to parse request", http.StatusBadRequest)
	}
}

func decodeJSON(requestID uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "failed to parse as application/json", http.StatusBadRequest)
			return
		}
		b, _ := json.MarshalIndent(body, "", "\t")
		log.Printf("[*] JSON request data (request %d):\n%s\n", requestID, string(b))
		_ = json.NewEncoder(w).Encode(responseOK(requestID))
	}
}

func decodeFormData(requestID uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(multipartFormMaxMemory); err != nil {
			http.Error(w, "failed to parse as multipart/form-data", http.StatusBadRequest)
			return
		}
		b, _ := json.MarshalIndent(r.MultipartForm, "", "\t")
		log.Printf("[*] Multipart form data (request %d):\n%s\n", requestID, string(b))
		_ = json.NewEncoder(w).Encode(responseOK(requestID))
	}
}
