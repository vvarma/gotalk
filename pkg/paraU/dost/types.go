package dost

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
)

type dostEventCalback func(ctx context.Context, event Event)
type dostEventType int

const (
	Approved dostEventType = iota
	Rejected
)

type Event struct {
	Dost      *Dost
	EventType dostEventType
}
type Store interface {
	ApproveOutgoing(ctx context.Context, id peer.ID, userName string)
	RejectOutgoing(ctx context.Context, id peer.ID)
	RegisterCallback(callback dostEventCalback)
	AddIncoming(_ context.Context, d *Dost)
	AddOutgoing(_ context.Context, d *Dost)
	Review(ctx context.Context, reviewFn func(*Dost) bool)
	List(ctx context.Context)[]*Dost
	ListIncoming(ctx context.Context) []*Dost
	AcceptIncoming(ctx context.Context, peerId peer.ID) error
	DostByUserName(ctx context.Context, userName string) (*Dost, error)
	DostByPeerId(ctx context.Context, peerid peer.ID) *Dost
	Save() error
}
