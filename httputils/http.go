package httputils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/YasarKaan/go-kit/enums"
	"github.com/YasarKaan/go-kit/exceptions"
	"github.com/YasarKaan/go-kit/fileutils"
	"github.com/YasarKaan/go-kit/loggerutils"
)

// Reusable singletons to share connection pool and avoid socket exhaustion under high load.
var (
	defaultClient  *http.Client
	insecureClient *http.Client
	clientOnce     sync.Once
)

func initClients() {
	clientOnce.Do(func() {
		// Tune standard HTTP connection pool settings.
		// MaxIdleConnsPerHost is increased to 100 (from default 2) to permit high concurrency.
		defaultTransport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		insecureTransport := defaultTransport.Clone()
		insecureTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		defaultClient = &http.Client{
			Timeout:   30 * time.Second, // reduced from 300s to prevent hang-ups
			Transport: defaultTransport,
		}

		insecureClient = &http.Client{
			Timeout:   30 * time.Second, // reduced from 300s to prevent hang-ups
			Transport: insecureTransport,
		}
	})
}

func getClient(insecure bool) *http.Client {
	initClients()
	if insecure {
		return insecureClient
	}
	return defaultClient
}

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

func (r *HttpResponse) IsSuccessful() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Map parses the JSON response body into a map[string]any.
func (r *HttpResponse) Map() (map[string]any, error) {
	if r.Body == "" {
		return nil, fmt.Errorf("response body is empty")
	}
	var m map[string]any
	err := json.Unmarshal([]byte(r.Body), &m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body to map: %w", err)
	}
	return m, nil
}

var IPHeaders = []string{
	"X-Real-Ip",
	"X-Forwarded-For",
	"Proxy-Client-IP",
	"WL-Proxy-Client-IP",
	"HTTP_X_FORWARDED_FOR",
	"HTTP_X_FORWARDED",
	"HTTP_X_CLUSTER_CLIENT_IP",
	"HTTP_CLIENT_IP",
	"HTTP_FORWARDED_FOR",
	"HTTP_FORWARDED",
	"HTTP_VIA",
	"REMOTE_ADDR",
}

// GetRequestIP extracts the real public IP from headers, falling back to RemoteAddr.
// WARNING: This function trusts headers like X-Forwarded-For. It should only be used behind
// a trusted reverse proxy (e.g. Nginx, Cloudflare, AWS ALB) that sanitizes or overwrites these headers.
func GetRequestIP(req *http.Request) string {
	for _, header := range IPHeaders {
		val := req.Header.Get(header)
		if val != "" {
			parts := strings.Split(val, ",")
			trimmed := strings.TrimSpace(parts[0])
			if trimmed != "" {
				return trimmed
			}
		}
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		return host
	}
	return req.RemoteAddr
}

// GetRequestUserAgent retrieves the User-Agent header from the request.
func GetRequestUserAgent(req *http.Request) string {
	return req.Header.Get("User-Agent")
}

// ParseRequestBody parses request bodies into map[string]any depending on Content-Type.
func ParseRequestBody(body any, contentType enums.ContentType) map[string]any {
	if body == nil {
		return make(map[string]any)
	}

	if m, ok := body.(map[string]any); ok {
		return m
	}

	contentTypeStr := contentType.Value()
	t, err := enums.ContentTypeFromString(contentTypeStr)
	if err != nil {
		return map[string]any{"rawText": fmt.Sprintf("%v", body)}
	}

	switch t {
	case enums.ContentTypeJSON:
		return parseJsonBody(body)
	case enums.ContentTypeFormUrlEncoded:
		return parseFormUrlEncodedBody(fmt.Sprintf("%v", body))
	case enums.ContentTypeTextPlain:
		return map[string]any{"rawText": fmt.Sprintf("%v", body)}
	default:
		return map[string]any{"rawText": fmt.Sprintf("%v", body)}
	}
}

func parseJsonBody(body any) map[string]any {
	result := make(map[string]any)
	switch v := body.(type) {
	case string:
		_ = json.Unmarshal([]byte(v), &result)
	case []byte:
		_ = json.Unmarshal(v, &result)
	default:
		bytes, err := json.Marshal(body)
		if err == nil {
			_ = json.Unmarshal(bytes, &result)
		}
	}
	return result
}

