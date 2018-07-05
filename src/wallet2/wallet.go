package wallet2

import (
	"errors"
	"github.com/skycoin/skycoin/src/cipher"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

type (
	// AssetType determines the asset type that the wallet holds.
	AssetType string
)

var (
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("failed to read wallet file, maybe due to incorrect credentials")
)

const (
	// Version determines the wallet file's version.
	Version uint64 = 0

	// KittyAsset represents the "kittycash" asset type.
	KittyAsset AssetType = "kittycash"

	// FileExt is the kittycash file extension.
	FileExt = ".kcw"
)

// Meta represents wallet meta that's stored on disk.
type Meta struct {
	AssetType AssetType `json:"type"`
	Seed      string    `json:"seed"`
	TS        int64     `json:"timestamp"`
}

// Entry represents a wallet entry that is stored on disk.
type Entry struct {
	Address cipher.Address
	PubKey  cipher.PubKey
	SecKey  cipher.SecKey
}

// Wallet represents a wallet that is stored on disk.
type Wallet struct {
	Meta    Meta
	Entries []Entry
}

// WalletFromRaw loads a wallet from raw bytes.
func WalletFromRaw(b []byte) (w *Wallet, err error) {
	w = new(Wallet)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to read wallet file: %v", r)
		}
	}()
	err = encoder.DeserializeRaw(b, w)
	return
}

