package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	authhttp "github.com/CMPNION/Car-Rental-API.git/internal/interface/http/auth"
	hellohttp "github.com/CMPNION/Car-Rental-API.git/internal/interface/http/hello"
	authuc "github.com/CMPNION/Car-Rental-API.git/internal/usecase/auth"
	hellouc "github.com/CMPNION/Car-Rental-API.git/internal/usecase/hello"
)

// Metrics stores HTTP metrics for Prometheus
type Metrics struct {
	RequestDuration *httpMetric
	RequestsTotal   *counterMetric
	ActiveRequests  *gaugeMetric
}

type httpMetric struct {
	mu        sync.Mutex
	buckets   map[string]float64
	count     int
	sum       float64
	histogram map[string]int
}

type counterMetric struct {
	mu    sync.Mutex
	value map[string]int
}

type gaugeMetric struct {
	mu    sync.Mutex
	value int
}

func NewMetrics() *Metrics {
	return &Metrics{
		RequestDuration: &httpMetric{
			buckets:   make(map[string]float64),
			histogram: make(map[string]int),
		},
		RequestsTotal: &counterMetric{
			value: make(map[string]int),
		},
		ActiveRequests: &gaugeMetric{},
	}
}

func GetNewServer(addr string, db *gorm.DB) *Server {
	srv := &Server{
		router:  http.NewServeMux(),
		mutex:   &sync.Mutex{},
		db:      db,
		addr:    addr,
		metrics: NewMetrics(),
	}
	srv.registerCarRoutes()
	srv.registerRoutes()

	return srv
}

func (s *Server) Start() error {

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.withCORS(s.router),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	println("Starting server on", s.addr)
	return srv.ListenAndServe()
}

func (s *Server) registerRoutes() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
		log.Println("JWT_SECRET is empty, using dev-secret")
	}

	s.jwtSecret = jwtSecret

	// Health and metrics endpoints (no auth required)
	s.router.HandleFunc("/health", s.healthHandler)
	s.router.HandleFunc("/metrics", s.prometheusMetricsHandler)
	s.router.HandleFunc("/api/v1/loadtest", s.loadTestHandler)

	jwtMiddleware := authhttp.JWTAuthMiddleware(jwtSecret)
	helloService := hellouc.NewService()
	s.router.Handle("/hello", jwtMiddleware(http.HandlerFunc(hellohttp.Handler(helloService))))

	authService := authuc.NewAuthService(s.db, jwtSecret)
	authhttp.RegisterHandlers(s.router, authService, jwtSecret)

	s.router.Handle("/api/v1/rentals", jwtMiddleware(http.HandlerFunc(s.rentalsHandler)))
	s.router.Handle("/api/v1/rentals/", jwtMiddleware(http.HandlerFunc(s.rentalActionHandler)))
	s.router.Handle("/api/v1/users/balance", jwtMiddleware(http.HandlerFunc(s.userBalanceHandler)))
	s.router.Handle("/api/v1/users/me", jwtMiddleware(http.HandlerFunc(s.userProfileHandler)))
	s.router.Handle("/api/v1/transactions", jwtMiddleware(http.HandlerFunc(s.transactionsHandler)))
	s.router.Handle("/api/v1/admin/metrics", jwtMiddleware(http.HandlerFunc(s.adminMetricsHandler)))
}

func (*Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// metricsMiddleware wraps handlers to collect metrics
func (s *Server) metricsMiddleware(next http.Handler, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.metrics.ActiveRequests.mu.Lock()
		s.metrics.ActiveRequests.value++
		s.metrics.ActiveRequests.mu.Unlock()

		defer func() {
			s.metrics.ActiveRequests.mu.Lock()
			s.metrics.ActiveRequests.value--
			s.metrics.ActiveRequests.mu.Unlock()
		}()

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()

		method := r.Method
		statusKey := method + "_" + route

		s.metrics.RequestDuration.mu.Lock()
		s.metrics.RequestDuration.count++
		s.metrics.RequestDuration.sum += duration
		s.metrics.RequestDuration.histogram[statusKey]++
		s.metrics.RequestDuration.mu.Unlock()

		s.metrics.RequestsTotal.mu.Lock()
		s.metrics.RequestsTotal.value[statusKey]++
		s.metrics.RequestsTotal.mu.Unlock()
	})
}

// healthHandler returns health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		http.Error(w, `{"status":"error","database":"disconnected"}`, http.StatusServiceUnavailable)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		http.Error(w, `{"status":"error","database":"unreachable"}`, http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"database":  "connected",
	}

	json.NewEncoder(w).Encode(response)
}

// prometheusMetricsHandler exposes metrics in Prometheus format
func (s *Server) prometheusMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var sb strings.Builder

	// HTTP Request Duration Histogram
	sb.WriteString("# HELP http_request_duration_seconds HTTP request duration in seconds\n")
	sb.WriteString("# TYPE http_request_duration_seconds histogram\n")

	s.metrics.RequestDuration.mu.Lock()
	count := s.metrics.RequestDuration.count
	sum := s.metrics.RequestDuration.sum
	sb.WriteString(fmt.Sprintf("http_request_duration_seconds_count %d\n", count))
	sb.WriteString(fmt.Sprintf("http_request_duration_seconds_sum %.4f\n", sum))

	for route, routeCount := range s.metrics.RequestDuration.histogram {
		sb.WriteString(fmt.Sprintf("http_request_duration_seconds_bucket{route=\"%s\"} %d\n", route, routeCount))
	}
	s.metrics.RequestDuration.mu.Unlock()

	// HTTP Requests Total
	sb.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
	sb.WriteString("# TYPE http_requests_total counter\n")

	s.metrics.RequestsTotal.mu.Lock()
	for route, count := range s.metrics.RequestsTotal.value {
		sb.WriteString(fmt.Sprintf("http_requests_total{route=\"%s\"} %d\n", route, count))
	}
	s.metrics.RequestsTotal.mu.Unlock()

	// Active Requests
	sb.WriteString("# HELP http_requests_active Total number of active HTTP requests\n")
	sb.WriteString("# TYPE http_requests_active gauge\n")

	s.metrics.ActiveRequests.mu.Lock()
	sb.WriteString(fmt.Sprintf("http_requests_active %d\n", s.metrics.ActiveRequests.value))
	s.metrics.ActiveRequests.mu.Unlock()

	// Database connections
	sb.WriteString("# HELP db_connections_open Number of open database connections\n")
	sb.WriteString("# TYPE db_connections_open gauge\n")

	sqlDB, err := s.db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		sb.WriteString(fmt.Sprintf("db_connections_open %d\n", stats.OpenConnections))
	}

	w.Write([]byte(sb.String()))
}
