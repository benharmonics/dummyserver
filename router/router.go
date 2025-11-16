package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
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
	req, err := decodeBody(r)
	if err != nil {
		_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: err.Error()})
		log.Printf("[!] Request %d (Content-Type: \"%s\"):\n%s\n", reqID, req.ContentType, err)
		return
	}
	_ = json.NewEncoder(w).Encode(response{ID: reqID, Message: "OK"})
	log.Printf("[*] Request %d (Content-Type: \"%s\", Implied Content-Type: \"%s\"):\n%s\n", reqID, req.ContentType, req.ImpliedContentType, string(req.Body))
}

type decoded struct {
	ContentType        string
	ImpliedContentType string
	Body               []byte
}

func decodeBody(r *http.Request) (*decoded, error) {
	contentType := r.Header.Get("Content-Type")

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %s", err)
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println("[!] Failed to close request body:", err)
		}
	}()

	ret := &decoded{Body: raw, ContentType: contentType}

	// Try JSON
	if pretty, err := decodeJSON(raw); err == nil {
		ret.Body = pretty
		ret.ImpliedContentType = "application/json"
		return ret, nil
	}

	// Try application/x-www-form-urlencoded
	if pretty, err := decodeURLEncoded(raw); err == nil {
		ret.Body = pretty
		ret.ImpliedContentType = "application/x-www-form-urlencoded"
		return ret, nil
	}

	// Try multipart/form-data
	if pretty, err := decodeFormData(raw, contentType); err == nil {
		ret.Body = pretty
		ret.ImpliedContentType = "multipart/form-data"
		return ret, nil
	}

	// Fallback; just treat as plain text
	ret.ImpliedContentType = "text/plain"
	return ret, nil
}

func decodeJSON(raw []byte) ([]byte, error) {
	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("failed to parse as application/json: %w", err)
	}
	return json.MarshalIndent(body, "", "\t")
}

func decodeURLEncoded(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	vals, err := url.ParseQuery(string(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to parse as application/x-www-form-urlencoded: %w", err)
	}
	if len(vals) == 0 {
		return nil, fmt.Errorf("no form values found")
	}
	return json.MarshalIndent(vals, "", "\t")
}

func decodeFormData(raw []byte, contentType string) ([]byte, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Content-Type header: %w", err)
	}
	if mediaType != "multipart/form-data" {
		return nil, fmt.Errorf("media type %s is not multipart/form-data", mediaType)
	}

	boundary, ok := params["boundary"]
	if !ok || boundary == "" {
		return nil, fmt.Errorf("multipart/form-data missing boundary")
	}

	mr := multipart.NewReader(bytes.NewReader(raw), boundary)

	type fileInfo struct {
		Filename string               `json:"filename"`
		Header   textproto.MIMEHeader `json:"header"`
		Size     int64                `json:"size"`
	}

	type multipartJSON struct {
		Value map[string][]string   `json:"value"`
		File  map[string][]fileInfo `json:"file"`
	}

	out := multipartJSON{
		Value: make(map[string][]string),
		File:  make(map[string][]fileInfo),
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate multipart parts: %w", err)
		}

		name := part.FormName()
		if name == "" {
			continue
		}

		data, err := io.ReadAll(part)
		if err != nil {
			return nil, fmt.Errorf("failed to read multipart part %s: %w", name, err)
		}

		if filename := part.FileName(); filename == "" {
			// regular field
			out.Value[name] = append(out.Value[name], string(data))
		} else {
			// file field
			fi := fileInfo{
				Filename: filename,
				Header:   part.Header,
				Size:     int64(len(data)),
			}
			out.File[name] = append(out.File[name], fi)
		}
	}

	// If nothing parsed, consider it a failure so caller can fallback
	if len(out.Value) == 0 && len(out.File) == 0 {
		return nil, fmt.Errorf("no multipart form data found")
	}

	return json.MarshalIndent(out, "", "\t")
}
