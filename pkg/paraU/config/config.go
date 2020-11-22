package config

import (
	"crypto/rand"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"io/ioutil"
	"os"
)

type config struct {
	priv     crypto.PrivKey
	peerId   peer.ID
	username string
}

func (i *config) SetPeerId(id peer.ID) {
	i.peerId = id
}

func (i *config) SetUsername(username string) {
	i.username = username
}

func (i *config) PrivKey() crypto.PrivKey {
	return i.priv
}

func (i *config) PeerId() peer.ID {
	return i.peerId
}

func (i *config) Username() string {
	return i.username
}

func configFileName() string {
	return "config.parau"
}

func LoadConfig() (Config, error) {
	file, err := ioutil.ReadFile(configFileName())
	if err != nil {
		return nil, err
	}
	msg := &StoredConfig{}
	err = proto.Unmarshal(file, msg)
	if err != nil {
		return nil, err
	}
	key, err := crypto.UnmarshalPrivateKey(msg.EncodedKey)
	if err != nil {
		return nil, err
	}
	pId, err := peer.Decode(msg.PeerId)
	if err != nil {
		return nil, err
	}
	return &config{
		priv:     key,
		username: msg.GetUsername(),
		peerId:   pId,
	}, nil

}
func NewIdentity() (*config, error) {
	r := rand.Reader
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}
	return &config{priv: priv}, nil

}
func (i *config) Save() error {
	keyEncoded, err := crypto.MarshalPrivateKey(i.priv)
	if err != nil {
		return err
	}
	msg := &StoredConfig{EncodedKey: keyEncoded, Username: i.username, PeerId: peer.Encode(i.peerId)}
	msgBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFileName(), msgBuf, os.ModePerm)
	return err
}
