package bitbrowser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockServer creates a test server that returns the given response.
func mockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// successResponse creates a successful BitBrowser API response.
func successResponse(data any) []byte {
	resp := Response{
		Success: true,
	}
	if data != nil {
		jsonData, _ := json.Marshal(data)
		resp.Data = jsonData
	}
	b, _ := json.Marshal(resp)
	return b
}

// errorResponse creates a failed BitBrowser API response.
func errorResponse(msg string) []byte {
	resp := Response{
		Success: false,
		Msg:     msg,
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestNew(t *testing.T) {
	t.Run("creates client with default settings", func(t *testing.T) {
		client := New("http://localhost:54345")

		if client.apiURL != "http://localhost:54345" {
			t.Errorf("apiURL = %q, want %q", client.apiURL, "http://localhost:54345")
		}
		if client.httpClient == nil {
			t.Error("httpClient should not be nil")
		}
		if client.retryConfig == nil {
			t.Error("retryConfig should not be nil")
		}
		if client.retryConfig.MaxAttempts != 1 {
			t.Errorf("MaxAttempts = %d, want 1", client.retryConfig.MaxAttempts)
		}
	})

	t.Run("trims trailing slash from URL", func(t *testing.T) {
		client := New("http://localhost:54345/")

		if client.apiURL != "http://localhost:54345" {
			t.Errorf("apiURL = %q, want %q", client.apiURL, "http://localhost:54345")
		}
	})

	t.Run("applies WithHTTPClient option", func(t *testing.T) {
		customClient := &http.Client{Timeout: 5 * time.Second}
		client := New("http://localhost:54345", WithHTTPClient(customClient))

		if client.httpClient != customClient {
			t.Error("httpClient should be the custom client")
		}
	})

	t.Run("applies WithAPIKey option", func(t *testing.T) {
		client := New("http://localhost:54345", WithAPIKey("test-api-key-123"))

		if client.apiKey != "test-api-key-123" {
			t.Errorf("apiKey = %q, want %q", client.apiKey, "test-api-key-123")
		}
	})

	t.Run("applies WithLogger option", func(t *testing.T) {
		logger := slog.Default()
		client := New("http://localhost:54345", WithLogger(logger))

		if client.logger != logger {
			t.Error("logger should be set")
		}
	})

	t.Run("applies WithRetryConfig option", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts: 5,
			BaseDelay:   2 * time.Second,
		}
		client := New("http://localhost:54345", WithRetryConfig(config))

		if client.retryConfig.MaxAttempts != 5 {
			t.Errorf("MaxAttempts = %d, want 5", client.retryConfig.MaxAttempts)
		}
	})

	t.Run("applies WithRetry convenience option", func(t *testing.T) {
		client := New("http://localhost:54345", WithRetry(3))

		if client.retryConfig.MaxAttempts != 3 {
			t.Errorf("MaxAttempts = %d, want 3", client.retryConfig.MaxAttempts)
		}
	})
}

