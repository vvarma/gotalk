package gotalk

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"math/rand"
	"os"
	"sync"
	"time"
)

var logger = log.Logger("gotalk")

const protocol = "/gotalk/1.0"

type Chat struct {
	h           host.Host
	connections []*PeerConnection
	mode        string
}
type peerConnectionStatus int

const (
	initStatus peerConnectionStatus = iota
	connectedStatus
	disconnectedStatus
)

type PeerConnection struct {
	peerId   peer.ID
	userName string
	status   peerConnectionStatus
	rw       *bufio.ReadWriter
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

func NewChat(username string, randevous string, mode string) (*Chat, error) {
	ctx := context.Background()
	var kDHT *dht.IpfsDHT
	var err error
	bwCounter := metrics.NewBandwidthCounter()
	go func() {
		t := time.Tick(10 * time.Second)
		for {
			select {
			case <-t:
				s := bwCounter.GetBandwidthTotals()
				logger.Infof("Total: rateIn %f rateOut %f totalIn %d totalOut %d", s.RateIn, s.RateOut, s.TotalIn, s.TotalOut)
			}
		}
	}()

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
		libp2p.BandwidthReporter(bwCounter),
	}
	opts = append(opts, libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/udp/0/quic", // a UDP endpoint for the QUIC transport
	))
	host, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	logger.Info("Host created, I am ", host.ID())
	logger.Info("@", host.Addrs())
	c := Chat{
		h:           host,
		connections: nil,
		mode:        mode,
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
		stream, err := host.NewStream(ctx, peer.ID, protocol)
		if err != nil {
			logger.Warn("Failed to connect to: ", peer.ID, err)
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
		peerId := peer.ID
		go func() {
			t := time.Tick(10 * time.Second)
			for {
				select {
				case <-t:
					stat := bwCounter.GetBandwidthForPeer(peerId)
					logger.Infof("Peer: %s rateIn %f rateOut %f totalIn %d totalOut %d", peerId, stat.RateIn, stat.RateOut, stat.TotalIn, stat.TotalOut)
				}

			}

		}()
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
	if c.mode == "load" {
		if err := c.doLoad(); err != nil {
			logger.Error("Error in loading ", err)
		}
	}
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

func (c *Chat) doLoad() error {
	for i := 0; i < 10000; i += 1 {
		buf := make([]byte, 1024)
		_, err := rand.Read(buf)
		if err != nil {
			return err
		}
		if err := c.Write(string(buf)); err != nil {
			return err
		}
	}
	return nil
}
