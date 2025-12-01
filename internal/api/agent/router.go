package agent

import (
	"net/http"
)

// Router creates and configures the HTTP router
type Router struct {
	mux *http.ServeMux
}

// NewRouter creates a new Router instance
func NewRouter(handler *Handler) *Router {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /check-manager", handler.CheckManager)
	return &Router{mux: mux}
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
