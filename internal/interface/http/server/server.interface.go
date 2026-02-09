package server

import (
	"net/http"
	"sync"

	"gorm.io/gorm"
)

type Server struct {
	mutex     *sync.Mutex
	router    *http.ServeMux
	db        *gorm.DB
	addr      string
	jwtSecret string
}
