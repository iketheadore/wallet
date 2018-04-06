package cxo

import (
	"github.com/kittycash/wallet/src/iko"
	"github.com/skycoin/cxo/skyobject/registry"
	"github.com/skycoin/skycoin/src/cipher"
)

type Container struct {
	Meta  []byte
	Txs   registry.Refs `skyobject:"schema=iko.Transaction"`
	Metas registry.Refs `skyobject:"schema=iko.TxMeta"`
}

var (
	reg = registry.NewRegistry(func(r *registry.Reg) {
		r.Register("cipher.Address", cipher.Address{})
		r.Register("iko.Transaction", iko.Transaction{})
		r.Register("iko.TxMeta", iko.TxMeta{})
		r.Register("iko.Container", Container{})
	})
)
