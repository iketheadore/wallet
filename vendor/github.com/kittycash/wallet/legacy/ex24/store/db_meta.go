package store

import (
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"io/ioutil"
	"os"
	"sync"
)

type MetaStorer interface {
	Append(meta KittyMeta, img []byte) (uint64, bool)
	Count() int
	ListKitties(pageIndex, pageSize int) []KittyMeta
	GetKitty(kID uint64) (KittyMeta, bool)
	GetImage(kID uint64, resolution string) ([]byte, bool)
}

type MetaDBConfig struct {
	Master   bool
	PK       cipher.PubKey
	SK       cipher.SecKey
	FilePath string
}

type MetaDB struct {
	sync.RWMutex `enc:"-"`
	c            MetaDBConfig `enc:"-"`
	Metas        []KittyMeta
	Images       [][]byte
	InnerHash    cipher.SHA256
	Sig          cipher.Sig
}

func NewMetaDB(config MetaDBConfig) (*MetaDB, error) {

	mDB := new(MetaDB)
	mDB.c = config

	if _, e := os.Stat(config.FilePath); os.IsNotExist(e) {
		if config.Master == false {
			return nil, e
		}
		mDB.InnerHash = mDB.hashInner()
		mDB.Sig = cipher.SignHash(mDB.InnerHash, mDB.c.SK)
		if e := mDB.Save(); e != nil {
			return nil, e
		}
	}

	raw, e := ioutil.ReadFile(config.FilePath)
	if e != nil {
		return nil, e
	}
	if e := encoder.DeserializeRaw(raw, &mDB); e != nil {
		return nil, e
	}
	mDB.c = config
	if e := mDB.verify(); e != nil {
		return nil, e
	}

	return mDB, nil
}

func (md *MetaDB) Append(meta KittyMeta, img []byte) (uint64, bool) {
	md.Lock()
	defer md.Unlock()

	// Needs to be master.
	if md.c.Master == false {
		return 0, false
	}

	// Append.
	meta.ID = uint64(len(md.Metas))
	md.Metas = append(md.Metas, meta)
	md.Images = append(md.Images, img)

	// Update signature.
	md.InnerHash = md.hashInner()
	cipher.SignHash(md.InnerHash, md.c.SK)

	return meta.ID, true
}

func (md *MetaDB) Count() int {
	md.RLock()
	defer md.RUnlock()

	return len(md.Metas)
}

func (md *MetaDB) ListKitties(pageIndex, pageSize int) []KittyMeta {
	md.RLock()
	defer md.RUnlock()

	if pageIndex < 0 || pageSize < 0 {
		return []KittyMeta{}
	}
	var (
		start = pageSize * pageIndex
		end   = start + pageSize
	)
	if end > len(md.Metas) {
		end = len(md.Metas)
	}
	if end <= start || start >= len(md.Metas) {
		return []KittyMeta{}
	}
	return md.Metas[start:end]
}

func (md *MetaDB) GetKitty(kID uint64) (KittyMeta, bool) {
	md.RLock()
	defer md.RUnlock()

	if kID >= uint64(len(md.Metas)) {
		return md.Metas[kID], false
	}
	return md.Metas[kID], true
}

func (md *MetaDB) GetImage(kID uint64, resolution string) ([]byte, bool) {
	md.RLock()
	defer md.RUnlock()

	if kID >= uint64(len(md.Metas)) {
		return nil, false
	}
	return md.Images[kID], true
}

func (md MetaDB) Save() error {
	f, e := os.Open(md.c.FilePath)
	if e != nil {
		return e
	}
	if _, e := f.Write(md.serialize()); e != nil {
		return e
	}
	return f.Close()
}

func (md *MetaDB) hashInner() cipher.SHA256 {
	return cipher.SumSHA256(append(
		encoder.Serialize(md.Metas),
		encoder.Serialize(md.Images)...,
	))
}

func (md *MetaDB) serialize() []byte {
	return encoder.Serialize(*md)
}

func (md *MetaDB) verify() error {
	return cipher.VerifySignature(
		md.c.PK,
		md.Sig,
		md.hashInner(),
	)
}
