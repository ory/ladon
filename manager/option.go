package manager

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/ory/ladon/policy"
)

type mdKey uint16

const (
	connKey mdKey = iota + 1
)

type Options struct {
	Connection  string
	Driver      string
	Database    string
	TablePrefix string
	TLSConfig   *tls.Config
	Policies    policy.Policies
	Metadata    context.Context
}

func (o *Options) GetConnection() (conn interface{}, ok bool) {
	val := o.Metadata.Value(connKey)
	return val, (val != nil)
}

type Option func(*Options)

func Address(addr string) Option {
	return func(opts *Options) {
		if opts.Connection != "" {
			log.Printf("warning: overwriting previous %s option '%s' with '%s'",
				"Address", addr, opts.Connection)
		}
		opts.Connection = addr
	}
}

func ConnectionString(conn string) Option {
	return func(opts *Options) {
		if opts.Connection != "" {
			log.Printf("warning: overwriting previous %s option '%s' with '%s'",
				"ConnectionString", conn, opts.Connection)
		}
		opts.Connection = conn
	}
}

func Driver(driver string) Option {
	return func(opts *Options) {
		opts.Driver = driver
	}
}

func Database(db string) Option {
	return func(opts *Options) {
		opts.Database = db
	}
}

func TablePrefix(table string) Option {
	return func(opts *Options) {
		opts.TablePrefix = table
	}
}

func TLSConfig(cfg *tls.Config) Option {
	return func(opts *Options) {
		opts.TLSConfig = cfg
	}
}

func Policy(pol policy.Policy) Option {
	return func(opts *Options) {
		opts.Policies = append(opts.Policies, pol)
	}
}

func Connection(conn interface{}) Option {
	return func(opts *Options) {
		opts.Metadata = context.WithValue(opts.Metadata, connKey, conn)
	}
}
