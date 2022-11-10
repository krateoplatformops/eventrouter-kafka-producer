package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

func HealtHandler(healthy *int32, name, version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(healthy) == 1 {
			data := map[string]string{
				"name":    name,
				"version": version,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}
