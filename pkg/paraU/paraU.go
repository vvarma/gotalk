package paraU

import (
	"context"
	"errors"
	"github.com/vvarma/gotalk/pkg/paraU/chat"
	"github.com/vvarma/gotalk/pkg/paraU/client"
	"github.com/vvarma/gotalk/pkg/paraU/control"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"os"
	"time"
)

type p struct {
	client.Client
	control.Controller
	dost.Store
	chat.Chatter
}

func New(ctx context.Context, opts client.Options) (ParaU, error) {
	c, err := client.New(ctx, opts)
	if err != nil {
		return nil, err
	}
	ds, err := dost.LoadDostStore()
	if errors.Is(err, os.ErrNotExist) {
		ds = dost.NewDostStore()
	} else if err != nil {
		return nil, err
	}
	co := control.New(c, ds)
	ds.RegisterCallback(co.DostEventCallback)
	chatter := chat.New(ctx, c, ds)
	return &p{
		c, co, ds, chatter,
	}, nil

}

func (_p *p) status(s Status) Status {
	if len(_p.Conn().Addrs()) != len(s.System.Addresses) {
		var addresses []string
		for _, ma := range _p.Conn().Addrs() {
			addresses = append(addresses, ma.String())
		}
		s.System.Addresses = addresses
	}
	if len(_p.Conn().Peerstore().Peers()) != s.System.NumPeers {
		s.System.NumPeers = len(_p.Conn().Peerstore().Peers())
	}
	return s
}
func (_p *p) Updates() <-chan Status {
	uc := make(chan Status, 10)
	s := Status{}
	s.Self.PeerId = _p.Conf().PeerId()
	s.Self.Username = _p.Conf().Username()
	s = _p.status(s)
	uc <- s
	go func() {
		tc := time.Tick(time.Second * 30)
		for {
			select {
			case <-tc:
				s = _p.status(s)
				uc <- s
			}
		}
	}()
	return uc
}
