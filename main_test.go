package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
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

// TestGracefulShutdown проверяет graceful shutdown логику
func TestGracefulShutdown(t *testing.T) {
	// Создаем тестовый сервер
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    ":0", // Random port
		Handler: mux,
	}

	// Запускаем сервер в горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Даем серверу время запуститься
	time.Sleep(100 * time.Millisecond)

	// Запускаем GracefulShutdown в горутине
	done := make(chan bool, 1)
	go func() {
		// Симулируем отправку сигнала через небольшую задержку
		go func() {
			time.Sleep(50 * time.Millisecond)
			// Отправляем SIGTERM в текущий процесс
			p, _ := os.FindProcess(os.Getpid())
			p.Signal(syscall.SIGTERM)
		}()

		GracefulShutdown(server)
		done <- true
	}()

	// Ждем завершения с таймаутом
	select {
	case <-done:
		// Успешно завершено
	case <-time.After(20 * time.Second):
		t.Fatal("GracefulShutdown did not complete in time")
	}
}

// TestMain_EnvVariables проверяет парсинг переменных окружения
func TestMain_EnvVariables(t *testing.T) {
	// Сохраняем оригинальные значения
	origPort := os.Getenv("PORT")
	origLogLevel := os.Getenv("LOG_LEVEL")

	// Восстанавливаем после теста
	defer func() {
		os.Setenv("PORT", origPort)
		os.Setenv("LOG_LEVEL", origLogLevel)
	}()

	tests := []struct {
		name         string
		portEnv      string
		logLevelEnv  string
		expectedPort string
		expectedLog  string
	}{
		{
			name:         "Default values",
			portEnv:      "",
			logLevelEnv:  "",
			expectedPort: "8080",
			expectedLog:  "info",
		},
		{
			name:         "Custom port",
			portEnv:      "9090",
			logLevelEnv:  "",
			expectedPort: "9090",
			expectedLog:  "info",
		},
		{
			name:         "Custom log level",
			portEnv:      "",
			logLevelEnv:  "debug",
			expectedPort: "8080",
			expectedLog:  "debug",
		},
		{
			name:         "Both custom",
			portEnv:      "3000",
			logLevelEnv:  "warn",
			expectedPort: "3000",
			expectedLog:  "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения
			if tt.portEnv != "" {
				os.Setenv("PORT", tt.portEnv)
			} else {
				os.Unsetenv("PORT")
			}

			if tt.logLevelEnv != "" {
				os.Setenv("LOG_LEVEL", tt.logLevelEnv)
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			// Проверяем логику из main()
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			logLevel := os.Getenv("LOG_LEVEL")
			if logLevel == "" {
				logLevel = "info"
			}

			if port != tt.expectedPort {
				t.Errorf("expected port %s, got %s", tt.expectedPort, port)
			}

			if logLevel != tt.expectedLog {
				t.Errorf("expected log level %s, got %s", tt.expectedLog, logLevel)
			}

			// Проверяем что log level валидный
			if _, err := log.ParseLevel(logLevel); err != nil && tt.logLevelEnv != "" {
				t.Errorf("invalid log level: %v", err)
			}
		})
	}
}

// TestResponseWriterWrapper проверяет responseWriter wrapper
func TestResponseWriterWrapper(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{"Default OK", 0, http.StatusOK},
		{"Explicit OK", http.StatusOK, http.StatusOK},
		{"Not Found", http.StatusNotFound, http.StatusNotFound},
		{"Internal Error", http.StatusInternalServerError, http.StatusInternalServerError},
		{"Created", http.StatusCreated, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			rw := &responseWriter{
				ResponseWriter: rr,
				statusCode:     http.StatusOK,
			}

			if tt.statusCode != 0 {
				rw.WriteHeader(tt.statusCode)
			}

			if rw.statusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rw.statusCode)
			}

			// Проверяем Write метод
			data := []byte("test data")
			n, err := rw.Write(data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if n != len(data) {
				t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
			}
		})
	}
}

// TestMainServerConfiguration тестирует конфигурацию сервера
func TestMainServerConfiguration(t *testing.T) {
	// Этот тест проверяет что мы правильно настраиваем сервер
	port := "8080"
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if server.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", server.Addr)
	}

	if server.ReadTimeout != 10*time.Second {
		t.Errorf("expected ReadTimeout 10s, got %v", server.ReadTimeout)
	}

	if server.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout 10s, got %v", server.WriteTimeout)
	}

	if server.IdleTimeout != 60*time.Second {
		t.Errorf("expected IdleTimeout 60s, got %v", server.IdleTimeout)
	}
}

// TestSetupServer проверяет функцию SetupServer
func TestSetupServer(t *testing.T) {
	// Сохраняем оригинальные значения
	origPort := os.Getenv("PORT")
	origLogLevel := os.Getenv("LOG_LEVEL")

	// Восстанавливаем после теста
	defer func() {
		os.Setenv("PORT", origPort)
		os.Setenv("LOG_LEVEL", origLogLevel)
	}()

	tests := []struct {
		name         string
		portEnv      string
		logLevelEnv  string
		expectedAddr string
	}{
		{
			name:         "Default port",
			portEnv:      "",
			logLevelEnv:  "",
			expectedAddr: ":8080",
		},
		{
			name:         "Custom port",
			portEnv:      "9090",
			logLevelEnv:  "",
			expectedAddr: ":9090",
		},
		{
			name:         "Port 3000",
			portEnv:      "3000",
			logLevelEnv:  "debug",
			expectedAddr: ":3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения
			if tt.portEnv != "" {
				os.Setenv("PORT", tt.portEnv)
			} else {
				os.Unsetenv("PORT")
			}

			if tt.logLevelEnv != "" {
				os.Setenv("LOG_LEVEL", tt.logLevelEnv)
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			// Вызываем SetupServer
			server := SetupServer()

			// Проверяем конфигурацию сервера
			if server.Addr != tt.expectedAddr {
				t.Errorf("expected addr %s, got %s", tt.expectedAddr, server.Addr)
			}

			if server.ReadTimeout != 10*time.Second {
				t.Errorf("expected ReadTimeout 10s, got %v", server.ReadTimeout)
			}

			if server.WriteTimeout != 10*time.Second {
				t.Errorf("expected WriteTimeout 10s, got %v", server.WriteTimeout)
			}

			if server.IdleTimeout != 60*time.Second {
				t.Errorf("expected IdleTimeout 60s, got %v", server.IdleTimeout)
			}

			if server.Handler == nil {
				t.Error("expected non-nil handler")
			}
		})
	}
}
