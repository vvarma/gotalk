package config

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Config interface {
	PrivKey() crypto.PrivKey
	PeerId() peer.ID
	Username() string
	SetPeerId(id peer.ID)
	SetUsername(username string)
	Save()error
}
