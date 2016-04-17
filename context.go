package ladon

import (
	"net/http"
	"time"

	"github.com/ory-am/common/pkg"
)

// Context is used as request's context.
type Context struct {
	Owner     string    `json:"owner"`
	ClientIP  string    `json:"clientIP"`
	Timestamp time.Time `json:"timestamp"`
	UserAgent string    `json:"userAgent"`
}

// NewContext creates a new Context.
func NewContext(req *http.Request, owner string) *Context {
	c := &Context{
		Timestamp: time.Now(),
	}
	c.SetOwner(owner)
	c.FromHTTP(req)
	return c
}

// FromHTTP hydrates the context using http.Request.
func (c *Context) FromHTTP(req *http.Request) *Context {
	c.ClientIP = pkg.GetIP(req)
	c.UserAgent = req.Header.Get("User-Agent")
	return c
}

// SetOwner sets the resource's owner.
func (c *Context) SetOwner(owner string) *Context {
	c.Owner = owner
	return c
}
