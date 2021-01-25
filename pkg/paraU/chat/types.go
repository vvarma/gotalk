package chat

import (
	"context"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
)

type Callback interface {
	OnIncoming(ctx context.Context, msg *ChatMessage)
	OnOutgoin(ctx context.Context, msg *ChatMessage)
}

type Chatter interface {
	Start(ctx context.Context, to *dost.Dost) error
	Send(ctx context.Context, to *dost.Dost, msg string) error
	Register(ctx context.Context, callback Callback)
	Read(ctx context.Context, to *dost.Dost) ([]*ChatMessage, error)
}
