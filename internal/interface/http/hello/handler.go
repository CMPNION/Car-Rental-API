package hellohttp

import (
	"net/http"

	hellouc "github.com/CMPNION/Car-Rental-API.git/internal/usecase/hello"
)

func RegisterHandlers(mux *http.ServeMux, uc hellouc.Usecase) {
	mux.HandleFunc("/hello", Handler(uc))
}

func Handler(uc hellouc.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(uc.Hello()))
	}
}
