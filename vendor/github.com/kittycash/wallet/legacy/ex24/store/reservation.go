package store

import "github.com/skycoin/skycoin/src/cipher"

type ReservationRequest struct {
	KittyID            uint64
	Address            cipher.Address
	DepositCoinType    string
	DepositCoinAddress string
	DepositCoinValue   int64
	Creation           int64
	Expiry             int64
}
