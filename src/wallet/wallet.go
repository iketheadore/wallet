package wallet

import (
	"errors"
	"os"
	"time"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/util/file"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

type (
	// AssetType determines the asset type that the wallet holds.
	AssetType string

	// Extension determines a file's extension.
	Extension string
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

const (
	// Version determines the wallet file's version.
	Version uint64 = 0

	// KittyAsset represents the "kittycash" asset type.
	KittyAsset AssetType = "kittycash"

	// FileExt is the kittycash file extension.
	FileExt Extension = ".kcw"
)

/*
	<<< TYPES >>>
*/

// FloatingMeta represents the wallet meta that is not saved, but displayed in api.
type FloatingMeta struct {
	Version   uint64 `json:"version"`
	Label     string `json:"label"`
	Encrypted bool   `json:"encrypted"`
	Password  string `json:"-"`
	Saved     bool   `json:"-"`
	Meta
}

// Meta represents the meta that is stored in file.
type Meta struct {
	AssetType AssetType `json:"type"`
	Seed      string    `json:"seed"`
	TS        int64     `json:"timestamp"`
}

// FloatingWallet represents the wallet that is not saved, but displayed in api.
type FloatingWallet struct {
	Meta    FloatingMeta     `json:"meta"`
	Entries []*FloatingEntry `json:"entries"`
}

// Wallet represents the wallet that is stored in memory.
type Wallet struct {
	Meta    FloatingMeta
	Entries []Entry
}

// File represents the wallet that is stored in file.
type File struct {
	Meta    Meta
	Entries []Entry
}

// FileFromRaw extracts File from raw data.
func FileFromRaw(b []byte) (*File, error) {
	var (
		out = new(File)
		err error
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to read wallet file: %v", r)
		}
	}()
	err = encoder.DeserializeRaw(b, out)
	return out, err
}

func (w File) Serialize() []byte {
	return encoder.Serialize(w)
}

/*
	<<< CREATION >>>
*/

type Options struct {
	Label     string `json:"string"`
	Seed      string `json:"seed"`
	Encrypted bool   `json:"encrypted"`
	Password  string `json:"password,omitempty"`
}

func (o *Options) Verify() error {
	if o.Label == "" {
		return errors.New("invalid label")
	}
	if o.Seed == "" {
		return errors.New("invalid seed")
	}
	if o.Encrypted && o.Password == "" {
		return errors.New("invalid password")
	}
	return nil
}

func NewFloatingWallet(options *Options) (*Wallet, error) {
	if err := options.Verify(); err != nil {
		return nil, err
	}

	return &Wallet{
		Meta: FloatingMeta{
			Version:   Version,
			Label:     options.Label,
			Encrypted: options.Encrypted,
			Password:  options.Password,
			Meta: Meta{
				AssetType: KittyAsset,
				Seed:      options.Seed,
				TS:        time.Now().UnixNano(),
			},
		},
		Entries: []Entry{},
	}, nil
}

func LoadFloatingWallet(raw []byte, label, password string) (*Wallet, error) {
	prefix, data, err := ExtractPrefix(raw)
	if err != nil {
		return nil, err
	}
	encrypted := prefix.Encrypted()

	fmt.Printf("WALLET: v(%v) e(%v) n(%v) \n",
		prefix.Version(), prefix.Encrypted(), prefix.Nonce())

	if encrypted {
		pHash := cipher.SumSHA256([]byte(password))
		data, err = cipher.Chacha20Decrypt(data, pHash[:], prefix.Nonce())
		if err != nil {
			return nil, err
		}
	} else {
		password = ""
	}

	wallet, err := FileFromRaw(data)
	if err != nil {
		return nil, err
	} else if wallet == nil {
		return nil, errors.New("failed to read wallet file, maybe due to incorrect credentials")
	}

	return &Wallet{
		Meta: FloatingMeta{
			Version:   prefix.Version(),
			Label:     label,
			Encrypted: encrypted,
			Password:  password,
			Meta:      wallet.Meta,
		},
		Entries: wallet.Entries,
	}, nil
}

func (w *Wallet) Save() error {
	version := w.Meta.Version

	nonce := EmptyNonce()
	if w.Meta.Encrypted {
		nonce = RandNonce()
	}

	prefix := NewPrefix(version, nonce)

	data := w.ToFile().Serialize()
	if w.Meta.Encrypted {
		var err error
		pHash := cipher.SumSHA256([]byte(w.Meta.Password))
		data, err = cipher.Chacha20Encrypt(data, pHash[:], nonce)
		if err != nil {
			return err
		}
	}

	err := SaveBinary(
		LabelPath(w.Meta.Label),
		append(prefix[:], data...),
	)
	if err != nil {
		return err
	}

	w.Meta.Saved = true
	return nil
}

func (w *Wallet) EnsureEntries(n int) error {
	switch {
	case n < 0:
		return errors.New("can not have negative number of entries")
	case n <= w.Count():
		return nil
	}
	sks := cipher.GenerateDeterministicKeyPairs([]byte(w.Meta.Seed), n)
	w.Entries = make([]Entry, n)
	for i := 0; i < n; i++ {
		entry, _ := NewEntry(sks[i])
		w.Entries[i] = *entry
	}

	w.Meta.Saved = false
	return nil
}

func (w *Wallet) Count() int {
	return len(w.Entries)
}

func (w *Wallet) ToFile() *File {
	return &File{
		Meta:    w.Meta.Meta,
		Entries: w.Entries,
	}
}

func (w *Wallet) ToFloating() *FloatingWallet {
	fw := &FloatingWallet{
		Meta:    w.Meta,
		Entries: make([]*FloatingEntry, len(w.Entries)),
	}
	for i, entry := range w.Entries {
		fw.Entries[i] = entry.ToFloating()
	}
	return fw
}

/*
	<<< HELPERS >>>
*/

func SaveBinary(fn string, data []byte) error {
	return file.SaveBinary(fn, data, os.FileMode(0600))
}
