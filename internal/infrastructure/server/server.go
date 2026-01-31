package server

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/CMPNION/Car-Rental-API.git/internal/auth/controllers"
	"github.com/CMPNION/Car-Rental-API.git/internal/auth/middleware"
	"github.com/CMPNION/Car-Rental-API.git/internal/auth/services"
	"github.com/CMPNION/Car-Rental-API.git/internal/hello"
)

func GetNewServer(addr string, db *gorm.DB) *Server {
	srv := &Server{
		router: http.NewServeMux(),
		mutex:  &sync.Mutex{},
		db:     db,
		addr:   addr,
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

	jwtMiddleware := middleware.JWTAuthMiddleware(jwtSecret)
	helloService := hello.NewService()
	s.router.Handle("/hello", jwtMiddleware(http.HandlerFunc(hello.Handler(helloService))))

	authService := services.NewAuthService(s.db, jwtSecret)
	controllers.RegisterHandlers(s.router, authService, jwtSecret)

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
