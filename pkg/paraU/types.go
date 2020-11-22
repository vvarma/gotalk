package paraU

import (
	"context"
	"github.com/vvarma/gotalk/pkg/paraU/chat"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
)

type Controller interface {
	AddFriend(ctx context.Context, userName string) error
}

type ParaU interface {
	Controller
	dost.Store
	chat.Chatter
}

type Options struct {
	Username string
}
