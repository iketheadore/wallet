package wallet

import "github.com/skycoin/skycoin/src/cipher/go-bip39"

const (
	SeedBitSize = 128
)

func NewSeed(seedBitSize int) (string, error) {
	entropy, e := bip39.NewEntropy(seedBitSize)
	if e != nil {
		return "", e
	}
	mnemonic, e := bip39.NewMnemonic(entropy)
	if e != nil {
		return "", e
	}
	return mnemonic, nil
}