func TestHealth(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/health" {
				t.Errorf("path = %q, want /health", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.Health(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("service unavailable"))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.Health(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("network error", func(t *testing.T) {
		client := New("http://localhost:1") // Invalid port

		err := client.Health(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
		if !errors.Is(err, ErrNetwork) {
			t.Errorf("expected ErrNetwork, got %T", err)
		}
	})
}

func TestCreateProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/update" {
				t.Errorf("path = %q, want /browser/update", r.URL.Path)
			}

			var config ProfileConfig
			json.NewDecoder(r.Body).Decode(&config)

			if config.Name != "Test Profile" {
				t.Errorf("name = %q, want %q", config.Name, "Test Profile")
			}

			w.Write(successResponse(map[string]string{"id": "profile-123"}))
		})
		defer server.Close()

		client := New(server.URL)
		id, err := client.CreateProfile(context.Background(), ProfileConfig{
			Name: "Test Profile",
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != "profile-123" {
			t.Errorf("id = %q, want %q", id, "profile-123")
		}
	})

	t.Run("sets default fingerprint", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			var config ProfileConfig
			json.NewDecoder(r.Body).Decode(&config)

			if config.BrowserFingerPrint == nil {
				t.Error("BrowserFingerPrint should be set")
			}
			if config.BrowserFingerPrint.CoreVersion != DefaultCoreVersion {
				t.Errorf("CoreVersion = %q, want %q", config.BrowserFingerPrint.CoreVersion, DefaultCoreVersion)
			}

			w.Write(successResponse(map[string]string{"id": "profile-123"}))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.CreateProfile(context.Background(), ProfileConfig{
			Name: "Test Profile",
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("quota exceeded"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.CreateProfile(context.Background(), ProfileConfig{
			Name: "Test Profile",
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUpdateProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.UpdateProfile(context.Background(), ProfileConfig{
			ID:   "profile-123",
			Name: "Updated Name",
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("validation error - missing ID", func(t *testing.T) {
		client := New("http://localhost:54345")
		err := client.UpdateProfile(context.Background(), ProfileConfig{
			Name: "Updated Name",
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation, got %T", err)
		}
	})
}

func TestGetProfileDetail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(ProfileDetail{
				ID:   "profile-123",
				Name: "Test Profile",
				Seq:  1,
			}))
		})
		defer server.Close()

		client := New(server.URL)
		detail, err := client.GetProfileDetail(context.Background(), "profile-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if detail.ID != "profile-123" {
			t.Errorf("ID = %q, want %q", detail.ID, "profile-123")
		}
		if detail.Name != "Test Profile" {
			t.Errorf("Name = %q, want %q", detail.Name, "Test Profile")
		}
	})
}

func TestListProfiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(ListResult{
				List: []ProfileDetail{
					{ID: "profile-1", Name: "Profile 1"},
					{ID: "profile-2", Name: "Profile 2"},
				},
				Page:  0,
				Total: 2,
			}))
		})
		defer server.Close()

		client := New(server.URL)
		result, err := client.ListProfiles(context.Background(), ListRequest{
			Page:     0,
			PageSize: 10,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result.List) != 2 {
			t.Errorf("len(List) = %d, want 2", len(result.List))
		}
		if result.Total != 2 {
			t.Errorf("Total = %d, want 2", result.Total)
		}
	})
}

func TestDeleteProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				ID string `json:"id"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.ID != "profile-123" {
				t.Errorf("ID = %q, want %q", req.ID, "profile-123")
			}

			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.DeleteProfile(context.Background(), "profile-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDeleteProfiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				IDs []string `json:"ids"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if len(req.IDs) != 2 {
				t.Errorf("len(IDs) = %d, want 2", len(req.IDs))
			}

			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.DeleteProfiles(context.Background(), []string{"profile-1", "profile-2"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestOpen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(OpenResult{
				Ws:   "ws://127.0.0.1:9222/devtools/browser/abc",
				Http: "127.0.0.1:9222",
			}))
		})
		defer server.Close()

		client := New(server.URL)
		result, err := client.Open(context.Background(), "profile-123", nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Ws != "ws://127.0.0.1:9222/devtools/browser/abc" {
			t.Errorf("Ws = %q", result.Ws)
		}
		if !strings.HasPrefix(result.Http, "http://") {
			t.Errorf("Http should have http:// prefix, got %q", result.Http)
		}
	})

	t.Run("with options", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			var config OpenConfig
			json.NewDecoder(r.Body).Decode(&config)

			// Check that headless and LAN flags are set
			hasHeadless := false
			hasLAN := false
			for _, arg := range config.Args {
				if arg == "--headless" {
					hasHeadless = true
				}
				if arg == "--remote-debugging-address=0.0.0.0" {
					hasLAN = true
				}
			}

			if !hasHeadless {
				t.Error("expected --headless arg")
			}
			if !hasLAN {
				t.Error("expected --remote-debugging-address=0.0.0.0 arg")
			}
			if !config.IgnoreDefaultUrls {
				t.Error("expected IgnoreDefaultUrls=true for headless")
			}

			w.Write(successResponse(OpenResult{
				Ws:   "ws://127.0.0.1:9222/devtools/browser/abc",
				Http: "127.0.0.1:9222",
			}))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.Open(context.Background(), "profile-123", &OpenOptions{
			Headless: true,
			AllowLAN: true,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/close" {
				t.Errorf("path = %q, want /browser/close", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.Close(context.Background(), "profile-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGetPorts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(map[string]string{
				"profile-1": "9222",
				"profile-2": "9223",
			}))
		})
		defer server.Close()

		client := New(server.URL)
		ports, err := client.GetPorts(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if ports["profile-1"] != "9222" {
			t.Errorf("ports[profile-1] = %q, want %q", ports["profile-1"], "9222")
		}
	})
}

