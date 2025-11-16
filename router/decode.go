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
	"strings"
)

type decodedRequest struct {
	ContentType        string
	ImpliedContentType string
	Body               []byte
}

func decodeRequest(r *http.Request) (*decodedRequest, error) {
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

	ret := &decodedRequest{
		Body:        raw, // this is the default if decoding fails
		ContentType: contentType,
	}

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
	s := string(raw)
	if len(s) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	// Reject if it looks nothing like a URL-encoded form.
	if !strings.Contains(s, "=") && !strings.Contains(s, "%") {
		return nil, fmt.Errorf("not application/x-www-form-urlencoded")
	}
	vals, err := url.ParseQuery(s)
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
