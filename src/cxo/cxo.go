package cxo

import (
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/cxo/skyobject/registry"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"

	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/util"
)

/*
	<<< CONFIG >>>
*/

type Config struct {
	Dir                string
	Public             bool
	Memory             bool
	MessengerAddresses []string
	CXOAddress         string
	CXORPCAddress      string

	MasterRooter    bool
	MasterRootPK    cipher.PubKey
	MasterRootSK    cipher.SecKey
	MasterRootNonce uint64 // Public
}

func (c *Config) Process(log *logrus.Logger) error {
	if e := c.MasterRootPK.Verify(); e != nil {
		return e
	}
	if c.MasterRooter {
		if e := c.MasterRootSK.Verify(); e != nil {
			return e
		}
		if c.MasterRootPK != cipher.PubKeyFromSecKey(c.MasterRootSK) {
			return errors.New("public and secret keys do not match")
		}
	}
	if c.Memory {
		c.Dir = ""
	}
	return nil
}

/*
	<<< CXO >>>
*/

type CXO struct {
	mux      sync.Mutex
	c        *Config
	l        *logrus.Logger
	node     *node.Node
	wg       sync.WaitGroup
	received chan *iko.TxWrapper
	accepted chan *iko.TxWrapper

	len util.SafeInt
}

func New(config *Config, modifyNC ...NodeConfigModifier) (*CXO, error) {
	log := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	if e := config.Process(log); e != nil {
		return nil, e
	}
	cxo := &CXO{
		c:        config,
		l:        log,
		received: make(chan *iko.TxWrapper),
		accepted: make(chan *iko.TxWrapper),
	}
	var modify NodeConfigModifier
	if len(modifyNC) > 0 {
		modify = modifyNC[0]
	}
	if e := prepareNode(cxo, modify); e != nil {
		return nil, e
	}
	if e := initChain(cxo); e != nil {
		log.WithError(e).
			Info("no blockchain yet")
	} else {
		log.WithField("height", cxo.len.Val()).
			Info("blockchain re-initialized")
	}
	return cxo, nil
}

func (c *CXO) Close() {
	defer c.lock()()

	close(c.received)
	c.wg.Wait()
	close(c.accepted)

	if e := c.node.Close(); e != nil {
		c.l.WithError(e).
			Error("error on cxo node close")
	}
	c.l.Print("closed CXO chain DB")
}

/*
	<<< PREP AND SERVICE FUNCTIONS >>>
*/

func (c *CXO) lock() func() {
	c.mux.Lock()
	return c.mux.Unlock
}

func (c *CXO) attemptPushAccepted(txWrap *iko.TxWrapper) {
	select {
	case c.accepted <- txWrap:
	default:
	}
}

type NodeConfigModifier func(nc *node.Config) error

func prepareNode(cxo *CXO, modifier NodeConfigModifier) error {

	nc := node.NewConfig()

	nc.DataDir = cxo.c.Dir
	nc.Public = cxo.c.Public
	nc.InMemoryDB = cxo.c.Memory

	nc.TCP.Listen = cxo.c.CXOAddress
	if len(cxo.c.MessengerAddresses) > 0 {
		nc.TCP.Discovery = node.Addresses(cxo.c.MessengerAddresses)
	}
	nc.RPC = cxo.c.CXORPCAddress

	nc.OnRootReceived = func(c *node.Conn, r *registry.Root) error {
		defer cxo.lock()()

		switch {
		case r.Pub != cxo.c.MasterRootPK:
			e := errors.New("received root is not of master public key")
			cxo.l.
				WithField("master_pk", cxo.c.MasterRootPK.Hex()).
				WithField("received_pk", r.Pub.Hex()).
				Warning(e.Error())
			return e

		case r.Nonce != cxo.c.MasterRootNonce:
			e := errors.New("received root is not of master nonce")
			cxo.l.
				WithField("master_nonce", cxo.c.MasterRootNonce).
				WithField("received_nonce", r.Nonce).
				Warning(e.Error())
			return e

		case len(r.Refs) <= 0:
			e := errors.New("empty refs")
			cxo.l.Warning(e.Error())
			return e

		default:
			cxo.l.Info("blockchain syncing")
			return nil
		}
	}

	nc.OnRootFilled = func(n *node.Node, r *registry.Root) {
		defer cxo.lock()()

		e := func(c *CXO, n *node.Node, r *registry.Root) error {
			var container = new(Container)

			p, e := n.Container().Pack(r, reg)
			if e != nil {
				return e
			}

			if e := r.Refs[0].Value(p, container); e != nil {
				return e
			}

			rLen, e := container.Txs.Len(p)
			if e != nil {
				return e
			}

			switch {
			case rLen < c.len.Val():
				return errors.New("received new root has less transactions")

			case rLen == c.len.Val():
				c.l.Info("received new root has no new transactions")
				return nil
			}

			for i := c.len.Val(); i < rLen; i++ {

				var wrapper = new(iko.TxWrapper)

				txHash, e := container.Txs.ValueByIndex(p, int(i), &wrapper.Tx)
				if e != nil {
					return e
				}

				_, e = container.Metas.ValueByIndex(p, int(i), &wrapper.Meta)
				if e != nil {
					return e
				}

				c.l.
					WithField("tx_hash", txHash.Hex()).
					WithField("tx_seq", i).
					Info("received new transaction")

				c.received <- wrapper
			}

			cxo.l.Info("blockchain synced")
			return nil

		}(cxo, n, r)

		if e != nil {
			cxo.l.Error(e.Error())
			return
		}
	}

	nc.OnConnect = func(c *node.Conn) error {
		defer cxo.lock()()

		// TODO: implement.
		return nil
	}

	nc.OnDisconnect = func(c *node.Conn, reason error) {
		defer cxo.lock()()

		// TODO: implement.
	}

	if modifier != nil {
		if e := modifier(nc); e != nil {
			return e
		}
	}

	var e error
	if cxo.node, e = node.NewNode(nc); e != nil {
		return e
	}

	if e := cxo.node.Share(cxo.c.MasterRootPK); e != nil {
		return e
	}

	return nil
}

