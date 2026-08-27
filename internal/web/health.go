package web

import (
	"net/http"
	"tactile-review/internal/application"
)

func HealthHandler(a *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		securityHeaders(w)
		writeJSON(w, map[string]string{"status": "ok", "service": "tactile-review"})
	}
}
