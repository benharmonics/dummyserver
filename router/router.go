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
	multipartFormMaxMemory int64 = 32 << 20 // 32 mb stored in memory; additional form data stored on disk

	requestID atomic.Uint64
)

func Router(w http.ResponseWriter, r *http.Request) {
	reqID := requestID.Add(1)
	log.Printf("[*] %s \"%s %s %s\" (Request %d)\n", r.RemoteAddr, r.Method, r.URL, r.Proto, reqID)
	if PrintHeaders {
		b, _ := json.MarshalIndent(r.Header, "", "\t")
		log.Printf("[*] Request %d - Headers:\n%v\n", reqID, string(b))
	}
	if r.Method == http.MethodOptions {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		// w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH")
	}
	if IgnoreParse {
		_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: "OK"})
		return
	}
	contentType, b, err := decodeBody(r)
	if err != nil {
		_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: err.Error()})
		log.Printf("[!] Request %d (Content-Type: \"%s\"):\n%s\n", reqID, contentType, err)
		return
	}
	_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: "OK"})
	log.Printf("[*] Request %d (Content-Type: \"%s\"):\n%s\n", reqID, contentType, string(b))
}

func decodeBody(r *http.Request) (string, []byte, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" || strings.HasPrefix(contentType, "application/json") {
		b, err := decodeJSON(r)
		return contentType, b, err
	} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		b, err := decodeURLEncoded(r)
		return contentType, b, err
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		b, err := decodeFormData(r)
		return contentType, b, err
	}
	return contentType, nil, fmt.Errorf("unsupported Content-Type: \"%s\"", contentType)
}

func decodeJSON(r *http.Request) ([]byte, error) {
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse as application/json: %s", err)
	}
	return json.MarshalIndent(body, "", "\t")
}

func decodeURLEncoded(r *http.Request) ([]byte, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse as application/x-www-form-urlencoded: %s", err)
	}
	return json.MarshalIndent(r.Form, "", "\t")
}

func decodeFormData(r *http.Request) ([]byte, error) {
	if err := r.ParseMultipartForm(multipartFormMaxMemory); err != nil {
		return nil, fmt.Errorf("failed to parse as multipart/form-data: %s", err)
	}
	return json.MarshalIndent(r.MultipartForm, "", "\t")
}
