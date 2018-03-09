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

		// Add transfer transactions.
		for i := len(txWraps); i < 5; i++ {
			tx, err := NewTransferTx(&txWraps[0].Tx, addr0, GenSK)
			require.NoError(t, err,
				"should generate transfer tx successfully")

			txWrap := TxWrapper{
				Tx: *tx,
				Meta: genTxMeta(uint64(i)),
			}

			txWraps = append(txWraps, txWrap)

			// Inject in master.
			err = master.AddTx(txWrap, func(tx *Transaction) error {
				return nil
			})
			require.NoError(t, err,
				"should successfully inject transfer tx in master")
			select {
			case recTxWrap := <-slave.TxChan():
				require.Equal(t, *recTxWrap, txWrap,
					"received tx is different from sent")

			case tm := <-time.After(time.Second * 5):
				require.Fail(t, "receive tx timed out", tm)
			}
		}

		// Manually check txs in slave.
		t.Run("SlaveReceivedTxsCheck", func(t *testing.T) {
			sTxWraps, err := slave.GetTxsOfSeqRange(0, uint64(len(txWraps)))

			require.NoError(t, err,
				"should successfully obtain txs of seq range")

			require.Equal(t, len(sTxWraps), len(txWraps),
				"should obtain same number of txs as that injected")

			for i, txWrap := range sTxWraps {
				t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
					require.Equal(t, txWrap, txWraps[i],
						"injected and received txs should be the same")
				})
			}
		})
	})

	// TODO (evanlinjin): Write these tests >>>
	//		- Run master/slave nodes, writing to disks.
	//			- Close slave, inject txs, re-open slave -> test sync.
	//			- Close master, reopen master, inject txs -> test sync.
	//			- Close discovery, reopen discovery, inject txs -> test sync.
}
