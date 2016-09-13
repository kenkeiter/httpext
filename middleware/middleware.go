package middleware

import (
	"net/http"
)

type Handler func(http.Handler) http.Handler
type Set struct {
	m []Handler
}

// Empty indicates whether any middleware have been defined.
func (m *Set) Empty() bool {
	return len(m.m) == 0
}

// Use allows the registration of one or more middleware http.Handlers that are
// aware of the middleware chain.
//
// Middleware are executed in FIFO order. The first middleware you use will
// be the first executed for each request.
func (m *Set) Use(newMiddleware Handler) {
	m.m = append(m.m, newMiddleware)
}

// UseHandler allows the registration of one or more http.Handler interfaces
// that will be executed before the primary request handler.
//
// Middleware are executed in FIFO order. The first middleware handler you use
// will be the first to be executed for each request.
func (m *Set) UseHandler(h http.Handler) {
	f := func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(w, req)
			n.ServeHTTP(w, req)
		})
	}
	m.m = append(m.m, f)
}

// Apply applies middleware to a handler.
func (m *Set) Apply(h http.Handler) http.Handler {
	n := h
	if !m.Empty() {
		for i := len(m.m) - 1; i >= 0; i-- {
			n = m.m[i](n)
		}
	}
	return n
}
