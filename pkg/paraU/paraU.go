package paraU

import (
	"context"
	"errors"
	"github.com/vvarma/gotalk/pkg/paraU/chat"
	"github.com/vvarma/gotalk/pkg/paraU/client"
	"github.com/vvarma/gotalk/pkg/paraU/control"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"os"
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
