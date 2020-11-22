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
	Dosts    map[string]*Dost
	Incoming []*Dost
	Outgoing []*Dost
	lock     sync.Mutex
	callback dostEventCalback
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
	ds := &dostStore{Dosts: make(map[string]*Dost)}
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
			ds.Outgoing = append(ds.Outgoing, &Dost{UserName: d.GetUserName(), PeerId: pId})
		case DostE_incoming:
			pId, err := peer.Decode(d.PeerId)
			if err != nil {
				return nil, err
			}
			ds.Incoming = append(ds.Incoming, &Dost{UserName: d.GetUserName(), PeerId: pId})
		}
	}
	return ds, err
}

func NewDostStore() Store {
	return &dostStore{
		Dosts: make(map[string]*Dost),
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
	for _, o := range ds.Outgoing {
		if o.PeerId == id {
			if o.UserName == "" {
				o.UserName = userName
			}
			ds.Dosts[o.UserName] = o
			ds.callback(ctx, Event{Dost: o, EventType: Approved})
		}
	}
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}

func (ds *dostStore) RejectOutgoing(ctx context.Context, id peer.ID) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	for i, o := range ds.Outgoing {
		if o.PeerId == id {
			ds.Outgoing[i] = ds.Outgoing[len(ds.Outgoing)-1]
			ds.Outgoing[len(ds.Outgoing)-1] = nil
			ds.Outgoing = ds.Outgoing[:len(ds.Outgoing)-1]
			ds.callback(ctx, Event{Dost: o, EventType: Rejected})
		}
	}
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}

func (ds *dostStore) RegisterCallback(callback dostEventCalback) {
	ds.callback = callback
}
func (ds *dostStore) send(ctx context.Context, event Event) {
	if ds.callback != nil {
		ds.callback(ctx, event)
	}
}

func (ds *dostStore) AddIncoming(_ context.Context, d *Dost) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ds.Incoming = append(ds.Incoming, d)
	logger.Info("One new Incoming request, total:", len(ds.Incoming))
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}

}
func (ds *dostStore) AddOutgoing(_ context.Context, d *Dost) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ds.Outgoing = append(ds.Outgoing, d)
	err := ds.Save()
	if err != nil {
		logger.Error("Error in saving ", err)
	}
}
func (ds *dostStore) Review(ctx context.Context, reviewFn func(*Dost) bool) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	incoming := ds.Incoming
	ds.Incoming = []*Dost{}
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
