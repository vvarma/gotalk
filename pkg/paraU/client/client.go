package client

import (
	"context"
	"fmt"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/vvarma/gotalk/pkg/paraU"
	"github.com/vvarma/gotalk/pkg/paraU/config"
	"os"
	"sync"
	"time"
)

var logger = log.Logger("client")

type client struct {
	conn *routedhost.RoutedHost
	conf config.Config
}

func (c *client) Conf() config.Config {
	return c.conf
}

func (c *client) Conn() *routedhost.RoutedHost {
	return c.conn
}

func getConnectOptions(ctx context.Context, conf config.Config) []libp2p.Option {
	opts := []libp2p.Option{
		libp2p.Identity(conf.PrivKey()),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,         // Lowwater
			400,         // HighWater,
			time.Minute, // GracePeriod
		)),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			kDHT, err := dht.New(ctx, h)
			return kDHT, err
		}),
		libp2p.EnableAutoRelay(),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/udp/0/quic", "/ip4/127.0.0.1/udp/0/quic",
		),
	}
	return opts
}

func bootstrapConn(ctx context.Context, conn host.Host) {
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := conn.Connect(ctx, *peerinfo); err != nil {
				logger.Warn(err)
			} else {
				logger.Debug("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()
}

func New(ctx context.Context, options paraU.Options) (Client, error) {
	conf, err := config.LoadConfig()
	if errors.Is(err, os.ErrNotExist) {
		var err error
		conf, err = config.NewIdentity()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	opts := getConnectOptions(ctx, conf)
	conn, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "Error starting a host")
	}
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", conn.ID().Pretty()))
	addr := conn.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	logger.Infof("I am %s", fullAddr)
	bootstrapConn(ctx, conn)
	logger.Debug("Connected to peers", conn.Peerstore().Peers())
	if conf.Username() == "" {
		conf.SetUsername(options.Username)
	}
	if conf.PeerId() == "" {
		conf.SetPeerId(conn.ID())
	}
	err = conf.Save()
	if err != nil {
		return nil, err
	}
	if routedHost, ok := conn.(*routedhost.RoutedHost); !ok {
		return nil, errors.New("Could not cast as routed host")
	} else {
		//pi, err := peer.Decode("QmPgDnkbZgPZiaMMHcWfbRbqSrP73fdG4UEhX526sNAzVT")
		//if err != nil {
		//	return nil, err
		//}
		//ai := peer.AddrInfo{
		//	ID: pi,
		//}
		//err = conn.Connect(ctx, ai)
		//if err != nil {
		//	logger.Info("Failed to connect to peer")
		//}
		//newMultiaddr, err := multiaddr.NewMultiaddr("/ip4/192.168.86.36/udp/51381/quic/p2p/QmPgDnkbZgPZiaMMHcWfbRbqSrP73fdG4UEhX526sNAzVT")
		//if err != nil {
		//	return nil, err
		//}
		//pAddr, err := peer.AddrInfoFromP2pAddr(newMultiaddr)
		//if err != nil {
		//	return nil, err
		//}
		//err = conn.Connect(ctx, *pAddr)
		//if err != nil {
		//	return nil, err
		//}
		//logger.Info("Connection established with", pAddr)
		return &client{conn: routedHost, conf: conf}, nil
	}
}
