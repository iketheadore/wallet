package store

import (
	"encoding/json"
)

type KittyMeta struct {
	ID           uint64   `json:"ID"`
	Name         string   `json:"name"`
	Bio          string   `json:"bio"`
	Breed        string   `json:"breed"`
	Attributes   []string `json:"attributes"`
	PriceBitcoin uint64   `json:"price_bitcoin"`
	PriceSkycoin uint64   `json:"price_skycoin"`
}

func (km KittyMeta) Encode() []byte {
	raw, _ := json.Marshal(km)
	return raw
}

func KittyMetaFromRaw(raw []byte) (KittyMeta, error) {
	var km KittyMeta
	if e := json.Unmarshal(raw, &km); e != nil {
		return km, e
	}
	return km, nil
}
