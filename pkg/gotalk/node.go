package gotalk

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery2 "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type NodeConfig struct {
	UserName string
}

type Node struct {
	Config          *NodeConfig
	Host            host.Host
	DHT             routing.Routing
	Discovery       discovery.Discovery
	Topics          []string
	PeerConnections []*PeerConnection
}

func Open(ctx context.Context, username string) (*Node, error) {
	configFile, err := ioutil.ReadFile(fmt.Sprintf("%s.config", username))
	config := &NodeConfig{}
	if errors.Is(err, os.ErrNotExist) {
		config.UserName = username
	} else {
		err := json.Unmarshal(configFile, config)
		if err != nil {
			return nil, err
		}
	}

	//bwCounter := metrics.NewBandwidthCounter()
	var kDHT routing.Routing
	opts := []libp2p.Option{
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
			kDHT, err = dht.New(ctx, h)
			return kDHT, err
		}),
		libp2p.EnableAutoRelay(),
		//libp2p.BandwidthReporter(bwCounter),
	}
	opts = append(opts, libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/udp/0/quic",
		"/ip4/127.0.0.1/udp/0/quic",
	))
	p2pHost, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p2pHost.Connect(ctx, *peerinfo); err != nil {
				logger.Warn(err)
			} else {
				logger.Debug("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s.config", username), configBytes, os.ModePerm)
	if err != nil {
		return nil, err
	}
	routingDiscovery := discovery2.NewRoutingDiscovery(kDHT)
	topics := []string{
		//username,
		"gotalk-public-2",
	}
	n := &Node{
		Config:    config,
		Host:      p2pHost,
		DHT:       kDHT,
		Topics:    topics,
		Discovery: routingDiscovery,
	}
	adminStreamHandler(n)
	for _, topic := range topics {
		_, err := routingDiscovery.Advertise(ctx, topic)
		if err != nil {
			return nil, err
		}
	}
	err = n.syncPeers(ctx)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Node) syncPeers(ctx context.Context) error {
	peers := map[peer.ID]*PeerConnection{}
	for _, pc := range n.PeerConnections {
		peers[pc.peerId] = pc
	}
	wg := sync.WaitGroup{}
	for _, topic := range n.Topics {
		newPeers, err := n.Discovery.FindPeers(ctx, topic)
		if err != nil {
			return err
		}
		for np := range newPeers {
			if np.ID == n.Host.ID() {
				continue
			}
			if _, ok := peers[np.ID]; ok {
				continue
			}
			pc := &PeerConnection{peerId: np.ID, status: initStatus}
			n.PeerConnections = append(n.PeerConnections, pc)
			wg.Add(1)
			go func() {
				defer wg.Done()
				logger.Info("Connecting to peer", pc.peerId)
				stream, err := n.Host.NewStream(ctx, adminProtocol)
				if err != nil {
					logger.Error("Could not connect to peer ", pc.peerId, err)
					pc.status = disconnectedStatus
					return
				}
				pc.status = connectedStatus
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				err = writeIdentity(rw, n.Host.ID().Pretty(), n.Config)
				if err != nil {
					logger.Error("Could not write identity to peer ", pc.peerId, err)
				}
				identity, err := readIdentity(rw)
				if err != nil {
					logger.Error("Could not read peers identity", pc.peerId, err)
				}
				pc.userName = identity.GetIdentity().GetUsername()
				logger.Info("Got user for peer", pc.peerId, pc.userName, pc.status)
			}()
		}
	}
	wg.Wait()
	return nil
}

func (n *Node) AddPeer(newPC *PeerConnection) {
	logger.Info("Found a new user ", newPC.userName)
	for _, pc := range n.PeerConnections {
		if pc.peerId == newPC.peerId {
			if pc.userName == "" {
				pc.userName = newPC.userName
			} else if pc.userName != newPC.userName {
				logger.Errorf("Imposter? peer %s known username %s new username %s", pc.peerId, pc.userName, newPC.userName)
			}
			return
		}
	}
	n.PeerConnections = append(n.PeerConnections, newPC)
}