func (c *CXO) RunTxService(txChecker iko.TxChecker) error {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		for {
			select {
			case txWrapper, ok := <-c.received:
				if !ok {
					return
				} else if err := txChecker(&txWrapper.Tx); err != nil {
					c.l.Println(err.Error())
				} else {
					c.len.Inc()
					c.attemptPushAccepted(txWrapper)
				}
			}
		}
	}()
	return nil
}

func initChain(cxo *CXO) error {
	defer cxo.lock()()

	r, e := getRoot(cxo)
	if e != nil {
		return e
	}

	p, e := getPack(cxo, r)
	if e != nil {
		return e
	}

	container, e := getContainer(r, p)
	if e != nil {
		return e
	}

	sLen, e := container.Txs.Len(p)
	if e != nil {
		return e
	}

	cxo.len.Set(sLen)
	return nil
}

/*
	<<< PUBLIC FUNCTIONS >>>
*/

func (c *CXO) Connectivity() *Connectivity {
	return &Connectivity{
		c: c.c,
		n: c.node,
	}
}

func (c *CXO) MasterInitChain() error {
	defer c.lock()()

	up, e := getUnpack(c)
	if e != nil {
		return e
	}

	cont := new(Container)
	sReg, e := newContainer(up, cont)
	if e != nil {
		return e
	}

	r := &registry.Root{
		Refs:  []registry.Dynamic{sReg},
		Reg:   reg.Reference(),
		Pub:   c.c.MasterRootPK,
		Nonce: c.c.MasterRootNonce,
	}

	if e := c.node.Container().Save(up, r); e != nil {
		return e
	}

	c.node.Publish(r)

	return nil
}

func (c *CXO) Head() (iko.TxWrapper, error) {
	defer c.lock()()

	var (
		txWrap iko.TxWrapper
		cLen   = c.len.Val()
	)

	if cLen < 1 {
		return txWrap, errors.New("no transactions available")
	}

	store, _, p, e := c.getContainer(gsRead)
	if e != nil {
		return txWrap, e
	}
	if _, e := store.Txs.ValueByIndex(p, cLen-1, &txWrap.Tx); e != nil {
		return txWrap, e
	}
	if _, e := store.Metas.ValueByIndex(p, cLen-1, &txWrap.Meta); e != nil {
		return txWrap, e
	}
	return txWrap, nil
}

func (c *CXO) Len() uint64 {
	defer c.lock()()
	return uint64(c.len.Val())
}

func (c *CXO) AddTx(txWrap iko.TxWrapper, check iko.TxChecker) error {

	if c.c.MasterRooter == false {
		return errors.New("not master node")
	}
	if e := check(&txWrap.Tx); e != nil {
		c.l.WithError(e).Error("failed")
		return e
	}

	defer c.lock()()
	cLen := c.len.Val()

	store, r, up, e := c.getContainer(gsWrite)
	if e != nil {
		return e
	}
	if e := store.Txs.AppendValues(up, txWrap.Tx); e != nil {
		return e
	}
	if e := store.Metas.AppendValues(up, txWrap.Meta); e != nil {
		return e
	}
	if e := r.Refs[0].SetValue(up, store); e != nil {
		return e
	}
	if e := c.node.Container().Save(up.(*skyobject.Unpack), r); e != nil {
		return e
	}
	c.node.Publish(r)
	c.len.Set(cLen + 1)
	c.attemptPushAccepted(&txWrap)
	return nil
}

func (c *CXO) GetTxOfHash(hash iko.TxHash) (iko.TxWrapper, error) {
	defer c.lock()()
	var txWrap iko.TxWrapper

	store, _, p, e := c.getContainer(gsRead)
	if e != nil {
		return txWrap, e
	}
	i, e := store.Txs.ValueOfHashWithIndex(p, cipher.SHA256(hash), &txWrap.Tx)
	if e != nil {
		return txWrap, e
	}
	if _, e := store.Metas.ValueByIndex(p, i, &txWrap.Meta); e != nil {
		return txWrap, e
	}
	return txWrap, nil
}

