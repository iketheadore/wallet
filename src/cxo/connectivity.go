package cxo

import (
	"github.com/kittycash/wallet/src/connectivity"
	"github.com/skycoin/cxo/node"
)

type Connectivity struct {
	c *Config
	n *node.Node
}

func (c *Connectivity) Status() (connectivity.Status, error) {
	connections := c.n.Connections()
	if len(connections) > 0 {
		return connectivity.Connected, nil
	}
	return connectivity.Disconnected, nil
}

func (c *Connectivity) Reconnect() bool {
	for _, address := range c.c.MessengerAddresses {
		c.n.TCP().ConnectToDiscoveryServer(address)
	}
	feeds := c.n.Feeds()
	for _, pk := range feeds {
		c.n.Container().AddFeed(pk)
	}
	return true
}