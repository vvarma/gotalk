package gotalk

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"os"
	"sync"
)

var logger = log.Logger("gotalk")

const protocol = "/gotalk/1.0"

type Chat struct {
	h           host.Host
	connections []*PeerConnection
}
type peerConnectionStatus int

const (
	initStatus peerConnectionStatus = iota
	connectedStatus
	disconnectedStatus
)

type PeerConnection struct {
	peerId peer.ID
	status peerConnectionStatus
	rw     *bufio.ReadWriter
}

func (pc *PeerConnection) read() {
	for {
		if pc.status != connectedStatus {
			return
		}
		line, err := pc.rw.ReadString('\n')
		if err != nil {
			logger.Error("Error reading from peer ", err)
			pc.status = disconnectedStatus
			break
		}
		fmt.Printf("\x1b[32m%s\x1b[0m>", line)
	}
	logger.Debug("Done with reading")
}
func (pc *PeerConnection) write(line string) error {
	logger.Debug("Writing ", line)
	if pc.status != connectedStatus {
		return nil
	}
	_, er := pc.rw.WriteString(fmt.Sprintf("%s\n", line))
	if er != nil {
		return er
	}
	return pc.rw.Flush()
}

func NewChat(username string, randevous string, listenAddress string) (*Chat, error) {
	ctx := context.Background()
	var opts []libp2p.Option
	if listenAddress != "" {
		//opts = append(opts, libp2p.ListenAddrStrings(listenAddress))
		extMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listenAddress, 14891))
		if err != nil {
			logger.Errorf("Error creating multiaddress: %v\n", err)
			return nil, err
		}
		addressFactory := func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
			if extMultiAddr != nil {
				addrs = append(addrs, extMultiAddr)
			}
			return addrs
		}
		opts = append(opts, libp2p.AddrsFactory(addressFactory))
	}
	host, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	logger.Info("Host created, I am ", host.ID())
	logger.Info("@", host.Addrs())
	c := Chat{
		h:           host,
		connections: nil,
	}
	host.SetStreamHandler(protocol, func(stream network.Stream) {
		logger.Debug("Got an incoming stream")
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		pc := &PeerConnection{
			peerId: "incomingUnknownPeer",
			status: connectedStatus,
			rw:     rw,
		}
		c.connections = append(c.connections, pc)
		go pc.read()
	})
	kDHT, err := dht.New(ctx, host)
	if err != nil {
		return nil, err
	}
	if err = kDHT.Bootstrap(ctx); err != nil {
		logger.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				logger.Warn(err)
			} else {
				logger.Debug("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()
	logger.Debug("Peers:", kDHT.RoutingTable().ListPeers())
	routingDiscovery := discovery.NewRoutingDiscovery(kDHT)
	discovery.Advertise(ctx, routingDiscovery, randevous)
	logger.Debug("Successfully announced!")
	peerChan, err := routingDiscovery.FindPeers(ctx, randevous)
	if err != nil {
		panic(err)
	}
	var peerConnections []*PeerConnection
	for peer := range peerChan {
		if peer.ID == host.ID() {
			continue
		}
		logger.Debug("New peer ", peer.ID)
		stream, err := host.NewStream(ctx, peer.ID, "/gotalk/1.0")
		if err != nil {
			logger.Warn("Failed to connect to: ", peer.ID)
			continue
		}
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		pc := &PeerConnection{
			peerId: peer.ID,
			status: connectedStatus,
			rw:     rw,
		}
		peerConnections = append(peerConnections, pc)
		go pc.read()
		logger.Info("Connected to peer: ", peer.ID)
	}
	logger.Info("Found all peers known")
	c.connections = append(c.connections, peerConnections...)
	return &c, nil
}
func (c *Chat) Write(line string) error {
	var err error
	for _, pc := range c.connections {
		er := pc.write(line)
		if er != nil {
			err = multierror.Append(err, er)
		}
	}
	return err
}
func (c *Chat) Input() {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		inputData, err := stdReader.ReadString('\n')
		if err != nil {
			logger.Error("Error reading stdin", err)
		}
		err = c.Write(inputData)
		if err != nil {
			logger.Error("Error writing to peers", err)
		}
	}
}

func (c *Chat) Close() error {
	return c.Close()
}
