package iko

import (
	"errors"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"gopkg.in/sirupsen/logrus.v1"
	"os"
	"sync"
	"time"
)

type BlockChainConfig struct {
	GenerationPK cipher.PubKey
	TxAction     TxAction
}

func (cc *BlockChainConfig) Prepare() error {
	if cc.TxAction == nil {
		cc.TxAction = func(tx *Transaction) error {
			return nil
		}
	}
	if e := cc.GenerationPK.Verify(); e != nil {
		return e
	}
	return nil
}

type BlockChain struct {
	c     *BlockChainConfig
	chain ChainDB
	state StateDB
	log   *logrus.Logger
	mux   sync.RWMutex

	wg   sync.WaitGroup
	quit chan struct{}
}

func NewBlockChain(config *BlockChainConfig, chainDB ChainDB, stateDB StateDB) (*BlockChain, error) {
	if e := config.Prepare(); e != nil {
		return nil, e
	}
	bc := &BlockChain{
		c:     config,
		chain: chainDB,
		state: stateDB,
		log: &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
		},
		quit: make(chan struct{}),
	}

	if e := bc.InitState(); e != nil {
		return nil, e
	}

	bc.wg.Add(1)
	go bc.service()

	return bc, nil
}

func (bc *BlockChain) InitState() error {
	var check = MakeTxChecker(bc)
	for i := uint64(1); i < bc.chain.Len(); i++ {

		// Val transaction.
		txWrap, e := bc.chain.GetTxOfSeq(i)
		if e != nil {
			return e
		}
		bc.log.
			WithField("tx", txWrap.Tx.String()).
			WithField("meta", txWrap.Meta).
			Infof("InitState (%d)", i)

		if e := check(&txWrap.Tx); e != nil {
			return e
		}
	}
	return nil
}

func (bc *BlockChain) Close() {
	bc.log.Info("closing blockchain manager")
	close(bc.quit)
}

func (bc *BlockChain) service() {
	defer bc.wg.Done()

	for {
		select {
		case <-bc.quit:
			return

		case txWrap, ok := <-bc.chain.TxChan():
			if !ok {
				return
			}
			if e := bc.c.TxAction(&txWrap.Tx); e != nil {
				panic(e)
			}
		}
	}
}

func (bc *BlockChain) GetHeadTx() (TxWrapper, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain.Head()
}

func (bc *BlockChain) GetTxOfHash(txHash TxHash) (TxWrapper, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain.GetTxOfHash(txHash)
}

func (bc *BlockChain) GetTxOfSeq(seq uint64) (TxWrapper, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain.GetTxOfSeq(seq)
}

func (bc *BlockChain) GetKittyState(kittyID KittyID) (*KittyState, bool) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.state.GetKittyState(kittyID)
}

func (bc *BlockChain) GetAddressState(address cipher.Address) *AddressState {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.state.GetAddressState(address)
}

func (bc *BlockChain) InjectTx(tx *Transaction) (*TxMeta, error) {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	var seq uint64
	if txWrap, e := bc.chain.Head(); e == nil {
		seq = txWrap.Meta.Seq + 1
	}

	meta := TxMeta{
		Seq: seq,
		TS:  time.Now().UnixNano(),
	}

	return &meta, bc.chain.AddTx(
		TxWrapper{
			Tx:   *tx,
			Meta: meta,
		},
		MakeTxChecker(bc),
	)
}

func MakeTxChecker(bc *BlockChain) TxChecker {
	return func(tx *Transaction) error {

		// Check duplicate.
		if _, e := bc.chain.GetTxOfHash(tx.Hash()); e == nil {
			return fmt.Errorf("tx of hash '%s' is a duplicate",
				tx.Hash())
		}

		var unspent *Transaction
		if tempHash, ok := bc.state.GetKittyUnspentTx(tx.KittyID); ok {
			temp, e := bc.chain.GetTxOfHash(tempHash)
			if e != nil {
				return e
			}
			unspent = &temp.Tx
		}

		if e := tx.VerifyWith(unspent, bc.c.GenerationPK); e != nil {
			return e
		}
		if tx.IsKittyGen(bc.c.GenerationPK) {
			bc.log.
				WithField("kitty_id", tx.KittyID).
				WithField("input", tx.In.Hex()).
				WithField("output", tx.Out.String()).
				Debug("processing generation tx")

			if e := bc.state.AddKitty(tx.Hash(), tx.KittyID, tx.Out); e != nil {
				return e
			}
		} else {
			bc.log.
				WithField("kitty_id", tx.KittyID).
				WithField("input", tx.In.Hex()).
				WithField("output", tx.Out.String()).
				Debug("processing transfer tx")

			// TEMPORARY: If tx is not signed from generation pk, disallow.
			if unspent.Out != cipher.AddressFromPubKey(bc.c.GenerationPK) {
				return errors.New("tx rejected")
			}

			if e := bc.state.MoveKitty(tx.Hash(), tx.KittyID, unspent.Out, tx.Out); e != nil {
				return e
			}
		}
		return nil
	}
}

type PaginatedTransactions struct {
	TotalPageCount uint64
	Transactions   []TxWrapper
}

// totalPageCount is a helper function for calculating the number of pages given
// the number of transactions and the number of transactions per page
func totalPageCount(len, pageSize uint64) uint64 {
	if len%pageSize == 0 {
		return len / pageSize
	} else {
		return (len / pageSize) + 1
	}
}

func (bc *BlockChain) GetTransactionPage(currentPage, perPage uint64) (PaginatedTransactions, error) {
	txWrappers, err := bc.chain.GetTxsOfSeqRange(
		uint64(perPage*currentPage),
		perPage)
	if err != nil {
		return PaginatedTransactions{}, err
	}
	cLen := bc.chain.Len()
	return PaginatedTransactions{
		TotalPageCount: totalPageCount(cLen, perPage),
		Transactions:   txWrappers,
	}, nil
}
