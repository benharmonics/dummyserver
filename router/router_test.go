package router_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benharmonics/dummyserver/router"
)

func TestRouter(t *testing.T) {
	type args struct {
		method         string
		route          string
		contentType    string
		wantStatusCode int
		body           io.Reader
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "accepts JSON request",
			args: args{
				method:         http.MethodPost,
				route:          "/",
				contentType:    "application/json",
				wantStatusCode: http.StatusOK,
				body:           bytes.NewReader([]byte(`{ "key": "value" }`)),
			},
		},
		{
			name: "accepts non-POST request",
			args: args{
				method:         http.MethodPatch,
				route:          "/",
				contentType:    "application/json",
				wantStatusCode: http.StatusOK,
				body:           bytes.NewReader([]byte(`{ "key": "value" }`)),
			},
		},
		{
			name: "empty JSON request is ok",
			args: args{
				method:         http.MethodPost,
				route:          "/",
				contentType:    "application/json",
				wantStatusCode: http.StatusOK,
				body:           nil,
			},
		},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(tt.args.method, tt.args.route, tt.args.body)
		r.Header.Set("Content-Type", tt.args.contentType)
		t.Run(tt.name, func(t *testing.T) {
			router.Router(w, r)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.args.wantStatusCode {
				t.Errorf("expected status to be %d, got %d", tt.args.wantStatusCode, res.StatusCode)
			}
		})
	}
}
