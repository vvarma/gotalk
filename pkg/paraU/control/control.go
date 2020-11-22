package control

import (
	"bufio"
	"context"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/vvarma/gotalk/pkg/paraU/client"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"github.com/vvarma/gotalk/util"
)

const controlProtocol = "/paraU/1.0/control"

var logger = log.Logger("control")

type controller struct {
	c  client.Client
	ds dost.Store
}

func (co *controller) AddFriend(ctx context.Context, peerId string) error {
	pId, err := peer.Decode(peerId)
	if err != nil {
		return err
	}
	stream, err := co.c.Conn().NewStream(ctx, pId, controlProtocol)
	if err != nil {
		return err
	}
	co.ds.AddOutgoing(ctx, &dost.Dost{PeerId: pId})
	return co.introductions(stream)
}

func (co *controller) sendMessage(ctx context.Context, toPeer peer.ID, msg *ControlMessage) error {
	stream, err := co.c.Conn().NewStream(ctx, toPeer, controlProtocol)
	if err != nil {
		return err
	}
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	return util.SizeDelimtedWriter(ctx, rw, msg)

}
func (co *controller) introductions(stream network.Stream) error {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	msg := &ControlMessage{
		Msg: &ControlMessage_Introduction_{
			Introduction: &ControlMessage_Introduction{
				PeerId:   peer.Encode(co.c.Conn().ID()),
				UserName: co.c.Conf().Username(),
			}},
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgSize := len(msgBytes)
	msgBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(msgBuf, uint64(msgSize))
	msgBuf = append(msgBuf, msgBytes...)
	_, err = rw.Write(msgBuf)
	if err != nil {
		return err
	}
	err = rw.Flush()
	if err != nil {
		return err
	}
	return nil
}
func readControlMsg(ctx context.Context, rw *bufio.ReadWriter) (*ControlMessage, error) {
	msg := &ControlMessage{}
	err := util.SizeDelimitedReader(ctx, rw, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
func (co *controller) readLoop(ctx context.Context, rw *bufio.ReadWriter) error {
	for {
		msg, err := readControlMsg(ctx, rw)
		if err != nil {
			return err
		}
		switch body := msg.Msg.(type) {
		case *ControlMessage_Introduction_:
			intro := body.Introduction
			logger.Info("Got an introduction from", intro.PeerId, intro.GetUserName())
			pId, err := peer.Decode(intro.PeerId)
			if err != nil {
				logger.Error("Unable to decode peer Id from introduction", err)
			}
			co.ds.AddIncoming(context.Background(), &dost.Dost{PeerId: pId, UserName: body.Introduction.GetUserName()})
		case *ControlMessage_DostStatusUpdate_:
			switch body.DostStatusUpdate.UpdatedStatus {
			case ControlMessage_DostStatusUpdate_approved:
				pId, err := peer.Decode(body.DostStatusUpdate.ToPeerId)
				if err != nil {
					logger.Error("Weird Peer id obtained", err)
				}
				userName := body.DostStatusUpdate.Meta.GetUserName()
				co.ds.ApproveOutgoing(ctx, pId, userName)
			case ControlMessage_DostStatusUpdate_rejected:
				pId, err := peer.Decode(body.DostStatusUpdate.ToPeerId)
				if err != nil {
					logger.Error("Weird Peer id obtained", err)
				}
				co.ds.RejectOutgoing(ctx, pId)
			}
		}
	}
}

func (co *controller) getStreamHandler() network.StreamHandler {

	return func(stream network.Stream) {
		ctx := context.Background()
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		err := co.readLoop(ctx, rw)
		logger.Error("Exiting the read loop", err)
		//err := co.introductions(stream)
		if err != nil {
			logger.Error("Error in making introductions ", err)
		}
	}
}

func (co *controller) DostEventCallback(ctx context.Context, event dost.Event) {
	switch event.EventType {
	case dost.Approved:
		msg := &ControlMessage{
			Msg: &ControlMessage_DostStatusUpdate_{
				DostStatusUpdate: &ControlMessage_DostStatusUpdate{
					FromPeerId:    peer.Encode(event.Dost.PeerId),
					ToPeerId:      peer.Encode(co.c.Conn().ID()),
					UpdatedStatus: ControlMessage_DostStatusUpdate_approved,
					Meta:          &ControlMessage_DostStatusUpdate_Meta{UserName: co.c.Conf().Username()},
				}}}
		if err := co.sendMessage(ctx, event.Dost.PeerId, msg); err != nil {
			logger.Error("error sending control message", err)
		}

	case dost.Rejected:
		msg := &ControlMessage{
			Msg: &ControlMessage_DostStatusUpdate_{
				DostStatusUpdate: &ControlMessage_DostStatusUpdate{
					FromPeerId:    peer.Encode(event.Dost.PeerId),
					ToPeerId:      peer.Encode(co.c.Conn().ID()),
					UpdatedStatus: ControlMessage_DostStatusUpdate_rejected,
				}}}
		if err := co.sendMessage(ctx, event.Dost.PeerId, msg); err != nil {
			logger.Error("error sending control message", err)
		}
	}
}

func New(c client.Client, ds dost.Store) Controller {
	co := &controller{c: c, ds: ds}
	c.Conn().SetStreamHandler(controlProtocol, co.getStreamHandler())
	return co
}
