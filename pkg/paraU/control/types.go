package control

import (
	"context"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
)

type Controller interface {
	AddFriend(ctx context.Context, peerId string) error
	DostEventCallback(ctx context.Context, event dost.Event)
}
