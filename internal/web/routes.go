package web

import "net/http"

func routeMethod(r *http.Request, methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	return false
}
func allow(w http.ResponseWriter, methods string) {
	w.Header().Set("Allow", methods)
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func pathSegments(path string) []string {
	var out []string
	for _, p := range splitPath(path) {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
func splitPath(path string) []string {
	var out []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			out = append(out, path[start:i])
			start = i + 1
		}
	}
	out = append(out, path[start:])
	return out
}
