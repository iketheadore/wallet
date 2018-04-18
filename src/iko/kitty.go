package iko

import (
	"github.com/kittycash/kittiverse/src/kitty"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"

	"github.com/kittycash/wallet/src/iko/transaction"
)

/*
	<<< KITTY STATE >>>
	>>> The state of a kitty as represented when the IKO Chain is compiled.
*/

type KittyState struct {
	Address      cipher.Address
	Transactions transaction.IDs
}

func (s KittyState) Serialize() []byte {
	return encoder.Serialize(s)
}

/*
	<<< ADDRESS STATE >>>
	>>> The state of an address as represented when the IKO Chain is compiled.
*/

type AddressState struct {
	Kitties      kitty.IDs
	Transactions transaction.IDs
}

func NewAddressState() *AddressState {
	return &AddressState{
		Kitties:      make(kitty.IDs, 0),
		Transactions: make(transaction.IDs, 0),
	}
}

func (a AddressState) Serialize() []byte {
	return encoder.Serialize(a)
}
