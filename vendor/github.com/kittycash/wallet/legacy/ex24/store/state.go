package store

import (
	"errors"
	"fmt"
	"sync"
)

type KittyState struct {
	Meta      KittyMeta          `json:"meta"`
	Ownership KittyOwnershipJson `json:"ownership"`
}

type State struct {
	sync.RWMutex `json:"-"`
	Kitties      []KittyState `json:"kitties"`
}

type LoadStateAction func(kID uint64, meta KittyMeta, ownership KittyOwnership) error

func (s *State) Load(metas MetaStorer, ownerships OwnershipStorer) error {
	var (
		count   = metas.Count()
		kitties = make([]KittyState, count)
	)
	for i := uint64(0); i < uint64(count); i++ {
		meta, ok := metas.GetKitty(i)
		if !ok {
			e := errors.New("missing meta")
			return errors.New(fmt.Sprintf(
				"failed to load state of kitty %d: %v", i, e))
		}
		ownership, e := ownerships.GetOwnershipState(i)
		if e != nil {
			return errors.New(fmt.Sprintf(
				"failed to load state of kitty %d: %v", i, e))
		}
		kitties[i] = KittyState{
			Meta:      meta,
			Ownership: ownership.Json(true),
		}
	}
	s.Lock()
	defer s.Unlock()
	s.Kitties = kitties
	return nil
}

func (s *State) GetAll() []KittyState {
	s.RLock()
	defer s.RUnlock()
	return (*s).Kitties
}