func (c *CXO) GetTxOfSeq(seq uint64) (iko.TxWrapper, error) {
	defer c.lock()()
	var txWrap iko.TxWrapper

	store, _, p, e := c.getContainer(gsRead)
	if e != nil {
		return txWrap, e
	}
	if _, e := store.Txs.ValueByIndex(p, int(seq), &txWrap.Tx); e != nil {
		return txWrap, e
	}
	if _, e := store.Metas.ValueByIndex(p, int(seq), &txWrap.Meta); e != nil {
		return txWrap, e
	}
	return txWrap, nil
}

func (c *CXO) TxChan() <-chan *iko.TxWrapper {
	return c.accepted
}

func (c *CXO) GetTxsOfSeqRange(startSeq uint64, pageSize uint64) ([]iko.TxWrapper, error) {
	defer c.lock()()
	var txWraps []iko.TxWrapper

	if pageSize == 0 {
		return txWraps, fmt.Errorf("invalid pageSize: %d", pageSize)
	}
	cLen := uint64(c.len.Val())
	if startSeq >= cLen {
		return txWraps, fmt.Errorf("invalid startSeq: %d", startSeq)
	}
	if startSeq+pageSize > cLen {
		diff := startSeq + pageSize - cLen
		if pageSize-diff <= 0 {
			return []iko.TxWrapper{}, nil
		}
		pageSize -= diff
	}
	store, _, p, e := c.getContainer(gsRead)
	if e != nil {
		return txWraps, e
	}
	txRefs, e := store.Txs.Slice(p, int(startSeq), int(startSeq+pageSize))
	if e != nil {
		return txWraps, e
	}
	refsLen, e := txRefs.Len(p)
	if e != nil {
		return txWraps, e
	}
	metaRefs, e := store.Metas.Slice(p, int(startSeq), int(startSeq+pageSize))
	if e != nil {
		return txWraps, e
	}
	txWraps = make([]iko.TxWrapper, refsLen)
	e = txRefs.Ascend(p, func(i int, hash cipher.SHA256) error {
		raw, _, e := c.node.Container().Get(hash, 0)
		if e != nil {
			return e
		}
		return encoder.DeserializeRaw(raw, &txWraps[i].Tx)
	})
	if e != nil {
		return txWraps, e
	}
	e = metaRefs.Ascend(p, func(i int, hash cipher.SHA256) error {
		raw, _, e := c.node.Container().Get(hash, 0)
		if e != nil {
			return e
		}
		return encoder.DeserializeRaw(raw, &txWraps[i].Meta)
	})
	return txWraps, e
}

type getContType int

const (
	gsRead  getContType = iota
	gsWrite getContType = iota
)

func (c *CXO) getContainer(t getContType) (*Container, *registry.Root, registry.Pack, error) {
	r, e := getRoot(c)
	if e != nil {
		return nil, nil, nil, e
	}
	var p registry.Pack
	switch t {
	case gsRead:
		if p, e = getPack(c, r); e != nil {
			return nil, nil, nil, e
		}
	case gsWrite:
		if p, e = getUnpack(c); e != nil {
			return nil, nil, nil, e
		}
	default:
		panic("invalid getContType")
	}
	store, e := getContainer(r, p)
	if e != nil {
		return nil, nil, nil, e
	}
	return store, r, p, nil
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func getRoot(c *CXO) (*registry.Root, error) {
	r, e := c.node.Container().LastRoot(c.c.MasterRootPK, c.c.MasterRootNonce)
	if e != nil {
		return nil, e
	}
	return r, nil
}

func getContainer(r *registry.Root, p registry.Pack) (*Container, error) {
	if len(r.Refs) < 1 {
		return nil, errors.New("corrupt root, invalid ref count")
	}
	store := new(Container)
	if e := r.Refs[0].Value(p, store); e != nil {
		return nil, e
	}
	return store, nil
}

func newContainer(up *skyobject.Unpack, container *Container) (registry.Dynamic, error) {
	raw := encoder.Serialize(container)
	hash, e := up.Add(raw)
	if e != nil {
		return registry.Dynamic{}, e
	}
	schema, e := up.Registry().SchemaByName("iko.Container")
	if e != nil {
		return registry.Dynamic{}, e
	}
	return registry.Dynamic{
		Hash:   hash,
		Schema: schema.Reference(),
	}, nil
}

func getPack(cxo *CXO, r *registry.Root) (*skyobject.Pack, error) {
	p, e := cxo.node.Container().Pack(r, reg)
	if e != nil {
		return nil, e
	}

	return p, nil
}

func getUnpack(cxo *CXO) (*skyobject.Unpack, error) {
	if cxo.c.MasterRooter == false {
		return nil, errors.New("not master")
	}
	return cxo.node.Container().Unpack(cxo.c.MasterRootSK, reg)
}
