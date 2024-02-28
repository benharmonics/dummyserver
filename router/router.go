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
)

func responseOK(id uint64) response { return response{id, "OK"} }

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
		res := response{requestID, fmt.Sprintf("Unsupported method: %s", r.Method)}
		_ = json.NewEncoder(w).Encode(res)
		log.Printf("[!] %s (request %d)\n", res.Message, requestID)
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
		res := response{requestID, fmt.Sprintf("Unsupported Content-Type: %s", contentType)}
		_ = json.NewEncoder(w).Encode(res)
		log.Printf("[!] %s (request %d)\n", res.Message, requestID)
	}
}

func decodeJSON(requestID uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			res := response{requestID, "Failed to parse as application/json"}
			_ = json.NewEncoder(w).Encode(res)
			log.Printf("[!] %s (request %d)\n", res.Message, requestID)
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
			res := response{requestID, "Failed to parse as multipart/form-data"}
			_ = json.NewEncoder(w).Encode(res)
			log.Printf("[!] %s (request %d)\n", res.Message, requestID)
			return
		}
		b, _ := json.MarshalIndent(r.MultipartForm, "", "\t")
		log.Printf("[*] Multipart form data (request %d):\n%s\n", requestID, string(b))
		_ = json.NewEncoder(w).Encode(responseOK(requestID))
	}
}
