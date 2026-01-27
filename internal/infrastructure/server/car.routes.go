package server

func (s *Server) registerCarRoutes() {
	s.router.HandleFunc("/api/v1/cars", s.carsHandler)
	s.router.HandleFunc("/api/v1/cars/", s.carByIDHandler)
}
