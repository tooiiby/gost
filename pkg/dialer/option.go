package dialer

import (
	"crypto/tls"
	"net/url"

	"github.com/go-gost/gost/pkg/logger"
)

type Options struct {
	Auth      *url.Userinfo
	TLSConfig *tls.Config
	Logger    logger.Logger
}

type Option func(opts *Options)

func AuthOption(auth *url.Userinfo) Option {
	return func(opts *Options) {
		opts.Auth = auth
	}
}

func TLSConfigOption(tlsConfig *tls.Config) Option {
	return func(opts *Options) {
		opts.TLSConfig = tlsConfig
	}
}

func LoggerOption(logger logger.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

type DialOptions struct {
	Host      string
	NetDialer *NetDialer
}

type DialOption func(opts *DialOptions)

func HostDialOption(host string) DialOption {
	return func(opts *DialOptions) {
		opts.Host = host
	}
}

func NetDialerDialOption(netd *NetDialer) DialOption {
	return func(opts *DialOptions) {
		opts.NetDialer = netd
	}
}

type HandshakeOptions struct {
	Addr string
}

type HandshakeOption func(opts *HandshakeOptions)

func AddrHandshakeOption(addr string) HandshakeOption {
	return func(opts *HandshakeOptions) {
		opts.Addr = addr
	}
}
