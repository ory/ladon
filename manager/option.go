package manager

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/ory/ladon/log"
)

type mdKey uint16

const (
	connKey mdKey = iota + 1
)

// Options store configuration that should be generally useful for connecting
// to any sort of persistence mechanism that a manager implementation might
// use. If other fields are needed, arbitrary data can be stored inside of the
// Metadata `context.Context`.
//
// One special case of storing data in the Metadata field is storing an already
// initialized database connection. The Connection/GetConnection convenience
// functions are provided. The value from GetConnection should get
// type-switched.
type Options struct {
	Connection  string
	Driver      string
	Database    string
	Password    string
	TablePrefix string
	TLSConfig   *tls.Config
	Timeout     time.Duration
	Metadata    context.Context
}

// GetConnection returns the initialized connection that was stored in Metadata
// using `manager.Connection(...)`.
func (o *Options) GetConnection() (conn interface{}, ok bool) {
	val := o.Metadata.Value(connKey)
	return val, (val != nil)
}

// Option is a functional option. Users may implement additional options that
// store/read from Metadata, which can store arbitrary key-value data.
type Option func(*Options)

// Address stores the connection string. It is an alias for `ConnectionString`
// and implies that it will be in host:port format.
func Address(addr string) Option {
	return func(opts *Options) {
		if opts.Connection != "" {
			log.Printf("warning: overwriting previous %s option '%s' with '%s'",
				"Address", addr, opts.Connection)
		}
		opts.Connection = addr
	}
}

// ConnectionString stores the connection string. It is an alias for `Address`
// and implies that more than just the host and port will be used to connect.
func ConnectionString(conn string) Option {
	return func(opts *Options) {
		if opts.Connection != "" {
			log.Printf("warning: overwriting previous %s option '%s' with '%s'",
				"ConnectionString", conn, opts.Connection)
		}
		opts.Connection = conn
	}
}

// Driver sets which driver might be used when the persistence method has an
// extra layer of abstraction. This is most common in SQL where there may be
// different drivers for MySQL, PostgreSQL, etc.
func Driver(driver string) Option {
	return func(opts *Options) {
		opts.Driver = driver
	}
}

// Database sets the database to use.
func Database(db string) Option {
	return func(opts *Options) {
		opts.Database = db
	}
}

// Password is for databases like Redis that use passwords, but do not have
// standardized connection strings.
func Password(pass string) Option {
	return func(opts *Options) {
		opts.Password = pass
	}
}

// TablePrefix can set the prefix for multiple table names, the name of the
// single table, the index of the NoSQL database, or something else entirely
// depending on the manager implementation. Read the manager implementation's
// package for details on how this option is used, if at all.
func TablePrefix(table string) Option {
	return func(opts *Options) {
		opts.TablePrefix = table
	}
}

// TLSConfig sets the client TLS config to use when connection to the database.
func TLSConfig(cfg *tls.Config) Option {
	return func(opts *Options) {
		opts.TLSConfig = cfg
	}
}

// Timeout sets a dial timeout (if available).
func Timeout(dur time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = dur
	}
}

// Connection stores a fully initialized database connection that can be
// retrieved with `Options.GetConnection`.
func Connection(conn interface{}) Option {
	return func(opts *Options) {
		opts.Metadata = context.WithValue(opts.Metadata, connKey, conn)
	}
}
