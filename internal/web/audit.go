package web

import (
	"net/http"
	"tactile-review/internal/application"
)

func AuditEndpoint(a *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			allow(w, "GET")
			return
		}
		v, e := a.AuditTrail()
		if e != nil {
			respondError(w, 500, "AUDIT_ERROR", e.Error())
			return
		}
		contentTypeJSON(w)
		writeJSON(w, v)
	}
}
