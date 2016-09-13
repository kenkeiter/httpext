package httpext

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	HeaderNameCORSAllowOrigin   = "Access-Control-Allow-Origin"
	HeaderNameCORSExposeHeaders = "Access-Control-Expose-Headers"
	HeaderNameCORSMaxAge        = "Access-Control-Max-Age"
	HeaderNameCORSAllowCreds    = "Access-Control-Allow-Credentials"
	HeaderNameCORSAllowMethods  = "Access-Control-Allow-Methods"
	HeaderNameCORSAllowHeaders  = "Access-Control-Allow-Headers"
	HeaderNameCORSVary          = "Vary"
)

var (
	ErrUnmatchedCORSOrigin = errors.New("Unmatched CORS origin.")
)

type CORSPolicy struct {
	allowAllOrigins bool
	origins         []string

	allowAllMethods bool
	methods         []string

	allowAllHeaders bool
	allowHeaders    []string

	exposeHeaders []string

	MaxAge           time.Duration
	AllowCredentials bool
}

func (c *CORSPolicy) AllowOrigins(o ...string) {
	c.allowAllOrigins = false
	c.origins = append(c.origins, o...)
}

func (c *CORSPolicy) AllowAllOrigins() {
	c.allowAllOrigins = true
	c.origins = []string{}
}

func (c *CORSPolicy) AllowMethods(m ...string) {
	c.allowAllMethods = false
	c.methods = append(c.methods, m...)
}

func (c *CORSPolicy) AllowAllMethods() {
	c.allowAllMethods = true
	c.methods = []string{}
}

func (c *CORSPolicy) AllowHeaders(h ...string) {
	c.allowAllHeaders = false
	c.allowHeaders = append(c.allowHeaders, h...)
}

func (c *CORSPolicy) AllowAllHeaders() {
	c.allowAllHeaders = true
	c.allowHeaders = []string{}
}

func (c *CORSPolicy) ExposeHeaders(h ...string) {
	c.exposeHeaders = append(c.exposeHeaders, h...)
}

func (c *CORSPolicy) OriginAllowed(o string) bool {
	if c.allowAllOrigins {
		return true
	}
	for _, origin := range c.origins {
		if o == origin {
			return true
		}
	}
	return false
}

// TODO(kk): Optimize this by joining strings and fomratting numbers ahead of time.
func (c *CORSPolicy) WriteHeaders(w http.ResponseWriter, req *http.Request) {
	// write Access-Control-Allow-Origin
	if c.allowAllOrigins {
		w.Header().Set(HeaderNameCORSAllowOrigin, "*")
	} else {
		if len(c.origins) > 1 {
			w.Header().Set(HeaderNameCORSVary, "Origin")
		}
		origin := req.Header.Get("Origin")
		if c.OriginAllowed(origin) {
			w.Header().Set(HeaderNameCORSAllowOrigin, origin)
		} else {
			w.Header().Set(HeaderNameCORSAllowOrigin, "null")
		}
	}
	// write Access-Control-Expose-Headers
	if len(c.exposeHeaders) > 0 {
		w.Header().Set(HeaderNameCORSExposeHeaders, strings.Join(c.exposeHeaders, ", "))
	}
	// write Access-Control-Max-Age
	w.Header().Set(HeaderNameCORSMaxAge, fmt.Sprintf("%d", int(c.MaxAge.Seconds())))
	// write Access-Control-Allow-Credentials
	if c.AllowCredentials {
		w.Header().Set(HeaderNameCORSAllowCreds, "true")
	} else {
		w.Header().Set(HeaderNameCORSAllowCreds, "false")
	}
	// write Access-Control-Allow-Methods
	if c.allowAllMethods {
		w.Header().Set(HeaderNameCORSAllowMethods, "*")
	} else if len(c.methods) > 0 {
		w.Header().Set(HeaderNameCORSAllowMethods, strings.Join(c.methods, ", "))
	}
	// write Access-Control-Allow-Headers
	if c.allowAllHeaders {
		w.Header().Set(HeaderNameCORSAllowHeaders, "*")
	} else if len(c.allowHeaders) > 0 {
		w.Header().Set(HeaderNameCORSAllowHeaders, strings.Join(c.allowHeaders, ", "))
	}
}
