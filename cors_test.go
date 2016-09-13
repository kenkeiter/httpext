package httpext

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// corsPolicyTest sets up a test of a CORS policy, returning a pointer to a
// CORSPolicy, the http.Request that will be made, and a func() which triggers
// the test, returning an http.ResponseWriter.
func corsPolicyTest(t *testing.T) (*CORSPolicy, *http.Request, func() http.ResponseWriter) {
	c := &CORSPolicy{}
	req, err := http.NewRequest("OPTIONS", "/example", nil)
	if err != nil {
		t.Fatal(err)
	}
	f := func() http.ResponseWriter {
		w := httptest.NewRecorder()
		c.WriteHeaders(w, req)
		return w
	}
	return c, req, f
}

func TestCORSWildcardOrigin(t *testing.T) {
	c, req, apply := corsPolicyTest(t)
	c.AllowAllOrigins()
	req.Header.Set("Origin", "http://example.com")
	resp := apply()

	assert.Equal(t, "*", resp.Header().Get(HeaderNameCORSAllowOrigin),
		"Permissive CORS request should accept all origins.")
	assert.Empty(t, resp.Header().Get(HeaderNameCORSVary),
		"Vary header should be empty.")
}

func TestCORSExplicitOrigin(t *testing.T) {
	c, req, apply := corsPolicyTest(t)
	// set the origin, and then allow that origin
	testOrigin := "http://example.com"
	req.Header.Set("Origin", testOrigin)
	c.AllowOrigins(testOrigin)
	// run the test, and validate the response
	resp := apply()
	assert.Equal(t, testOrigin, resp.Header().Get(HeaderNameCORSAllowOrigin),
		"Access-Control-Allow-Origin should match accepted origin.")
	assert.Empty(t, resp.Header().Get("Vary"),
		"Vary header should be empty unless the server supports more than one origin.")

	// Add an additional origin.
	c.AllowOrigins("http://google.com")
	resp = apply()
	assert.Equal(t, "Origin", resp.Header().Get("Vary"),
		"Vary header should be set if more than one origin is supported by the server.")

	// Server defines origin (non-wildcard), but it does not match the Origin
	// that the client provides.
	// http://www.w3.org/TR/cors/#access-control-allow-origin-response-header
	req.Header.Set("Origin", "http://kenkeiter.com")
	resp = apply()
	assert.Equal(t, "null", resp.Header().Get(HeaderNameCORSAllowOrigin),
		"If server defines origins, and an unknown origin is provided, server should respond with null.")
}

func TestCORSExposeHeaders(t *testing.T) {
	c, _, apply := corsPolicyTest(t)
	c.ExposeHeaders("X-Test-Header", "X-Another-Test-Header")
	resp := apply()

	assert.Equal(t, "X-Test-Header, X-Another-Test-Header",
		resp.Header().Get(HeaderNameCORSExposeHeaders),
		"Exposed headers should be listed in Access-Control-Expose-Headers header.")
}

func TestCORSMaxAge(t *testing.T) {
	c, _, apply := corsPolicyTest(t)
	c.MaxAge = time.Duration(time.Second * 60)
	resp := apply()

	assert.Equal(t, "60", resp.Header().Get(HeaderNameCORSMaxAge),
		"Access-Control-Max-Age header should indicate the maximum allowed cache duration.")
}

func TestCORSAllowCreds(t *testing.T) {
	c, _, apply := corsPolicyTest(t)

	c.AllowCredentials = true
	resp := apply()
	assert.Equal(t, "true", resp.Header().Get(HeaderNameCORSAllowCreds),
		"Access-Control-Allow-Credentials header should set to true when enabled.")

	c.AllowCredentials = false
	resp = apply()
	assert.Equal(t, "false", resp.Header().Get(HeaderNameCORSAllowCreds),
		"Access-Control-Allow-Credentials header should set to false when disabled.")
}

func TestCORSAllowMethods(t *testing.T) {
	c, _, apply := corsPolicyTest(t)

	c.AllowAllMethods()
	resp := apply()
	assert.Equal(t, "*", resp.Header().Get(HeaderNameCORSAllowMethods),
		"Access-Control-Allow-Methods header should be wildcard when all are accepted.")

	c.AllowMethods("GET", "POST")
	resp = apply()
	assert.Equal(t, "GET, POST", resp.Header().Get(HeaderNameCORSAllowMethods),
		"Access-Control-Allow-Methods header should contain list of methods when "+
			"limited set of methods are allowed.")
}

func TestCORSAllowHeaders(t *testing.T) {
	c, _, apply := corsPolicyTest(t)

	c.AllowAllHeaders()
	resp := apply()
	assert.Equal(t, "*", resp.Header().Get(HeaderNameCORSAllowHeaders),
		"Access-Control-Allow-Headers header should be wildcard when all are accepted.")

	c.AllowHeaders("X-Test-Header", "X-Another-Test-Header")
	resp = apply()
	assert.Equal(t, "X-Test-Header, X-Another-Test-Header",
		resp.Header().Get(HeaderNameCORSAllowHeaders),
		"Access-Control-Allow-Headers header should contain list of headers when "+
			"a specific subset is allowed.")
}
