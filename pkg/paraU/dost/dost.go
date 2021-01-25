package dost

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"io/ioutil"
	"os"
	"sync"
)

var logger = log.Logger("dost")

type Dost struct {
	PeerId   peer.ID
	UserName string
}

type dostStore struct {
	Dosts     map[string]*Dost
	Incoming  map[peer.ID]*Dost
	Outgoing  map[peer.ID]*Dost
	lock      sync.Mutex
	callbacks []dostEventCalback
}

func (ds *dostStore) DostByPeerId(ctx context.Context, peerid peer.ID) *Dost {
	for _, v := range ds.Dosts {
		if v.PeerId == peerid {
			return v
		}
	}
	return nil
}

func (ds *dostStore) DostByUserName(ctx context.Context, userName string) (*Dost, error) {
	if d, ok := ds.Dosts[userName]; ok {
		return d, nil
	}
	return nil, errors.New("unknown dost")
}

func (ds *dostStore) List(ctx context.Context) []*Dost {
	var o []*Dost
	for _, d := range ds.Dosts {
		o = append(o, d)
	}
	return o
}

func (ds *dostStore) ListIncoming(ctx context.Context) []*Dost {
	var o []*Dost
	for _, d := range ds.Incoming {
		o = append(o, d)
	}
	return o
}

func (ds *dostStore) AcceptIncoming(ctx context.Context, peerId peer.ID) error {
	if d, ok := ds.Incoming[peerId]; ok {
		ds.Dosts[d.UserName] = d
		ds.send(ctx, Event{Dost: d, EventType: Approved})
		delete(ds.Incoming, peerId)
		return ds.Save()
	}
	return nil
}

func LoadDostStore() (Store, error) {
	file, err := ioutil.ReadFile(dostFilename())
	if err != nil {
		return nil, err
	}
	msg := &DostStoreE{}
	err = proto.Unmarshal(file, msg)
	if err != nil {
		return nil, err
	}
	ds := &dostStore{Dosts: make(map[string]*Dost), Outgoing: make(map[peer.ID]*Dost), Incoming: make(map[peer.ID]*Dost)}
	for _, d := range msg.Dosts {
		switch d.Status {
		case DostE_accepted:
			pId, err := peer.Decode(d.PeerId)
			if err != nil {
				return nil, err
			}
			ds.Dosts[d.UserName] = &Dost{UserName: d.GetUserName(), PeerId: pId}
		case DostE_outgoing:
			pId, err := peer.Decode(d.PeerId)
			if err != nil {
				return nil, err
			}
			ds.Outgoing[pId] = &Dost{UserName: d.GetUserName(), PeerId: pId}
		case DostE_incoming:
			pId, err := peer.Decode(d.PeerId)
			if err != nil {
				return nil, err
			}
			ds.Incoming[pId] = &Dost{UserName: d.GetUserName(), PeerId: pId}
		}
	}
	return ds, err
}

func NewDostStore() Store {
	return &dostStore{
		Dosts:    make(map[string]*Dost),
		Incoming: make(map[peer.ID]*Dost),
		Outgoing: make(map[peer.ID]*Dost),
	}
}
func (ds *dostStore) Save() error {
	var dosts []*DostE
	for _, d := range ds.Dosts {
		dosts = append(dosts, &DostE{
			PeerId:   peer.Encode(d.PeerId),
			UserName: d.UserName,
			Status:   DostE_accepted,
		})
	}
	for _, d := range ds.Incoming {
		dosts = append(dosts, &DostE{
			PeerId:   peer.Encode(d.PeerId),
			UserName: d.UserName,
			Status:   DostE_incoming,
		})
	}
	for _, d := range ds.Outgoing {
		dosts = append(dosts, &DostE{
			PeerId:   peer.Encode(d.PeerId),
			UserName: d.UserName,
			Status:   DostE_outgoing,
		})
	}
	msg := &DostStoreE{Dosts: dosts}
	msgBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dostFilename(), msgBuf, os.ModePerm)
}
func dostFilename() string {
	return "dost.paraU"
}

func (ds *dostStore) ApproveOutgoing(ctx context.Context, id peer.ID, userName string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	if outgoingDost, ok := ds.Outgoing[id]; ok {
		if outgoingDost.UserName == "" {
			outgoingDost.UserName = userName
		}
		ds.Dosts[userName] = outgoingDost
		ds.send(ctx, Event{
			Dost:      outgoingDost,
			EventType: Approved,
		})
		delete(ds.Outgoing, id)
	}
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}

func (ds *dostStore) RejectOutgoing(ctx context.Context, id peer.ID) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if outgoingDost, ok := ds.Outgoing[id]; ok {
		delete(ds.Outgoing, id)
		ds.send(ctx, Event{Dost: outgoingDost, EventType: Rejected})
	}
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}

func (ds *dostStore) RegisterCallback(callback dostEventCalback) {
	ds.callbacks = append(ds.callbacks, callback)
}
func (ds *dostStore) send(ctx context.Context, event Event) {
	logger.Debug("Sending callback", event.EventType)
	for _, c := range ds.callbacks {
		c(ctx, event)
	}
}

func (ds *dostStore) AddIncoming(_ context.Context, d *Dost) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ds.Incoming[d.PeerId] = d
	logger.Info("One new Incoming request, total:", len(ds.Incoming))
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}

}
func (ds *dostStore) AddOutgoing(_ context.Context, d *Dost) {
	if _, ok := ds.Outgoing[d.PeerId]; ok {
		return
	}
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ds.Outgoing[d.PeerId] = d
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}
func (ds *dostStore) Review(ctx context.Context, reviewFn func(*Dost) bool) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	incoming := ds.Incoming
	ds.Incoming = make(map[peer.ID]*Dost)
	for _, iDost := range incoming {
		if reviewFn(iDost) {
			logger.Info("Accepting friend request from", iDost)
			ds.Dosts[iDost.UserName] = iDost
			ds.send(ctx, Event{Dost: iDost, EventType: Approved})
		} else {
			logger.Info("Rejecting friend request from", iDost)
			ds.send(ctx, Event{Dost: iDost, EventType: Rejected})
		}
	}
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}
