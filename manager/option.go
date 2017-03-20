package manager

import (
	"context"
	"crypto/tls"

	"github.com/ory/ladon/policy"
)

type mdKey uint16

const (
	connKey mdKey = iota + 1
)

type Options struct {
	Connection  string
	Database    string
	PolicyTable string
	TLSConfig   *tls.Config
	Policies    policy.Policies
	Metadata    context.Context
}

func (o *Options) GetConnection() (conn interface{}, ok bool) {
	val := o.Metadata.Value(connKey)
	return val, (val != nil)
}

type Option func(*Options)

func ConnectionString(conn string) Option {
	return func(opts *Options) {
		opts.Connection = conn
	}
}

func Database(db string) Option {
	return func(opts *Options) {
		opts.Database = db
	}
}

func PolicyTable(table string) Option {
	return func(opts *Options) {
		opts.PolicyTable = table
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
