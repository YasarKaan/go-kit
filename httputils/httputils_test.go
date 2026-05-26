package httputils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	ctx := context.Background()
	m, err := SendRequestForMap(ctx, server.URL, enums.MethodGet, nil, nil, enums.ContentTypeJSON)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if m["success"] != true || m["message"] != "hello" {
		t.Errorf("unexpected map response values: %+v", m)
	}
}

func TestSendRequestWithCancel(t *testing.T) {
	// Start mock HTTP server that hangs
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(500 * time.Millisecond)
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50 * time.Millisecond)
	defer cancel()

	_, err := SendRequest(ctx, server.URL, enums.MethodGet, nil, nil, enums.ContentTypeJSON)
	if err == nil {
		t.Fatal("expected request to fail due to context timeout/cancel, but it succeeded")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "canceled") {
		t.Logf("request failed as expected: %v", err)
	}
}

func TestIdempotencyRetries(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		callCount++
		rw.WriteHeader(http.StatusInternalServerError) // Always fail to trigger retries
	}))
	defer server.Close()

	ctx := context.Background()

	// 1. GET request (idempotent) -> should retry 4 times
	callCount = 0
	_, err := SendRequestWithRetries(ctx, server.URL, enums.MethodGet, nil, nil)
	if err == nil {
		t.Fatal("expected error from failing server")
	}
	if callCount != 4 {
		t.Errorf("expected GET to retry 4 times, got: %d", callCount)
	}

	// 2. POST request (non-idempotent) -> should retry only 1 time (no idempotency key)
	callCount = 0
	_, err = SendRequestWithRetries(ctx, server.URL, enums.MethodPost, nil, nil)
	if err == nil {
		t.Fatal("expected error from failing server")
	}
	if callCount != 1 {
		t.Errorf("expected POST without idempotency key to run exactly 1 time, got: %d", callCount)
	}

	// 3. POST request with Idempotency-Key -> should retry 4 times
	callCount = 0
	headers := map[string]string{"Idempotency-Key": "unique-uuid"}
	_, err = SendRequestWithRetries(ctx, server.URL, enums.MethodPost, headers, nil)
	if err == nil {
		t.Fatal("expected error from failing server")
	}
	if callCount != 4 {
		t.Errorf("expected POST with idempotency key to retry 4 times, got: %d", callCount)
	}
}

func TestClientOptionsPattern(t *testing.T) {
	client := NewClient(
		WithTimeout(100 * time.Millisecond),
		WithMaxRetries(2),
		WithInsecureTLS(),
	)

	if client.maxRetries != 2 {
		t.Errorf("expected maxRetries 2, got %d", client.maxRetries)
	}

	if client.httpClient.Timeout != 100 * time.Millisecond {
		t.Errorf("expected timeout 100ms, got %v", client.httpClient.Timeout)
	}
}
