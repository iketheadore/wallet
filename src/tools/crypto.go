package tools

import (
	"context"

	"github.com/kittycash/kittiverse/src/kitty"
	"github.com/pkg/errors"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

type SignTransferParamsIn struct {
	KittyID         kitty.ID // ID of kitty to transfer.
	LastTransferSig string   // Signature of last kitty transfer (leave empty if not needed).
	ToAddress       string   // Address to transfer kitty to (destination address).
	SecretKey       string   // Secret key of current owner, used to sign the kitty away!
}

type SignTransferParamsOut struct {
	Data string `json:"data"`
	Hash string `json:"hash"`
	Sig  string `json:"sig"`
}

// TransferParams defines the parameters used in a initiate transfer signature
type TransferParams struct {
	KittyID               kitty.ID
	LastTransferSignature string
	DestAddress           string
}

func SignTransferParams(ctx context.Context, in *SignTransferParamsIn) (*SignTransferParamsOut, error) {

	// Obtain last sig.
	var lastSig string
	switch in.LastTransferSig {
	case "":
		lastSig = ""
	default:
		sig, err := cipher.SigFromHex(in.LastTransferSig)
		if err != nil {
			return nil, errors.WithMessage(err, "provided last transfer signature is invalid")
		}
		lastSig = sig.Hex()
	}

	// Obtain destination address.
	dstAddr, err := cipher.DecodeBase58Address(in.ToAddress)
	if err != nil {
		return nil, errors.WithMessage(err, "provided destination address is invalid")
	}

	// Obtain secret key.
	secKey, err := cipher.SecKeyFromHex(in.SecretKey)
	if err != nil {
		return nil, errors.WithMessage(err, "provided secret key is invalid")
	}

	// Sign.
	var (
		params = TransferParams{
			KittyID:               in.KittyID,
			LastTransferSignature: lastSig,
			DestAddress:           dstAddr.String(),
		}
		data = encoder.Serialize(params)
		hash = cipher.SumSHA256(data)
		sig  = cipher.SignHash(hash, secKey)
	)

	// Output.
	return &SignTransferParamsOut{
		Data: string(data),
		Hash: hash.Hex(),
		Sig:  sig.Hex(),
	}, nil
}
