package ladon

import (
	"github.com/ory-am/common/pkg"
	"net/http"
	"time"
)

const SubjectKey = "subject"

type Context struct {
	Owner     string    `json:"owner"`
	ClientIP  string    `json:"clientIP"`
	Timestamp time.Time `json:"timestamp"`
	UserAgent string    `json:"userAgent"`
}

func NewContext() *Context {
	return &Context{
		Timestamp: time.Now(),
	}
}

func (c *Context) HTTP(req *http.Request) *Context {
	c.ClientIP = pkg.GetIP(req)
	c.UserAgent = req.Header.Get("User-Agent")
	return c
}

func (c *Context) Owner(owner string) *Context {
	c.Owner = owner
	return c
}