func TestGetBrowserVersion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/json/version" {
				t.Errorf("path = %q, want /json/version", r.URL.Path)
			}
			json.NewEncoder(w).Encode(BrowserVersion{
				Browser:              "Chrome/130.0.0.0",
				WebSocketDebuggerURL: "ws://127.0.0.1:9222/devtools/browser/abc",
			})
		})
		defer server.Close()

		client := New(server.URL)
		version, err := client.GetBrowserVersion(context.Background(), server.URL)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if version.Browser != "Chrome/130.0.0.0" {
			t.Errorf("Browser = %q", version.Browser)
		}
	})

	t.Run("validation error - empty endpoint", func(t *testing.T) {
		client := New("http://localhost:54345")
		_, err := client.GetBrowserVersion(context.Background(), "")

		if err == nil {
			t.Error("expected error, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation, got %T", err)
		}
	})
}

func TestVerifyDebugURL(t *testing.T) {
	t.Run("valid URL returns true", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/json/version" {
				t.Errorf("path = %q, want /json/version", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		})
		defer server.Close()

		client := New(server.URL)
		valid := client.VerifyDebugURL(context.Background(), server.URL)

		if !valid {
			t.Error("expected true for valid URL")
		}
	})

	t.Run("invalid URL returns false", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		defer server.Close()

		client := New(server.URL)
		valid := client.VerifyDebugURL(context.Background(), server.URL)

		if valid {
			t.Error("expected false for 404 response")
		}
	})

	t.Run("empty URL returns false", func(t *testing.T) {
		client := New("http://localhost:54345")
		valid := client.VerifyDebugURL(context.Background(), "")

		if valid {
			t.Error("expected false for empty URL")
		}
	})
}

func TestRetryBehavior(t *testing.T) {
	t.Run("retries on server error", func(t *testing.T) {
		attempts := 0
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("server error"))
				return
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL, WithRetry(3))
		err := client.Health(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attempts != 3 {
			t.Errorf("attempts = %d, want 3", attempts)
		}
	})

	t.Run("does not retry on client error", func(t *testing.T) {
		attempts := 0
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
		})
		defer server.Close()

		client := New(server.URL, WithRetry(3))
		err := client.Health(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
		if attempts != 1 {
			t.Errorf("attempts = %d, want 1 (should not retry 400)", attempts)
		}
	})
}

func TestLogging(t *testing.T) {
	t.Run("logs with logger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL, WithLogger(logger))
		_ = client.Health(context.Background())

		logs := buf.String()
		if !strings.Contains(logs, "bitbrowser") {
			t.Errorf("expected logs to contain 'bitbrowser', got: %s", logs)
		}
	})

	t.Run("works without logger", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL) // No logger
		err := client.Health(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestContextCancellation(t *testing.T) {
	t.Run("request cancelled", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Write(successResponse(nil))
		})
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		client := New(server.URL)
		err := client.Health(ctx)

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestHTTPStatusErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantAPI    bool
	}{
		{"400 Bad Request", http.StatusBadRequest, true},
		{"401 Unauthorized", http.StatusUnauthorized, true},
		{"403 Forbidden", http.StatusForbidden, true},
		{"404 Not Found", http.StatusNotFound, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("error message"))
			})
			defer server.Close()

			client := New(server.URL)
			err := client.Health(context.Background())

			if err == nil {
				t.Error("expected error, got nil")
			}
			if tt.wantAPI && !errors.Is(err, ErrAPI) {
				t.Errorf("expected ErrAPI for status %d, got %T", tt.statusCode, err)
			}
		})
	}
}

