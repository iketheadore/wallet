package wallet

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

var (
	ErrWalletNotFound     = errors.New("wallet of label is not found")
	ErrWalletLocked       = errors.New("wallet is locked")
	ErrLabelAlreadyExists = errors.New("label already exists")
)

type ManagerConfig struct {
	RootDir string
}

func (mc *ManagerConfig) Process() error {
	var err error
	if mc.RootDir, err = filepath.Abs(mc.RootDir); err != nil {
		return err
	}
	if err = os.MkdirAll(mc.RootDir, os.FileMode(0700)); err != nil {
		return err
	}
	return nil
}

// Manager manages the wallet files.
type Manager struct {
	c       *ManagerConfig
	mux     sync.Mutex
	labels  []string
	wallets map[string]*Wallet
}

// NewManager creates a new wallet manager.
func NewManager(config *ManagerConfig) (*Manager, error) {
	m := &Manager{
		c: config,
	}
	if err := m.c.Process(); err != nil {
		return nil, err
	}
	if err := m.Refresh(); err != nil {
		return nil, err
	}
	return m, nil
}

// Refresh reloads the list of wallets.
// All wallets will be locked.
func (m *Manager) Refresh() error {
	defer m.lock()()

	m.labels = make([]string, 0)
	m.wallets = make(map[string]*Wallet)
	err := RangeLabels(m.c.RootDir, func(raw []byte, label, fPath string, prefix Prefix) error {
		if prefix.Version() != Version {
			log.Warningf(
				"wallet file `%s` is of version %v, while only version %v is supported",
				label, prefix.Version(), Version)
			return nil
		}
		var wallet *Wallet
		if prefix.Encrypted() == false {
			var err error
			if wallet, err = LoadWallet(raw, label, ""); err != nil {
				return err
			}
		}
		m.append(label, wallet)
		return nil
	})
	if err != nil {
		return err
	}
	return m.sort()
}

// Stat represents a wallet when listed by 'ListWallets'.
type Stat struct {
	Label     string `json:"label"`
	Encrypted bool   `json:"encrypted"`
	Locked    *bool  `json:"locked,omitempty"`
}

// Lists the wallets available.
func (m *Manager) ListWallets() []Stat {
	defer m.lock()()

	var out = make([]Stat, len(m.labels))
	for i, label := range m.labels {
		fw := m.wallets[label]
		var (
			encrypted bool
			locked    *bool
		)
		if fw == nil {
			encrypted = true
			locked = new(bool)
			*locked = true
		} else {
			encrypted = fw.Meta.Encrypted
			if encrypted {
				locked = new(bool)
				*locked = false
			}
		}
		out[i] = Stat{
			Label:     label,
			Encrypted: encrypted,
			Locked:    locked,
		}
	}
	return out
}

// NewWallet creates a new wallet (and it's associated file)
// with specified options, and the number of addresses to generate under it.
func (m *Manager) NewWallet(opts *Options, addresses int) error {
	defer m.lock()()

	if addresses < 0 {
		return errors.New("can not have negative number of entries")
	}

	if _, ok := m.wallets[opts.Label]; ok {
		return ErrLabelAlreadyExists
	}

	fw, e := NewWallet(opts)
	if e != nil {
		return e
	}
	if e := fw.EnsureEntries(addresses); e != nil {
		return e
	}
	if e := fw.Save(m.c.RootDir); e != nil {
		return e
	}
	m.append(opts.Label, fw)
	return m.sort()
}

// DeleteWallet deletes a wallet of a given label.
func (m *Manager) DeleteWallet(label string) error {
	defer m.lock()()

	if m.remove(label) {
		return os.Remove(LabelPath(m.c.RootDir, label))
	}
	return ErrWalletNotFound
}

// DisplayWallet displays the wallet of specified label.
// Password needs to be given if a wallet is still locked.
// Addresses ensures that wallet has at least the number of address entries.
func (m *Manager) DisplayWallet(label, password string, addresses int) (*FloatingWallet, error) {
	defer m.lock()()

	switch w, err := m.getWallet(label); err {
	case nil:
		if err := w.EnsureEntries(addresses); err != nil {
			return nil, err
		}
		if !w.Meta.Saved {
			if err := w.Save(m.c.RootDir); err != nil {
				return nil, err
			}
		}
		return w.ToFloating(), nil

	case ErrWalletNotFound:
		return nil, ErrWalletNotFound

	case ErrWalletLocked:
		raw, err := OpenAndReadAll(LabelPath(m.c.RootDir, label))
		if err != nil {
			return nil, err
		}
		if w, err = LoadWallet(raw, label, password); err != nil {
			return nil, err
		}
		m.wallets[label] = w
		if err := w.EnsureEntries(addresses); err != nil {
			return nil, err
		}
		if !w.Meta.Saved {
			if err := w.Save(m.c.RootDir); err != nil {
				return nil, err
			}
		}
		return w.ToFloating(), nil

	default:
		return nil, errors.New("unknown error")
	}
}

func (m *Manager) DisplayPaginatedWallet(label, password string, startIndex, pageSize, forceTotal int) (*PaginatedFloatingWallet, error) {
	defer m.lock()()

	toPaginatedTotal := func(w *Wallet, startIndex, pageSize, forceTotal int) (*PaginatedFloatingWallet, error) {
		if forceTotal != -1 {
			if err := w.EnsureEntries(forceTotal); err != nil {
				return nil, err
			}
			if !w.Meta.Saved {
				if err := w.Save(m.c.RootDir); err != nil {
					return nil, err
				}
			}
		}
		return w.ToPaginatedFloating(startIndex, pageSize)
	}

	switch w, err := m.getWallet(label); err {
	case nil:
		return toPaginatedTotal(w, startIndex, pageSize, forceTotal)

	case ErrWalletNotFound:
		return nil, ErrWalletNotFound

	case ErrWalletLocked:
		raw, err := OpenAndReadAll(LabelPath(m.c.RootDir, label))
		if err != nil {
			return nil, err
		}
		if w, err = LoadWallet(raw, label, password); err != nil {
			return nil, err
		}
		m.wallets[label] = w
		return toPaginatedTotal(w, startIndex, pageSize, forceTotal)

	default:
		return nil, errors.New("unknown error")
	}
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func (m *Manager) lock() func() {
	m.mux.Lock()
	return m.mux.Unlock
}

func (m *Manager) append(label string, fw *Wallet) {
	m.labels = append(m.labels, label)
	m.wallets[label] = fw
}

func (m *Manager) remove(label string) bool {
	for i, l := range m.labels {
		if l == label {
			m.labels = append(m.labels[:i], m.labels[i+1:]...)
			delete(m.wallets, label)
			return true
		}
	}
	return false
}

func (m *Manager) sort() error {
	sort.Strings(m.labels)
	return nil
}

func (m *Manager) getWallet(label string) (*Wallet, error) {
	w, ok := m.wallets[label]
	if !ok {
		return nil, ErrWalletNotFound
	}
	if w == nil {
		return nil, ErrWalletLocked
	}
	return w, nil
}
