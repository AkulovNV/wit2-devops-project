package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	// Version будет переопределена при сборке
	Version = "1.0.0"

	// Prometheus metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application info",
		},
		[]string{"version", "go_version"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(appInfo)

	// Set app info
	appInfo.WithLabelValues(Version, "1.22").Set(1)

	// Setup logging
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

// HealthResponse структура для health endpoint
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// VersionResponse структура для version endpoint
type VersionResponse struct {
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	BuildTime string `json:"build_time,omitempty"`
}

// ErrorResponse структура для ошибок
type ErrorResponse struct {
	Error     string `json:"error"`
	Status    int    `json:"status"`
	Timestamp string `json:"timestamp"`
}

// MetricsMiddleware обертка для метрик
func MetricsMiddleware(handler http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		endpoint := r.URL.Path

		// Создаем обертку для перехвата статус кода
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		handler(wrapped, r)

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", wrapped.statusCode)

		// Записываем метрики
		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		// JSON логирование
		log.WithFields(log.Fields{
			"method":   method,
			"endpoint": endpoint,
			"status":   wrapped.statusCode,
			"duration": duration,
		}).Info("HTTP request processed")
	}
}

// responseWriter wrapper для перехвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.ResponseWriter.Write(data)
}

// HealthHandler обработчик /health endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   Version,
	}

	json.NewEncoder(w).Encode(response)
}

// VersionHandler обработчик /api/version endpoint
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := VersionResponse{
		Version:   Version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// NotFoundHandler для несуществующих endpoint'ов
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	response := ErrorResponse{
		Error:     "endpoint not found",
		Status:    http.StatusNotFound,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// GracefulShutdown обработка сигналов для graceful shutdown
func GracefulShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.WithFields(log.Fields{
		"signal": sig.String(),
	}).Info("Shutdown signal received")

	// Даем 15 секунд на graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	done := make(chan struct{})
	go func() {
		if err := server.Shutdown(ctx); err != nil {
			log.WithError(err).Error("Shutdown error")
		}
		close(done)
	}()

	// Ждем завершения или timeout
	select {
	case <-done:
		log.Info("Server gracefully stopped")
	case <-time.After(15 * time.Second):
		log.Warn("Shutdown timeout exceeded")
	}
}

// SetupServer настраивает и возвращает HTTP сервер
func SetupServer() *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Set log level from env
	if level, err := log.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
	}

	// Создаем новый mux вместо использования глобального
	mux := http.NewServeMux()

	// Регистрируем handlers с метриками
	mux.HandleFunc("/health", MetricsMiddleware(HealthHandler, "GET"))
	mux.HandleFunc("/api/version", MetricsMiddleware(VersionHandler, "GET"))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", NotFoundHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.WithFields(log.Fields{
		"port":    port,
		"version": Version,
	}).Info("Starting HTTP server")

	return server
}

func main() {
	server := SetupServer()

	// Graceful shutdown
	go GracefulShutdown(server)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Fatal("Server error")
	}

	log.Info("Server stopped")
}