func TestSetCookies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/cookies/set" {
				t.Errorf("path = %q, want /browser/cookies/set", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.SetCookies(context.Background(), "profile-123", []Cookie{
			{Name: "session", Value: "abc123", Domain: ".example.com"},
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGetCookies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse([]Cookie{
				{Name: "session", Value: "abc123", Domain: ".example.com"},
			}))
		})
		defer server.Close()

		client := New(server.URL)
		cookies, err := client.GetCookies(context.Background(), "profile-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(cookies) != 1 {
			t.Errorf("len(cookies) = %d, want 1", len(cookies))
		}
		if cookies[0].Name != "session" {
			t.Errorf("cookies[0].Name = %q, want %q", cookies[0].Name, "session")
		}
	})
}

func TestUpdateProxy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/proxy/update" {
				t.Errorf("path = %q, want /browser/proxy/update", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.UpdateProxy(context.Background(), ProxyUpdateRequest{
			IDs:       []string{"profile-123"},
			ProxyType: "http",
			Host:      "proxy.example.com",
			Port:      8080,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClearCache(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/cache/clear" {
				t.Errorf("path = %q, want /cache/clear", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.ClearCache(context.Background(), []string{"profile-123"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRandomizeFingerprint(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(Fingerprint{
				CoreVersion: "130",
				UserAgent:   "Mozilla/5.0...",
			}))
		})
		defer server.Close()

		client := New(server.URL)
		fp, err := client.RandomizeFingerprint(context.Background(), "profile-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if fp.CoreVersion != "130" {
			t.Errorf("CoreVersion = %q, want %q", fp.CoreVersion, "130")
		}
	})
}

func TestCloseBySeqs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/close/byseqs" {
				t.Errorf("path = %q, want /browser/close/byseqs", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.CloseBySeqs(context.Background(), []int{1, 2, 3})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestCloseAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/close/all" {
				t.Errorf("path = %q, want /browser/close/all", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.CloseAll(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGetPIDs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(map[string]int{
				"profile-1": 1234,
				"profile-2": 5678,
			}))
		})
		defer server.Close()

		client := New(server.URL)
		pids, err := client.GetPIDs(context.Background(), []string{"profile-1", "profile-2"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if pids["profile-1"] != 1234 {
			t.Errorf("pids[profile-1] = %d, want 1234", pids["profile-1"])
		}
	})
}

func TestGetAllPIDs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(map[string]int{
				"profile-1": 1234,
			}))
		})
		defer server.Close()

		client := New(server.URL)
		pids, err := client.GetAllPIDs(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(pids) != 1 {
			t.Errorf("len(pids) = %d, want 1", len(pids))
		}
	})
}

func TestGetAlivePIDs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(map[string]int{
				"profile-1": 1234,
			}))
		})
		defer server.Close()

		client := New(server.URL)
		pids, err := client.GetAlivePIDs(context.Background(), []string{"profile-1"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if pids["profile-1"] != 1234 {
			t.Errorf("pids[profile-1] = %d, want 1234", pids["profile-1"])
		}
	})
}

func TestUpdateProfilePartial(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/update/partial" {
				t.Errorf("path = %q, want /browser/update/partial", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.UpdateProfilePartial(context.Background(), PartialUpdateRequest{
			IDs: []string{"profile-1", "profile-2"},
			ProfileConfig: ProfileConfig{
				Name: "Updated",
			},
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestResetClosingState(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/closing/reset" {
				t.Errorf("path = %q, want /browser/closing/reset", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.ResetClosingState(context.Background(), "profile-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestCheckProxy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(ProxyCheckResult{
				Success: true,
			}))
		})
		defer server.Close()

		client := New(server.URL)
		result, err := client.CheckProxy(context.Background(), ProxyCheckRequest{
			Host:      "proxy.example.com",
			Port:      8080,
			ProxyType: "http",
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Error("expected success=true")
		}
	})
}

func TestUpdateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/group/update" {
				t.Errorf("path = %q, want /browser/group/update", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.UpdateGroup(context.Background(), "group-1", []string{"profile-1"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestUpdateRemark(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/remark/update" {
				t.Errorf("path = %q, want /browser/remark/update", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.UpdateRemark(context.Background(), "new remark", []string{"profile-1"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestArrangeWindows(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/windowbounds" {
				t.Errorf("path = %q, want /windowbounds", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.ArrangeWindows(context.Background(), WindowBoundsRequest{
			Type:   "box",
			Width:  800,
			Height: 600,
			Col:    2,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestArrangeWindowsFlexible(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/windowbounds/flexable" {
				t.Errorf("path = %q, want /windowbounds/flexable", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.ArrangeWindowsFlexible(context.Background(), []int{1, 2, 3})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClearCacheExceptExtensions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/cache/clear/exceptExtensions" {
				t.Errorf("path = %q, want /cache/clear/exceptExtensions", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.ClearCacheExceptExtensions(context.Background(), []string{"profile-123"})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClearCookies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/browser/cookies/clear" {
				t.Errorf("path = %q, want /browser/cookies/clear", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.ClearCookies(context.Background(), "profile-123", true)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestFormatCookies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse([]Cookie{
				{Name: "session", Value: "abc123", Domain: ".example.com"},
			}))
		})
		defer server.Close()

		client := New(server.URL)
		cookies, err := client.FormatCookies(context.Background(), "session=abc123", "example.com")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(cookies) != 1 {
			t.Errorf("len(cookies) = %d, want 1", len(cookies))
		}
	})
}

func TestGetAllDisplays(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse([]Display{
				{ID: 1, Label: "Primary"},
			}))
		})
		defer server.Close()

		client := New(server.URL)
		displays, err := client.GetAllDisplays(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(displays) != 1 {
			t.Errorf("len(displays) = %d, want 1", len(displays))
		}
	})
}

func TestRunRPA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rpa/run" {
				t.Errorf("path = %q, want /rpa/run", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.RunRPA(context.Background(), "task-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestStopRPA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rpa/stop" {
				t.Errorf("path = %q, want /rpa/stop", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.StopRPA(context.Background(), "task-123")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestAutoPaste(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/autopaste" {
				t.Errorf("path = %q, want /autopaste", r.URL.Path)
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.AutoPaste(context.Background(), "profile-123", "https://example.com")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestReadExcel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/utils/readexcel" {
				t.Errorf("path = %q, want /utils/readexcel", r.URL.Path)
			}
			w.Write(successResponse([][]string{{"A1", "B1"}, {"A2", "B2"}}))
		})
		defer server.Close()

		client := New(server.URL)
		result, err := client.ReadExcel(context.Background(), "/path/to/file.xlsx")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result, got nil")
		}
	})
}

func TestReadFile(t *testing.T) {
	t.Run("success with string", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/utils/readfile" {
				t.Errorf("path = %q, want /utils/readfile", r.URL.Path)
			}
			w.Write(successResponse("file content here"))
		})
		defer server.Close()

		client := New(server.URL)
		content, err := client.ReadFile(context.Background(), "/path/to/file.txt")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content != "file content here" {
			t.Errorf("content = %q, want %q", content, "file content here")
		}
	})

	t.Run("success with non-string data", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			// Return an array instead of string to test fallback
			w.Write(successResponse([]string{"line1", "line2"}))
		})
		defer server.Close()

		client := New(server.URL)
		content, err := client.ReadFile(context.Background(), "/path/to/file.txt")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if content == "" {
			t.Error("expected content, got empty string")
		}
	})
}

func TestOpenRaw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse(OpenResult{
				Ws:   "ws://127.0.0.1:9222/devtools/browser/abc",
				Http: "127.0.0.1:9222",
			}))
		})
		defer server.Close()

		client := New(server.URL)
		result, err := client.OpenRaw(context.Background(), OpenConfig{
			ID:    "profile-123",
			Queue: true,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(result.Http, "http://") {
			t.Errorf("Http should have http:// prefix, got %q", result.Http)
		}
	})
}

