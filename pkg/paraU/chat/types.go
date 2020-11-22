package chat

import (
	"bufio"
	"context"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
)

type Chatter interface {
	Start(ctx context.Context, to *dost.Dost, rw *bufio.ReadWriter) error
}
