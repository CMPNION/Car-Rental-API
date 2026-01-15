package server

import (
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"
)

func GetNewServer(addr string, db *gorm.DB) *Server {
    return &Server{
        router: http.NewServeMux(),
        mutex: &sync.Mutex{},
        db: db,
        addr: addr,
    }
}


func (s *Server) Start() error {

	srv := &http.Server{
		Addr:         s.addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	println("Starting server on", s.addr)
	return srv.ListenAndServe()
}
