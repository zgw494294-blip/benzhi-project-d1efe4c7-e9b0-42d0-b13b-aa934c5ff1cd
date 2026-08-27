package web

import "net/http"

func method(w http.ResponseWriter, ok bool) {
	if !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
