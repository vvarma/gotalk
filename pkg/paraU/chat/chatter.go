package chat

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/vvarma/gotalk/pkg/paraU/client"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"github.com/vvarma/gotalk/util"
	"os"
	"sync"
)

const chatProtocol = "/paraU/1.0/chat"

var logger = log.Logger("chat")

type chatLogIO struct {
	readF  *os.File
	writeF *os.File
}

func (cio *chatLogIO) append(ctx context.Context, msg *ChatMessage) error {
	w := bufio.NewWriter(cio.writeF)
	return util.SizeDelimtedWriter(ctx, w, msg)
}

type store struct {
	openChats map[string]*chatLogIO
	lock      sync.Mutex
}

func (s *store) loadOrOpen(ctx context.Context, dostName string) (*chatLogIO, error) {
	if cIO, ok := s.openChats[dostName]; ok {
		return cIO, nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if cIO, ok := s.openChats[dostName]; ok {
		return cIO, nil
	}
	writeF, err := os.OpenFile(fmt.Sprintf("%s.chat.paraU", dostName), os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}
	readF, err := os.Open(fmt.Sprintf("%s.chat.paraU", dostName))
	if err != nil {
		return nil, err
	}
	return &chatLogIO{
		writeF: writeF,
		readF:  readF,
	}, nil

}

func (s *store) appendMessage(ctx context.Context, dostName string, msg *ChatMessage) error {
	cio, err := s.loadOrOpen(ctx, dostName)
	if err != nil {
		return err
	}
	return cio.append(ctx, msg)
}

type chatter struct {
	c  client.Client
	s  *store
	ds dost.Store

	activeConnections map[string]*bufio.ReadWriter
	connectionLock    sync.Mutex
	activeChats       map[string]*bufio.ReadWriter
	chatLock          sync.Mutex
}

func New(ctx context.Context, c client.Client, ds dost.Store) Chatter {
	ch := &chatter{
		c:                 c,
		s:                 &store{},
		activeChats:       make(map[string]*bufio.ReadWriter),
		activeConnections: make(map[string]*bufio.ReadWriter),
		ds:                ds}
	c.Conn().SetStreamHandler(chatProtocol, ch.streamHandler())
	return ch
}
func (ch *chatter) connectionIn(ctx context.Context, cxRw *bufio.ReadWriter) {
	for {
		msg := &ChatMessage{}
		err := util.SizeDelimitedReader(ctx, cxRw.Reader, msg)
		if err != nil {
			logger.Error("Breaking from read loop", err)
			break
		}
		pId, err := peer.Decode(msg.Meta.FromPeer)
		if err != nil {
			logger.Error("Malformed peer in message", err)
			break
		}
		d := ch.ds.DostByPeerId(ctx, pId)
		if d == nil {
			logger.Error("Message from unknown peer, discarding", peer.Encode(pId))
			break
		}
		if _, ok := ch.activeConnections[d.UserName]; !ok {
			ch.activeConnections[d.UserName] = cxRw
		}
		err = ch.s.appendMessage(ctx, d.UserName, msg)
		if err != nil {
			logger.Error("unable to save message", err)
		}
		if chRw, ok := ch.activeChats[d.UserName]; ok {
			switch body := msg.Msg.(type) {
			case *ChatMessage_Text_:
				_, err := chRw.WriteString(body.Text.GetBody())
				if err != nil {
					logger.Error("Error writing to chat out", err)
				}
				err = chRw.Flush()
				if err != nil {
					logger.Error("Error flushing to chat out", err)
				}
			}
		}
	}
}
func (ch *chatter) connectionOut(ctx context.Context, cxRw *bufio.ReadWriter, chRw *bufio.ReadWriter, fromPeer, toPeer peer.ID, dostName string) error {
	for {
		readString, err := chRw.ReadString('\n')
		if err != nil {
			return err
		}
		msg := &ChatMessage{
			Msg:  &ChatMessage_Text_{Text: &ChatMessage_Text{Body: readString}},
			Meta: &ChatMessage_Meta{FromPeer: peer.Encode(fromPeer), ToPeer: peer.Encode(toPeer)},
		}
		err = ch.s.appendMessage(ctx, dostName, msg)
		if err != nil {
			return err
		}
		err = util.SizeDelimtedWriter(ctx, cxRw.Writer, msg)
		if err != nil {
			return err
		}
	}
}

func (ch *chatter) streamHandler() network.StreamHandler {
	return func(stream network.Stream) {
		ctx := context.Background()
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		ch.connectionIn(ctx, rw)
	}
}

func (ch *chatter) Start(ctx context.Context, to *dost.Dost, rw *bufio.ReadWriter) error {
	if _, ok := ch.activeChats[to.UserName]; !ok {
		ch.activeChats[to.UserName] = rw
	}
	var activeConn *bufio.ReadWriter
	var ok bool
	if activeConn, ok = ch.activeConnections[to.UserName]; !ok {
		stream, err := ch.c.Conn().NewStream(ctx, to.PeerId, chatProtocol)
		if err != nil {
			return err
		}
		activeConn = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	}

	go func() {
		ch.connectionIn(ctx, activeConn)
	}()
	return ch.connectionOut(ctx, activeConn, rw, ch.c.Conf().PeerId(), to.PeerId, to.UserName)
}
