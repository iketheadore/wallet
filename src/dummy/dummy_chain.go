package dummy

import (
	"github.com/kittycash/wallet/src/iko/transaction"
	"errors"
	"github.com/kittycash/wallet/src/iko"
)

var (
	ErrNoBlockChain = errors.New("no block chain")
)

func (_ *Dummy) Head() (transaction.Wrapper, error) {
	return transaction.Wrapper{}, ErrNoBlockChain
}

func (_ *Dummy) Len() uint64 {
	return 0
}

func (_ *Dummy) AddTx(txWrapper transaction.Wrapper, check iko.TxChecker) error {
	if err := check(&txWrapper.Tx); err != nil {
		return err
	}
	return ErrNoBlockChain
}

func (_ *Dummy) GetTxOfHash(_ transaction.ID) (transaction.Wrapper, error) {
	return transaction.Wrapper{}, ErrNoBlockChain
}

func (_ *Dummy) GetTxOfSeq(_ uint64) (transaction.Wrapper, error) {
	return transaction.Wrapper{}, ErrNoBlockChain
}

func (_ *Dummy) TxChan() <-chan *transaction.Wrapper {
	return make(<-chan *transaction.Wrapper)
}

func (_ *Dummy) GetTxsOfSeqRange(_ uint64, _ uint64) ([]transaction.Wrapper, error) {
	return nil, ErrNoBlockChain
}