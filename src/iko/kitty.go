package iko

import (
	"sort"
	"strconv"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"encoding/json"
)

/*
	<<< KITTY DETAILS >>>
	>>> Used by multiple services, provides off-chain details for kitties and IKO.
*/

type Kitty struct {
	ID    KittyID `json:"kitty_id"`    // Identifier for kitty.
	Name  string  `json:"name"`        // Name of kitty.
	Desc  string  `json:"description"` // Description of kitty.
	Breed string  `json:"breed"`       // Kitty breed.

	PriceBTC    int64  `json:"price_btc"`   // Price of kitty in BTC.
	PriceSKY    int64  `json:"price_sky"`   // Price of kitty in SKY.

	BoxOpen   bool   `json:"box_open"`   // Whether box is open.
	BirthDate int64  `json:"birth_date"` // Timestamp of box opening.
	KittyDNA  string `json:"kitty_dna"`  // Hex representation of kitty DNA (after box opening).

	BoxImgURL   string `json:"box_image_url"`   // Box image URL.
	KittyImgURL string `json:"kitty_image_url"` // Kitty image URL.
}

func (k *Kitty) Json() []byte {
	raw, _ := json.Marshal(k)
	return raw
}

func (k *Kitty) Hash() cipher.SHA256 {
	return cipher.SumSHA256(k.Json())
}

func (k *Kitty) Sign(sk cipher.SecKey) cipher.Sig {
	return cipher.SignHash(k.Hash(), sk)
}

func (k *Kitty) Verify(pk cipher.PubKey, sig cipher.Sig) error {
	return cipher.VerifySignature(pk, sig, k.Hash())
}

/*
	<<< KITTY ID >>>
	>>> For IKO, kitties are indexed with IDs, not DNA.
*/

type KittyID uint64

func KittyIDFromString(idStr string) (KittyID, error) {
	id, e := strconv.ParseUint(idStr, 10, 64)
	return KittyID(id), e
}

type KittyIDs []KittyID

func (ids KittyIDs) Sort() {
	sort.Slice(ids, func(i, j int) bool {
		return (ids)[i] < (ids)[j]
	})
}

func (ids *KittyIDs) Add(id KittyID) {
	*ids = append(*ids, id)
	ids.Sort()
}

func (ids *KittyIDs) Remove(id KittyID) {
	for i, v := range *ids {
		if v == id {
			*ids = append((*ids)[:i], (*ids)[i+1:]...)
			return
		}
	}
}

/*
	<<< KITTY STATE >>>
	>>> The state of a kitty as represented when the IKO Chain is compiled.
*/

type KittyState struct {
	Address      cipher.Address
	Transactions TxHashes
}

func (s KittyState) Serialize() []byte {
	return encoder.Serialize(s)
}

/*
	<<< ADDRESS STATE >>>
	>>> The state of an address as represented when the IKO Chain is compiled.
*/

type AddressState struct {
	Kitties      KittyIDs
	Transactions TxHashes
}

func NewAddressState() *AddressState {
	return &AddressState{
		Kitties:      make(KittyIDs, 0),
		Transactions: make(TxHashes, 0),
	}
}

func (a AddressState) Serialize() []byte {
	return encoder.Serialize(a)
}
