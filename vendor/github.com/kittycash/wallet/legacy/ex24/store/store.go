package store

import (
	"errors"
)

type StoreOption func(s *Store) error

func SetMetaStorer(db MetaStorer) StoreOption {
	return func(s *Store) error {
		s.metas = db
		return nil
	}
}

func SetOwnershipStorer(db OwnershipStorer) StoreOption {
	return func(s *Store) error {
		s.ownerships = db
		return nil
	}
}

func SetCertificateStorer(db OwnershipCertificateStorer) StoreOption {
	return func(s *Store) error {
		s.certificates = db
		return nil
	}
}

type Store struct {
	metas        MetaStorer
	ownerships   OwnershipStorer
	certificates OwnershipCertificateStorer
	state        State
}

func NewStore(options ...StoreOption) (*Store, error) {
	var s = new(Store)
	for _, option := range options {
		if e := option(s); e != nil {
			return nil, e
		}
	}
	if s.metas == nil || s.ownerships == nil || s.certificates == nil {
		return nil, errors.New("nil error")
	}
	if e := s.state.Load(s.metas, s.ownerships); e != nil {
		return nil, e
	}
	return s, nil
}

func (s *Store) GetAllKitties() []KittyState {
	return s.state.GetAll()
}

func (s *Store) ReserveKitty() {

}
