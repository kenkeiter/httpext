package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	ms := &Set{}
	assert.True(t, ms.Empty(), "Newly-created middleware sets should be empty.")

	checks := []int{}
	h0 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checks = append(checks, 0)
	})
	ms.UseHandler(h0)
	h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checks = append(checks, 1)
	})
	ms.UseHandler(h1)

	h2 := func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			checks = append(checks, 2)
			n.ServeHTTP(w, r)
		})
	}
	ms.Use(h2)

	h3 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checks = append(checks, 3)
	})
	hnd := ms.Apply(h3)
	hnd.ServeHTTP(nil, nil)
	assert.Equal(t, []int{0, 1, 2, 3}, checks, "HandlerFunc chain should run completely.")

}