func TestAPIErrorScenarios(t *testing.T) {
	t.Run("API returns success=false", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("profile not found"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetProfileDetail(context.Background(), "nonexistent")

		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "profile not found") {
			t.Errorf("error should contain 'profile not found', got %v", err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		})
		defer server.Close()

		client := New(server.URL)
		err := client.Health(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWaitForReady(t *testing.T) {
	t.Run("returns when browser ready", func(t *testing.T) {
		callCount := 0
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.URL.Path == "/browser/ports" {
				// First call returns empty, second returns port
				if callCount <= 1 {
					w.Write(successResponse(map[string]string{}))
				} else {
					w.Write(successResponse(map[string]string{"profile-123": "9222"}))
				}
			} else if r.URL.Path == "/json/version" {
				json.NewEncoder(w).Encode(BrowserVersion{
					WebSocketDebuggerURL: "ws://127.0.0.1:9222/devtools/browser/abc",
				})
			}
		})
		defer server.Close()

		client := New(server.URL)
		result, err := client.WaitForReady(context.Background(), "profile-123", 10)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected result, got nil")
		}
	})
}

func TestAPIFailureScenarios(t *testing.T) {
	t.Run("GetProfileDetail failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("profile not found"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetProfileDetail(context.Background(), "nonexistent")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetPIDs failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("internal error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetPIDs(context.Background(), []string{"profile-1"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetAllPIDs failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("internal error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetAllPIDs(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetAlivePIDs failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("internal error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetAlivePIDs(context.Background(), []string{"profile-1"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetPorts failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("internal error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetPorts(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("RandomizeFingerprint failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("internal error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.RandomizeFingerprint(context.Background(), "profile-1")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetCookies failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("browser not open"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetCookies(context.Background(), "profile-1")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("FormatCookies failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("invalid cookie format"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.FormatCookies(context.Background(), "invalid", "example.com")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("ReadExcel failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("file not found"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.ReadExcel(context.Background(), "/nonexistent.xlsx")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("ReadFile failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("file not found"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.ReadFile(context.Background(), "/nonexistent.txt")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("ListProfiles failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("database error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.ListProfiles(context.Background(), ListRequest{Page: 0, PageSize: 10})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("CheckProxy failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("proxy unreachable"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.CheckProxy(context.Background(), ProxyCheckRequest{Host: "bad", Port: 1234})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetAllDisplays failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("internal error"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.GetAllDisplays(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Open failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("profile not found"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.Open(context.Background(), "nonexistent", nil)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("OpenRaw failure", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse("profile not found"))
		})
		defer server.Close()

		client := New(server.URL)
		_, err := client.OpenRaw(context.Background(), OpenConfig{ID: "nonexistent"})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestAPIKeyAuthentication(t *testing.T) {
	t.Run("sends x-api-key header when configured", func(t *testing.T) {
		var receivedAPIKey string
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			receivedAPIKey = r.Header.Get("x-api-key")
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL, WithAPIKey("my-secret-token-123"))
		_ = client.Health(context.Background())

		if receivedAPIKey != "my-secret-token-123" {
			t.Errorf("x-api-key header = %q, want %q", receivedAPIKey, "my-secret-token-123")
		}
	})

	t.Run("does not send x-api-key header when not configured", func(t *testing.T) {
		var hasAPIKeyHeader bool
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			_, hasAPIKeyHeader = r.Header["X-Api-Key"]
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL) // No API key
		_ = client.Health(context.Background())

		if hasAPIKeyHeader {
			t.Error("x-api-key header should not be present when not configured")
		}
	})

	t.Run("authentication failure returns API error", func(t *testing.T) {
		server := mockServer(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("x-api-key")
			if apiKey != "valid-key" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"success":false,"msg":"Invalid API key"}`))
				return
			}
			w.Write(successResponse(nil))
		})
		defer server.Close()

		client := New(server.URL, WithAPIKey("invalid-key"))
		err := client.Health(context.Background())

		if err == nil {
			t.Error("expected error for invalid API key")
		}
		if !errors.Is(err, ErrAPI) {
			t.Errorf("expected ErrAPI, got %T", err)
		}
	})
}