func parseFormUrlEncodedBody(body string) map[string]any {
	result := make(map[string]any)
	if body == "" {
		return result
	}

	values, err := url.ParseQuery(body)
	if err != nil {
		return result
	}

	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// executeRequest performs HTTP request and returns mapped response.
func executeRequest(ctx context.Context, client *http.Client, urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (*HttpResponse, error) {
	var bodyReader io.Reader
	if body != nil && method != enums.MethodGet {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	req, err := http.NewRequestWithContext(ctx, string(method), urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	// Add headers
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Add Content-Type if not present and body exists
	if body != nil && method != enums.MethodGet && req.Header.Get("Content-Type") == "" {
		finalContentType := contentType
		if finalContentType == "" {
			finalContentType = enums.ContentTypeJSON
		}
		req.Header.Set("Content-Type", finalContentType.Value())
	}

	// Add Accept if not present
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}

	resp, err := client.Do(req)
	if err != nil {
		loggerutils.Error("HTTP request to {} failed: {}", urlStr, err.Error())
		return nil, exceptions.NewCustomWebServerException(500, "HTTP request failed.", nil)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, exceptions.NewCustomWebServerException(500, fmt.Sprintf("Failed to read response body: %v", err), nil)
	}

	headersMap := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersMap[k] = v[0]
		}
	}

	customResp := &HttpResponse{
		StatusCode: resp.StatusCode,
		Headers:    headersMap,
		Body:       string(respBodyBytes),
	}

	if !customResp.IsSuccessful() {
		return customResp, exceptions.NewCustomWebServerException(
			resp.StatusCode,
			fmt.Sprintf("HTTP request failed with status: %d", resp.StatusCode),
			map[string]any{"url": urlStr, "body": customResp.Body, "headers": customResp.Headers},
		)
	}

	return customResp, nil
}

// SendRequest sends standard HTTP request using the shared pool.
func SendRequest(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (*HttpResponse, error) {
	client := getClient(false)
	return executeRequest(ctx, client, urlStr, method, headers, body, contentType)
}

// DangerousSendRequestWithoutSSL sends HTTP request skipping SSL certificate verification.
// WARNING: This is insecure and should only be used in dev/test environments.
func DangerousSendRequestWithoutSSL(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (*HttpResponse, error) {
	client := getClient(true)
	return executeRequest(ctx, client, urlStr, method, headers, body, contentType)
}

// Exponential backoff with jitter calculation.
func getExponentialBackoffWithJitter(attempt int) time.Duration {
	baseDelay := 1000 * time.Millisecond * time.Duration(1<<(attempt-1))
	jitterFactor := 0.2
	randomFraction := (rand.Float64()*2 - 1) * jitterFactor // range [-0.2, 0.2]
	jitter := time.Duration(float64(baseDelay) * randomFraction)
	return baseDelay + jitter
}

// sendWithRetries implements retry logic for HTTP execution.
// To prevent double-transaction bugs, non-idempotent methods (POST, PATCH) are not retried unless an Idempotency-Key header is supplied.
func sendWithRetries(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, action func() (*HttpResponse, error)) (*HttpResponse, error) {
	var lastErr error

	// Determine idempotency safety
	methodUpper := strings.ToUpper(string(method))
	isIdempotent := methodUpper == "GET" || methodUpper == "PUT" || methodUpper == "DELETE" || methodUpper == "HEAD" || methodUpper == "OPTIONS"

	hasIdempotencyKey := false
	if headers != nil {
		for k := range headers {
			if strings.ToLower(k) == "idempotency-key" {
				hasIdempotencyKey = true
				break
			}
		}
	}

	canRetry := isIdempotent || hasIdempotencyKey
	maxRetries := 4
	if !canRetry {
		maxRetries = 1 // Limit to a single execution for safety
		loggerutils.Info("[HTTP-CLIENT] Request is non-idempotent ({}). Automatic retries are disabled unless an Idempotency-Key header is supplied.", methodUpper)
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Respect context cancellation/timeout
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if maxRetries > 1 {
			loggerutils.Info("[HTTP-CLIENT] Attempt {}/{} for request to URL: {}", attempt, maxRetries, urlStr)
		}
		
		resp, err := action()
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// If it's a CustomWebServerException, analyze HTTP status
		if webErr, ok := err.(*exceptions.CustomWebServerException); ok {
			statusCode := webErr.ErrorCode
			// Non-retriable client errors
			if statusCode >= 400 && statusCode < 500 && statusCode != 429 {
				loggerutils.Error("[HTTP-CLIENT] Non-retriable client error. Status: {}. Failing immediately.", statusCode)
				return nil, err
			}
			// Rate limiting: wait 15 seconds
			if statusCode == 429 {
				if attempt >= maxRetries {
					break
				}
				loggerutils.Warn("[HTTP-CLIENT] Rate limited. Status: 429. Waiting 15 seconds.")
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(15 * time.Second):
				}
				continue
			}
			// Retriable server errors
			if statusCode >= 500 && statusCode < 600 {
				if attempt >= maxRetries {
					break
				}
				delay := getExponentialBackoffWithJitter(attempt)
				loggerutils.Warn("[HTTP-CLIENT] Retriable server error. Status: {}. Waiting for {}ms.", statusCode, delay.Milliseconds())
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
				continue
			}
			return nil, err
		}

		// Other errors (e.g. network IO errors) are retriable
		if attempt < maxRetries {
			delay := getExponentialBackoffWithJitter(attempt)
			loggerutils.Warn("[HTTP-CLIENT] Network IO error: {}. Waiting for {}ms.", err.Error(), delay.Milliseconds())
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	if maxRetries > 1 {
		loggerutils.Error("[HTTP-CLIENT] Request to {} failed after {} attempts.", urlStr, maxRetries)
	}
	return nil, lastErr
}

// SendRequestWithRetries performs request with retries.
func SendRequestWithRetries(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, body any) (*HttpResponse, error) {
	return sendWithRetries(ctx, urlStr, method, headers, func() (*HttpResponse, error) {
		return SendRequest(ctx, urlStr, method, headers, body, enums.ContentTypeJSON)
	})
}

// DangerousSendRequestWithoutSSLWithRetries performs insecure request with retries.
// WARNING: This is insecure and should only be used in dev/test environments.
func DangerousSendRequestWithoutSSLWithRetries(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, body any) (*HttpResponse, error) {
	return sendWithRetries(ctx, urlStr, method, headers, func() (*HttpResponse, error) {
		return DangerousSendRequestWithoutSSL(ctx, urlStr, method, headers, body, enums.ContentTypeJSON)
	})
}

// executeMultipartRequest performs multipart form upload using io.Pipe for true OOM-safe streaming.
func executeMultipartRequest(ctx context.Context, client *http.Client, urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		var err error
		defer func() {
			if err != nil {
				_ = pw.CloseWithError(err)
			} else {
				_ = pw.Close()
			}
		}()

		// Add text fields
		if formFields != nil {
			for k, v := range formFields {
				if err = writer.WriteField(k, v); err != nil {
					return
				}
			}
		}

		// Add file field
		if file != nil && fileFieldName != "" && !file.IsEmpty() {
			var part io.Writer
			part, err = writer.CreateFormFile(fileFieldName, file.GetOriginalFilename())
			if err != nil {
				return
			}
			
			// OOM-Safe Stream copying
			fileReader := file.GetReader()
			if closer, ok := fileReader.(io.Closer); ok {
				defer closer.Close()
			}
			_, err = io.Copy(part, fileReader)
			if err != nil {
				return
			}
		}

		err = writer.Close()
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, pr)
	if err != nil {
		_ = pr.Close()
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "*/*")

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, exceptions.NewCustomWebServerException(500, fmt.Sprintf("Multipart request failed: %v", err), nil)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, exceptions.NewCustomWebServerException(500, fmt.Sprintf("Failed to read response body: %v", err), nil)
	}

	headersMap := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersMap[k] = v[0]
		}
	}

	customResp := &HttpResponse{
		StatusCode: resp.StatusCode,
		Headers:    headersMap,
		Body:       string(respBodyBytes),
	}

	if !customResp.IsSuccessful() {
		return customResp, exceptions.NewCustomWebServerException(
			resp.StatusCode,
			fmt.Sprintf("HTTP multipart request failed with status: %d", resp.StatusCode),
			map[string]any{"url": urlStr, "body": customResp.Body, "headers": customResp.Headers},
		)
	}

	return customResp, nil
}

