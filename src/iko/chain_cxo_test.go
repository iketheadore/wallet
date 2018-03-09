package iko

import (
	"fmt"
	"github.com/skycoin/net/skycoin-messenger/factory"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func newCXOChainDB(
	dir string,
	master bool,
	doInit bool,
	addr string,
	dAddrs []string,
) (*CXOChain, error) {
	chainDB, err := NewCXOChain(&CXOChainConfig{
		Public:             true,
		Memory:             dir == "",
		MessengerAddresses: dAddrs,
		CXOAddress:         addr,
		MasterRooter:       master,
		MasterRootPK:       RootPK,
		MasterRootSK:       RootSK,
		MasterRootNonce:    MasterRootNonce,
	})
	if err != nil {
		return nil, err
	}
	if doInit {
		if err := chainDB.MasterInitChain(); err != nil {
			return nil, err
		}
	}
	return chainDB, chainDB.RunTxService(func(tx *Transaction) error {
		return nil
	})
}

func newDiscoveryServer(addr string) (func(), error) {
	f := factory.NewMessengerFactory()
	return func() {
		f.Close()
	}, f.Listen(addr)
}

func genTxMeta(seq uint64) TxMeta {
	return TxMeta{
		Seq: seq,
		TS:  time.Now().UnixNano(),
	}
}

/*
	<<< TESTS BEGIN >>>
*/

func TestNewCXOChain(t *testing.T) {

	const (
		DiscoveryAddr = ":5412"
		MasterAddr    = ":7314"
		SlaveAddr     = ":6013"
	)

	// Start discovery node.
	fClose, err := newDiscoveryServer(DiscoveryAddr)
	require.NoError(t, err,
		"discovery server should start")
	defer fClose()

	time.Sleep(time.Second * 2)

	t.Run("MasterSlave_ReceiveOnMemory", func(t *testing.T) {

		master, err := newCXOChainDB(
			"", true, true, MasterAddr, []string{DiscoveryAddr})
		require.NoError(t, err,
			"creation of master should succeed")
		defer master.Close()

		time.Sleep(time.Second * 2)

		slave, err := newCXOChainDB(
			"", false, false, SlaveAddr, []string{DiscoveryAddr})
		require.NoError(t, err,
			"creation of slave should succeed")
		defer slave.Close()

		txWraps := []TxWrapper{
			{
				Tx:   *NewGenTx(KittyID(0), GenSK),
				Meta: genTxMeta(0),
			},
			{
				Tx:   *NewGenTx(KittyID(1), GenSK),
				Meta: genTxMeta(1),
			},
			{
				Tx:   *NewGenTx(KittyID(2), GenSK),
				Meta: genTxMeta(2),
			},
		}

		for i, txWrap := range txWraps {
			t.Run(fmt.Sprintf("ReceiveTx_%d", i), func(t *testing.T) {
				err := master.AddTx(txWrap, func(tx *Transaction) error {
					return nil
				})
				require.NoErrorf(t, err,
					"failed to add tx %d:'%s'", i, txWrap.Tx.String())

				select {
				case recTxWrap := <-slave.TxChan():
					require.Equal(t, *recTxWrap, txWrap,
						"received tx is different from sent")

				case tm := <-time.After(time.Second * 5):
					require.Fail(t, "receive tx timed out", tm)
				}
			})
		}

		var (
			_, sk0 = cipher.GenerateDeterministicKeyPair([]byte("user 0"))
			addr0  = cipher.AddressFromSecKey(sk0)
		)

		// Add more transactions.
		for i := len(txWraps); i < 5; i++ {
			NewTransferTx(&txWraps[0].Tx, addr0, GenSK)
		}

		//txWraps := append(txWraps, []TxWrapper{
		//	{
		//		Tx:
		//	},
		//}...)
	})
}
