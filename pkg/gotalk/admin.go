package gotalk

import (
	"bufio"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const adminProtocol = "/gotalk/1.0/admin"

func writeIdentity(rw *bufio.ReadWriter, peerId string, config *NodeConfig) error {
	msg := &AdminMessage{Identity: &AdminMessage_Identity{Username: config.UserName}, CurrentHost: &AdminMessage_CurrentHost{PeerId: peerId}}
	marshal, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	size := len(marshal)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(size))
	_, err = rw.Write(bs)
	if err != nil {
		return err
	}
	_, err = rw.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
func readIdentity(rw *bufio.ReadWriter) (*AdminMessage, error) {
	bs := make([]byte, 4)
	_, err := rw.Read(bs)
	if err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint64(bs)
	msg := make([]byte, size)
	_, err = rw.Read(msg)
	if err != nil {
		return nil, err
	}
	adminMsg := &AdminMessage{}
	err = proto.Unmarshal(msg, adminMsg)
	if err != nil {
		return nil, err
	}
	return adminMsg, err

}

func adminStreamHandler(node *Node) {
	handler := func(stream network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		err := writeIdentity(rw, node.Host.ID().Pretty(), node.Config)
		if err != nil {
			logger.Error("Error writing Identity", err)
		}
		adminMsg, err := readIdentity(rw)
		if err != nil {
			logger.Error("Error reading Identify", err)
		}
		peerId, err := peer.Decode(adminMsg.CurrentHost.PeerId)
		if err != nil {
			logger.Error("Unable to decode peer", err)
		} else {
			pc := &PeerConnection{peerId: peerId, userName: adminMsg.GetIdentity().GetUsername()}
			node.AddPeer(pc)
		}
	}
	node.Host.SetStreamHandler(adminProtocol, handler)
}
