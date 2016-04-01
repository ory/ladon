package ladon

import (
	"net/http"
	"time"

	"github.com/ory-am/common/pkg"
)

const SubjectKey = "subject"

type Context struct {
	Owner     string    `json:"owner"`
	ClientIP  string    `json:"clientIP"`
	Timestamp time.Time `json:"timestamp"`
	UserAgent string    `json:"userAgent"`
}

func NewContext(req *http.Request, owner string) *Context {
	c := &Context{
		Timestamp: time.Now(),
	}
	c.SetOwner(owner)
	c.FromHTTP(req)
	return c
}

func (c *Context) FromHTTP(req *http.Request) *Context {
	c.ClientIP = pkg.GetIP(req)
	c.UserAgent = req.Header.Get("User-Agent")
	return c
}

func (c *Context) SetOwner(owner string) *Context {
	c.Owner = owner
	return c
}
