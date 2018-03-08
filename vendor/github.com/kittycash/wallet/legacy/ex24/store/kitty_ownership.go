package store

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

type OwnershipState string

const (
	StateNoOwner  OwnershipState = "no_owner"
	StateReserved OwnershipState = "reserved"
	StateOwned    OwnershipState = "owned"
)

type KittyOwnership struct {
	KId     uint64
	State   OwnershipState
	Address cipher.Address
}

func (ko KittyOwnership) Json(ignoreReserved bool) KittyOwnershipJson {
	if ignoreReserved {
		ko.State = StateNoOwner
	}
	switch ko.State {
	case StateNoOwner:
		return KittyOwnershipJson{
			KId:   ko.KId,
			State: ko.State,
		}
	case StateOwned, StateReserved:
		return KittyOwnershipJson{
			KId:     ko.KId,
			State:   ko.State,
			Address: ko.Address.String(),
		}
	default:
		panic(errors.New("unknown OwnershipState"))
	}
}

type KittyOwnershipJson struct {
	KId     uint64         `json:"kitty_id"`
	State   OwnershipState `json:"state"`
	Address string         `json:"address,omitempty"`
}

type OwnershipCertificate struct {
	Timestamp          int64
	KId                uint64
	OwnerAddress       cipher.Address
	DepositCoinType    string
	DepositCoinAddress string
	DepositCoinValue   int64

	InnerHash cipher.SHA256
	Sig       cipher.Sig
}

func (oc *OwnershipCertificate) GetInnerHash() cipher.SHA256 {
	return cipher.SumSHA256(bytes.Join([][]byte{
		encoder.Serialize(oc.Timestamp),
		encoder.Serialize(oc.KId),
		encoder.Serialize(oc.OwnerAddress),
		encoder.Serialize(oc.DepositCoinType),
		encoder.Serialize(oc.DepositCoinAddress),
		encoder.Serialize(oc.DepositCoinValue),
	}, nil))
}

func (oc *OwnershipCertificate) Sign(sk cipher.SecKey) {
	oc.InnerHash = oc.GetInnerHash()
	oc.Sig = cipher.SignHash(oc.InnerHash, sk)
}

func (oc *OwnershipCertificate) Verify(pk cipher.PubKey) error {
	oc.InnerHash = oc.GetInnerHash()
	return cipher.VerifySignature(pk, oc.Sig, oc.InnerHash)
}

func (oc *OwnershipCertificate) Serialize() []byte {
	return encoder.Serialize(*oc)
}

func (oc *OwnershipCertificate) SerializeToHexString() string {
	return hex.EncodeToString(oc.Serialize())
}

func (oc *OwnershipCertificate) Json() OwnershipCertificateJson {
	return OwnershipCertificateJson{
		Timestamp:          oc.Timestamp,
		KId:                oc.KId,
		OwnerAddress:       oc.OwnerAddress.String(),
		DepositCoinType:    oc.DepositCoinType,
		DepositCoinAddress: oc.DepositCoinAddress,
		DepositCoinValue:   oc.DepositCoinValue,
		InnerHash:          oc.InnerHash.Hex(),
		Sig:                oc.Sig.Hex(),
		HexRaw:             oc.SerializeToHexString(),
	}
}

type OwnershipCertificateJson struct {
	Timestamp          int64  `json:"timestamp"`
	KId                uint64 `json:"kitty_id"`
	OwnerAddress       string `json:"owner_address"`
	DepositCoinType    string `json:"deposit_coin_type"`
	DepositCoinAddress string `json:"deposit_coin_address"`
	DepositCoinValue   int64  `json:"deposit_coin_value"`
	InnerHash          string `json:"inner_hash"`
	Sig                string `json:"signature"`
	HexRaw             string `json:"hex_raw"`
}
