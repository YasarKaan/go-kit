package httputils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YasarKaan/go-kit/enums"
)

func TestHttpResponseMap(t *testing.T) {
	resp := &HttpResponse{
		StatusCode: 200,
		Body:       `{"status": "ok", "count": 10}`,
	}

	m, err := resp.Map()
	if err != nil {
		t.Fatalf("unexpected error parsing map: %v", err)
	}

	if m["status"] != "ok" || m["count"] != float64(10) {
		t.Errorf("unexpected map values: %+v", m)
	}
}

func TestSendRequestForMap(t *testing.T) {
	// Start mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`{"success": true, "message": "hello"}`))
	}))
	defer server.Close()

	m, err := SendRequestForMap(server.URL, enums.MethodGet, nil, nil, enums.ContentTypeJSON)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if m["success"] != true || m["message"] != "hello" {
		t.Errorf("unexpected map response values: %+v", m)
	}
}
