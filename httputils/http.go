package httputils

import (
	"bytes"
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
	"time"

	"github.com/YasarKaan/go-kit/enums"
	"github.com/YasarKaan/go-kit/exceptions"
	"github.com/YasarKaan/go-kit/fileutils"
)

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

func (r *HttpResponse) IsSuccessful() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
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

// Helper to build HTTP client.
func buildClient(insecure bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}
	return &http.Client{
		Timeout:   300 * time.Second, // 300 seconds read timeout
		Transport: transport,
	}
}

// executeRequest performs HTTP request and returns mapped response.
func executeRequest(client *http.Client, urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (*HttpResponse, error) {
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

	req, err := http.NewRequest(string(method), urlStr, bodyReader)
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
		return nil, exceptions.NewCustomWebServerException(500, fmt.Sprintf("HTTP request failed: %v", err), nil)
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

// SendRequest sends standard HTTP request.
func SendRequest(urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (*HttpResponse, error) {
	client := buildClient(false)
	return executeRequest(client, urlStr, method, headers, body, contentType)
}

// SendRequestWithoutSSL sends HTTP request skipping SSL certificate verification.
func SendRequestWithoutSSL(urlStr string, method enums.HttpMethod, headers map[string]string, body any, contentType enums.ContentType) (*HttpResponse, error) {
	client := buildClient(true)
	return executeRequest(client, urlStr, method, headers, body, contentType)
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
func sendWithRetries(urlStr string, action func() (*HttpResponse, error)) (*HttpResponse, error) {
	var lastErr error
	maxRetries := 4

	for attempt := 1; attempt <= maxRetries; attempt++ {
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
				return nil, err
			}
			// Rate limiting: wait 15 seconds
			if statusCode == 429 {
				time.Sleep(15 * time.Second)
				continue
			}
			// Retriable server errors
			if statusCode >= 500 && statusCode < 600 {
				if attempt >= maxRetries {
					break
				}
				delay := getExponentialBackoffWithJitter(attempt)
				time.Sleep(delay)
				continue
			}
			return nil, err
		}

		// Other errors (e.g. network IO errors) are retriable
		if attempt < maxRetries {
			delay := getExponentialBackoffWithJitter(attempt)
			time.Sleep(delay)
		}
	}

	return nil, lastErr
}

// SendRequestWithRetries performs request with retries.
func SendRequestWithRetries(urlStr string, method enums.HttpMethod, headers map[string]string, body any) (*HttpResponse, error) {
	return sendWithRetries(urlStr, func() (*HttpResponse, error) {
		return SendRequest(urlStr, method, headers, body, enums.ContentTypeJSON)
	})
}

// SendRequestWithoutSSLWithRetries performs insecure request with retries.
func SendRequestWithoutSSLWithRetries(urlStr string, method enums.HttpMethod, headers map[string]string, body any) (*HttpResponse, error) {
	return sendWithRetries(urlStr, func() (*HttpResponse, error) {
		return SendRequestWithoutSSL(urlStr, method, headers, body, enums.ContentTypeJSON)
	})
}

// executeMultipartRequest performs multipart form upload.
func executeMultipartRequest(client *http.Client, urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add text fields
	if formFields != nil {
		for k, v := range formFields {
			_ = writer.WriteField(k, v)
		}
	}

	// Add file field
	if file != nil && fileFieldName != "" && !file.IsEmpty() {
		part, err := writer.CreateFormFile(fileFieldName, file.GetOriginalFilename())
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file.GetReader())
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
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
func SendMultipartRequest(urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	client := buildClient(false)
	return executeMultipartRequest(client, urlStr, headers, formFields, fileFieldName, file)
}

// SendMultipartRequestWithoutSSL sends multipart data without SSL validation.
func SendMultipartRequestWithoutSSL(urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	client := buildClient(true)
	return executeMultipartRequest(client, urlStr, headers, formFields, fileFieldName, file)
}

// SendMultipartRequestWithRetries sends multipart data with retries.
func SendMultipartRequestWithRetries(urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	return sendWithRetries(urlStr, func() (*HttpResponse, error) {
		return SendMultipartRequest(urlStr, headers, formFields, fileFieldName, file)
	})
}

// SendMultipartRequestWithoutSSLWithRetries sends multipart data without SSL validation with retries.
func SendMultipartRequestWithoutSSLWithRetries(urlStr string, headers map[string]string, formFields map[string]string, fileFieldName string, file *fileutils.MultipartFile) (*HttpResponse, error) {
	return sendWithRetries(urlStr, func() (*HttpResponse, error) {
		return SendMultipartRequestWithoutSSL(urlStr, headers, formFields, fileFieldName, file)
	})
}
