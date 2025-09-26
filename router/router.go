package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type response struct {
	ID      uint64 `json:"id"`
	Message string `json:"message"`
}

var (
	PrintHeaders, IgnoreParse bool

	// https://pkg.go.dev/net/http#Request.ParseMultipartForm
	multipartFormMaxMemory int64 = 50_000_000 // 50 mb stored in memory; additional form data stored on disk

	requestID atomic.Uint64
)

func logResponse(requestID uint64, data []byte, err error) {
	if err == nil {
		log.Printf("[*] Request %d:\n%s\n", requestID, string(data))
	} else {
		log.Printf("[!] Request %d:\n%s\n", requestID, err)
	}
}

func responseOK(id uint64) response { return response{id, "OK"} }

func Router(w http.ResponseWriter, r *http.Request) {
	reqID := requestID.Add(1)
	log.Printf("[*] %s \"%s %s %s\" (Request %d)\n", r.RemoteAddr, r.Method, r.URL, r.Proto, reqID)
	if PrintHeaders {
		b, _ := json.MarshalIndent(r.Header, "", "\t")
		log.Printf("[*] Request header (request %d):\n%v\n", reqID, string(b))
	}
	switch r.Method {
	case http.MethodOptions: // CORS
		w.Header().Add("Access-Control-Allow-Origin", "*")
		// w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(responseOK(reqID))
	case http.MethodPost:
		if IgnoreParse {
			_ = json.NewEncoder(w).Encode(responseOK(reqID))
			return
		}
		next := decodeBody(reqID)
		next(w, r)
	default:
		err := fmt.Errorf("unsupported method: %s", r.Method)
		_ = json.NewEncoder(w).Encode(response{reqID, err.Error()})
		logResponse(reqID, nil, err)
	}
}

func decodeBody(requestID uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			next := decodeJSON(requestID, contentType)
			next(w, r)
			return
		} else if strings.HasPrefix(contentType, "multipart/form-data") {
			next := decodeFormData(requestID)
			next(w, r)
			return
		}
		err := fmt.Errorf("unsupported Content-Type: %s", contentType)
		_ = json.NewEncoder(w).Encode(response{requestID, err.Error()})
		logResponse(requestID, nil, err)
	}
}

func decodeJSON(requestID uint64, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			err := fmt.Errorf("failed to parse as %s", contentType)
			_ = json.NewEncoder(w).Encode(response{requestID, err.Error()})
			logResponse(requestID, nil, err)
			return
		}
		_ = json.NewEncoder(w).Encode(responseOK(requestID))
		b, _ := json.MarshalIndent(body, "", "\t")
		logResponse(requestID, b, nil)
	}
}

func decodeFormData(requestID uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(multipartFormMaxMemory); err != nil {
			err := fmt.Errorf("failed to parse as multipart/form-data")
			_ = json.NewEncoder(w).Encode(response{requestID, err.Error()})
			logResponse(requestID, nil, err)
			return
		}
		_ = json.NewEncoder(w).Encode(responseOK(requestID))
		b, _ := json.MarshalIndent(r.MultipartForm, "", "\t")
		logResponse(requestID, b, nil)
	}
}
