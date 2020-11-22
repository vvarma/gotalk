package client

import (
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/vvarma/gotalk/pkg/paraU/config"
)

type Client interface {
	Conn() *routedhost.RoutedHost
	Conf() config.Config
}

type Options struct {
	Username string
}
