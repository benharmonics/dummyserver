package router

import (
	"encoding/json"
	"log"
	"net/http"
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
	if r.Method == http.MethodOptions { // TODO: CORS?
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "*")
	}
	if IgnoreParse {
		_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: "OK"})
		return
	}
	decoded, err := decodeRequest(r)
	if err != nil {
		_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: err.Error()})
		log.Printf("[!] Request %d (Content-Type: \"%s\"):\n%s\n", reqID, decoded.ContentType, err)
		return
	}
	_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: "OK"})
	log.Printf("[*] Request %d (Content-Type: \"%s\", Implied Content-Type: \"%s\"):\n%s\n", reqID, decoded.ContentType, decoded.ImpliedContentType, string(decoded.Body))
}
