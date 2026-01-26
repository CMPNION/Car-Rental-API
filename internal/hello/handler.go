package hello

import "net/http"

func RegisterHandlers(mux *http.ServeMux, uc Usecase) {
	mux.HandleFunc("/hello", Handler(uc))
}

func Handler(uc Usecase) http.HandlerFunc {
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
