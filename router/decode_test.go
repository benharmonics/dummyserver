package router

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newMultipartRequest() (buf bytes.Buffer, contentType string) {
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("username", "alice")
	_ = w.WriteField("password", "secret")
	contentType = w.FormDataContentType()
	_ = w.Close()

	return buf, contentType
}

func Test_decodeRequest(t *testing.T) {
	simpleMultipartBuf, simpleMultipartContentType := newMultipartRequest()
	type args struct {
		body        io.Reader
		contentType string
	}
	tests := []struct {
		name                   string
		args                   args
		wantImpliedContentType string
		wantErr                bool
	}{
		{
			name: "decode JSON",
			args: args{
				body:        bytes.NewReader([]byte(`{ "key": "value" }`)),
				contentType: "application/json",
			},
			wantImpliedContentType: "application/json",
			wantErr:                false,
		},
		{
			name: "decode x-www-form-urlencoded",
			args: args{
				body:        strings.NewReader(url.Values{"key": []string{"value"}}.Encode()),
				contentType: "application/x-www-form-urlencoded",
			},
			wantImpliedContentType: "application/x-www-form-urlencoded",
			wantErr:                false,
		},
		{
			name: "decode simple multipart request",
			args: args{
				body:        &simpleMultipartBuf,
				contentType: simpleMultipartContentType,
			},
			wantImpliedContentType: "multipart/form-data",
			wantErr:                false,
		},
		{
			name: "decode text/plain",
			args: args{
				body:        bytes.NewBufferString("some plain text"),
				contentType: "text/plain",
			},
			wantImpliedContentType: "text/plain",
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", tt.args.body)
			r.Header.Set("Content-Type", tt.args.contentType)
			got, err := decodeRequest(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ImpliedContentType != tt.wantImpliedContentType {
				fmt.Printf("%+v", string(got.Body))
				t.Errorf("ImpliedContentType = %v, want %v", got.ImpliedContentType, tt.wantImpliedContentType)
			}
		})
	}
}
