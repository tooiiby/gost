package relay

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/go-gost/gost/pkg/common/util/mux"
	"github.com/go-gost/gost/pkg/common/util/socks"
	"github.com/go-gost/gost/pkg/common/util/udp"
	"github.com/go-gost/gost/pkg/connector"
	"github.com/go-gost/gost/pkg/logger"
	"github.com/go-gost/relay"
)

// Bind implements connector.Binder.
func (c *relayConnector) Bind(ctx context.Context, conn net.Conn, network, address string, opts ...connector.BindOption) (net.Listener, error) {
	log := c.options.Logger.WithFields(map[string]any{
		"network": network,
		"address": address,
	})
	log.Infof("bind on %s/%s", address, network)

	options := connector.BindOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	switch network {
	case "tcp", "tcp4", "tcp6":
		return c.bindTCP(ctx, conn, network, address, log)
	case "udp", "udp4", "udp6":
		return c.bindUDP(ctx, conn, network, address, &options, log)
	default:
		err := fmt.Errorf("network %s is unsupported", network)
		log.Error(err)
		return nil, err
	}
}

func (c *relayConnector) bindTCP(ctx context.Context, conn net.Conn, network, address string, log logger.Logger) (net.Listener, error) {
	laddr, err := c.bind(conn, relay.BIND, network, address)
	if err != nil {
		return nil, err
	}
	log.Debugf("bind on %s/%s OK", laddr, laddr.Network())

	session, err := mux.ServerSession(conn)
	if err != nil {
		return nil, err
	}

	return &tcpListener{
		addr:    laddr,
		session: session,
		logger:  log,
	}, nil
}

func (c *relayConnector) bindUDP(ctx context.Context, conn net.Conn, network, address string, opts *connector.BindOptions, log logger.Logger) (net.Listener, error) {
	laddr, err := c.bind(conn, relay.FUDP|relay.BIND, network, address)
	if err != nil {
		return nil, err
	}
	log.Debugf("bind on %s/%s OK", laddr, laddr.Network())

	ln := udp.NewListener(
		socks.UDPTunClientPacketConn(conn),
		laddr,
		opts.Backlog,
		opts.UDPDataQueueSize, opts.UDPDataBufferSize,
		opts.UDPConnTTL,
		log)

	return ln, nil
}

func (c *relayConnector) bind(conn net.Conn, cmd uint8, network, address string) (net.Addr, error) {
	req := relay.Request{
		Version: relay.Version1,
		Flags:   cmd,
	}

	if c.options.Auth != nil {
		pwd, _ := c.options.Auth.Password()
		req.Features = append(req.Features, &relay.UserAuthFeature{
			Username: c.options.Auth.Username(),
			Password: pwd,
		})
	}
	fa := &relay.AddrFeature{}
	fa.ParseFrom(address)
	req.Features = append(req.Features, fa)
	if _, err := req.WriteTo(conn); err != nil {
		return nil, err
	}

	// first reply, bind status
	resp := relay.Response{}
	if _, err := resp.ReadFrom(conn); err != nil {
		return nil, err
	}

	if resp.Status != relay.StatusOK {
		return nil, fmt.Errorf("bind on %s/%s failed", address, network)
	}

	var addr string
	for _, f := range resp.Features {
		if f.Type() == relay.FeatureAddr {
			if fa, ok := f.(*relay.AddrFeature); ok {
				addr = net.JoinHostPort(fa.Host, strconv.Itoa(int(fa.Port)))
			}
		}
	}

	var baddr net.Addr
	var err error
	switch network {
	case "tcp", "tcp4", "tcp6":
		baddr, err = net.ResolveTCPAddr(network, addr)
	case "udp", "udp4", "udp6":
		baddr, err = net.ResolveUDPAddr(network, addr)
	default:
		err = fmt.Errorf("unknown network %s", network)
	}
	if err != nil {
		return nil, err
	}

	return baddr, nil
}
