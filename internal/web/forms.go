package web

import (
	"net/http"
	"strings"
)

func formValue(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	if r.Form == nil {
		r.ParseForm()
	}
	return strings.TrimSpace(r.FormValue(key))
}
func required(values ...string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return false
		}
	}
	return true
}
func contentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}