// SendMultipartRequest sends multipart data.
func SendMultipartRequest(ctx context.Context, urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	client := getClient(false)
	return executeMultipartRequest(ctx, client, urlStr, headers, formFields, fileFieldName, file)
}

// DangerousSendMultipartRequestWithoutSSL sends multipart data without SSL validation.
// WARNING: This is insecure and should only be used in dev/test environments.
func DangerousSendMultipartRequestWithoutSSL(ctx context.Context, urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	client := getClient(true)
	return executeMultipartRequest(ctx, client, urlStr, headers, formFields, fileFieldName, file)
}

// SendMultipartRequestWithRetries sends multipart data with retries.
func SendMultipartRequestWithRetries(ctx context.Context, urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	return sendWithRetries(ctx, urlStr, enums.MethodPost, headers, func() (*HttpResponse, error) {
		return SendMultipartRequest(ctx, urlStr, headers, formFields, fileFieldName, file)
	})
}

// DangerousSendMultipartRequestWithoutSSLWithRetries sends multipart data without SSL validation with retries.
// WARNING: This is insecure and should only be used in dev/test environments.
func DangerousSendMultipartRequestWithoutSSLWithRetries(ctx context.Context, urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	return sendWithRetries(ctx, urlStr, enums.MethodPost, headers, func() (*HttpResponse, error) {
		return DangerousSendMultipartRequestWithoutSSL(ctx, urlStr, headers, formFields, fileFieldName, file)
	})
}

// SendRequestForMap sends standard HTTP request and returns parsed JSON response as a map.
func SendRequestForMap(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (map[string]any, error) {
	resp, err := SendRequest(ctx, urlStr, method, headers, body, contentType)
	if err != nil {
		return nil, err
	}
	return resp.Map()
}

// SendRequestWithRetriesForMap sends standard HTTP request with retries and returns parsed JSON response as a map.
func SendRequestWithRetriesForMap(ctx context.Context, urlStr string, method enums.HttpMethod, headers map[string]string, body any) (map[string]any, error) {
	resp, err := SendRequestWithRetries(ctx, urlStr, method, headers, body)
	if err != nil {
		return nil, err
	}
	return resp.Map()
}
