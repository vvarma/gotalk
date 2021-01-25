package paraU

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/vvarma/gotalk/pkg/paraU/chat"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"strings"
)

type Controller interface {
	AddFriend(ctx context.Context, peerID string) error
}

type Status struct {
	Self struct {
		Username string
		PeerId   peer.ID
	}
	System struct {
		Protocols []string
		Addresses []string
		Peers     []peer.ID
		NumPeers  int
	}
	Dosts struct {
		NumFriends  int
		NumIncoming int
		NumOutgoing int
	}
}

const statusFmt string = `
Self
Username: %s 
PeerId: %s

System
NumPeers: %d
Protocols: %s

`

func (s Status) String() string {
	return fmt.Sprintf(statusFmt, s.Self.Username, s.Self.PeerId.String(),
		s.System.NumPeers, strings.Join(s.System.Protocols, ","))
}

type ParaU interface {
	Controller
	dost.Store
	chat.Chatter
	Updates() <-chan Status
}
