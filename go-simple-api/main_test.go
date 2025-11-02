package gosimpleapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthHandler проверяет /health endpoint
func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем Content-Type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Проверяем body
	var response HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", response.Status)
	}

	if response.Version != Version {
		t.Errorf("expected version '%s', got '%s'", Version, response.Version)
	}

	if response.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

// TestVersionHandler проверяет /api/version endpoint
func TestVersionHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(VersionHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем body
	var response VersionResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if response.Version != Version {
		t.Errorf("expected version '%s', got '%s'", Version, response.Version)
	}

	if response.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

// TestNotFoundHandler проверяет обработку неизвестных endpoint'ов
func TestNotFoundHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NotFoundHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	// Проверяем body
	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if response.Status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, response.Status)
	}

	if response.Error != "endpoint not found" {
		t.Errorf("expected error 'endpoint not found', got '%s'", response.Error)
	}
}

// TestHealthHandlerContentType проверяет Content-Type
func TestHealthHandlerContentType(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)
	handler.ServeHTTP(rr, req)

	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("expected Content-Type '%s', got '%s'", expected, ct)
	}
}

// TestVersionHandlerReturnsValidJSON проверяет валидность JSON
func TestVersionHandlerReturnsValidJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(VersionHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем что это валидный JSON
	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	// Проверяем наличие необходимых полей
	requiredFields := []string{"version", "timestamp"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("expected field '%s' not found in response", field)
		}
	}
}

// TestMetricsMiddleware проверяет что middleware работает
func TestMetricsMiddleware(t *testing.T) {
	// Создаем simple handler
	simpleHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}

	// Оборачиваем в middleware
	wrappedHandler := MetricsMiddleware(simpleHandler, "GET")

	// Делаем request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Проверяем что response прошел через middleware
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}

	body := rr.Body.String()
	if body != `{"status":"ok"}` {
		t.Errorf("expected body '{\"status\":\"ok\"}', got '%s'", body)
	}
}

// TestHealthHandlerResponseFields проверяет наличие всех полей
func TestHealthHandlerResponseFields(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)
	handler.ServeHTTP(rr, req)

	var response HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	// Проверяем что все поля заполнены
	if response.Status == "" {
		t.Error("expected non-empty Status")
	}
	if response.Timestamp == "" {
		t.Error("expected non-empty Timestamp")
	}
	if response.Version == "" {
		t.Error("expected non-empty Version")
	}
}
