package cxo

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/skycoin/cxo/node"
	"github.com/skycoin/net/skycoin-messenger/factory"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"

	"github.com/kittycash/kittiverse/src/kitty"

	"github.com/kittycash/wallet/src/iko/transaction"
)

const (
	MasterRootNonce = 12345
	MasterRootSeed  = "root seed"
	MasterGenSeed   = "gen seed"
)

var (
	RootPK, RootSK = cipher.GenerateDeterministicKeyPair([]byte(MasterRootSeed))
	GenPK, GenSK   = cipher.GenerateDeterministicKeyPair([]byte(MasterGenSeed))
)

func newCXOChainDB(
	dir string,
	master bool,
	doInit bool,
	addr string,
	dAddrs []string,
	logPrefix ...string,
) (*CXO, error) {
	chainDB, err := New(
		&Config{
			Dir:                dir,
			Public:             true,
			Memory:             dir == "",
			MessengerAddresses: dAddrs,
			CXOAddress:         addr,
			MasterRooter:       master,
			MasterRootPK:       RootPK,
			MasterRootSK:       RootSK,
			MasterRootNonce:    MasterRootNonce,
		},
		func(nc *node.Config) error {
			if len(logPrefix) > 0 {
				nc.Logger.Prefix = logPrefix[0] + " "
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	if doInit {
		if err := chainDB.MasterInitChain(); err != nil {
			return nil, err
		}
	}
	return chainDB, chainDB.RunTxService(func(tx *transaction.Transaction) error {
		return nil
	})
}

func newDiscoveryServer(addr string) (func(), error) {
	f := factory.NewMessengerFactory()
	return func() {
		f.Close()
	}, f.Listen(addr)
}

func genTxWraps(count, start int) []transaction.Wrapper {
	var (
		out = make([]transaction.Wrapper, count)
	)
	for i := range out {
		out[i] = transaction.Wrapper{
			Tx:   *transaction.NewGenTx(kitty.ID(i+start), GenSK),
			Meta: genTxMeta(uint64(i + start)),
		}
	}
	return out
}

func genTxMeta(seq uint64) transaction.Meta {
	return transaction.Meta{
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

		var (
			txWraps = genTxWraps(3, 0)
		)

		for i, txWrap := range txWraps {
			t.Run(fmt.Sprintf("ReceiveTx_%d", i), func(t *testing.T) {
				err := master.AddTx(txWrap, func(tx *transaction.Transaction) error {
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
			tx, err := transaction.NewTransferTx(&txWraps[0].Tx, addr0, GenSK)
			require.NoError(t, err,
				"should generate transfer tx successfully")

			txWrap := transaction.Wrapper{
				Tx:   *tx,
				Meta: genTxMeta(uint64(i)),
			}

			txWraps = append(txWraps, txWrap)

			// Inject in master.
			err = master.AddTx(txWrap, func(tx *transaction.Transaction) error {
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

	t.Run("Master_RecoverOnRestart", func(t *testing.T) {

		// Create temporary directory.
		temp, err := ioutil.TempDir("", "kc_chain_cxo_test_Master_RecoverOnRestart")
		require.NoError(t, err, "creation of temp dir should succeed")
		defer os.RemoveAll(temp)

		var (
			txWraps = genTxWraps(10, 0)
		)

		t.Run("Master_InjectTxsAndClose", func(t *testing.T) {
			master, err := newCXOChainDB(
				temp, true, true, MasterAddr, []string{DiscoveryAddr})
			require.NoError(t, err,
				"creation of master should succeed")
			defer master.Close()

			for _, txWrap := range txWraps {
				err := master.AddTx(txWrap, func(_ *transaction.Transaction) error {
					return nil
				})
				require.NoError(t, err, "inject tx should succeed")
			}
		})

		t.Run("Master_ReopenAndCheckTxs", func(t *testing.T) {
			master, err := newCXOChainDB(
				temp, true, false, MasterAddr, []string{DiscoveryAddr})
			require.NoError(t, err,
				"creation of master should succeed")
			defer master.Close()

			// Loop txs and check.
			for i, txWrap := range txWraps {
				gotTxWrap, err := master.GetTxOfSeq(uint64(i))
				require.NoError(t, err,
					"should successfully obtain tx")
				require.Equal(t, gotTxWrap, txWrap,
					"obtained tx should be the same as injected tx")
			}
		})
	})

	// TODO (evanlinjin): Write these tests >>>
	//		- Run master/slave nodes, writing to disks.
	//			- Close slave, inject txs, re-open slave -> test sync.
	//			- Close master, reopen master, inject txs -> test sync.
	//			- Close discovery, reopen discovery, inject txs -> test sync.

	t.Run("MasterSlave_ReconnectAfterRestart", func(t *testing.T) {

		// Create temporary directories.

		masterDir, err := ioutil.TempDir("", "kc_test_masterCXODir")
		require.NoError(t, err)
		//defer os.RemoveAll(masterDir)

		slaveDir, err := ioutil.TempDir("", "kc_test_slaveCXODir")
		require.NoError(t, err)
		//defer os.RemoveAll(slaveDir)

		// Start master and initiate chain.

		master, err := newCXOChainDB(
			masterDir, true, true, MasterAddr, []string{DiscoveryAddr}, "MASTER")
		require.NoError(t, err)

		// Inject some txs.

		var (
			count, start = 5, 0
			txWraps      = genTxWraps(count, start)
		)
		for _, txWrap := range txWraps {
			require.NoError(t, master.AddTx(txWrap, func(_ *transaction.Transaction) error {
				return nil
			}))
		}

		// Start slave.

		slave, err := newCXOChainDB(
			slaveDir, false, false, SlaveAddr, []string{DiscoveryAddr}, "SLAVE")
		require.NoError(t, err)

		// Inject more txs, waiting for slave to receive.

		t.Run("InjectTxs", func(t *testing.T) {

			start += count
			txWraps = append(txWraps, genTxWraps(count, start)...)

			for i := start; i < start+count; i++ {
				// Inject.
				require.NoError(t, master.AddTx(txWraps[i], func(_ *transaction.Transaction) error {
					return nil
				}))
				// Wait for slave to receive.
				select {
				case _, ok := <-slave.TxChan():
					require.True(t, ok)
				case <-time.After(time.Second * 2):
					require.Fail(t, "slave tx receive timed out")
				}
			}
		})

		// Restart master.

		master.Close()

		master, err = newCXOChainDB(
			masterDir, true, false, MasterAddr, []string{DiscoveryAddr}, "MASTER")
		require.NoError(t, err)

		// Inject more txs, waiting for slave to receive.

		t.Run("InjectTxsAfterMasterRestart", func(t *testing.T) {

			start += count
			txWraps = append(txWraps, genTxWraps(count, start)...)

			for i := start; i < start+count; i++ {
				// Inject.
				require.NoError(t, master.AddTx(txWraps[i], func(_ *transaction.Transaction) error {
					return nil
				}))
				// Wait.
				select {
				case _, ok := <-slave.TxChan():
					require.True(t, ok)
				case <-time.After(time.Second * 2):
					require.Fail(t, "slave tx receive timed out")
				}
			}

			// Check txs stored in master and slave nodes.
			for i, txWrap := range txWraps {

				masterTxWrap, err := master.GetTxOfSeq(uint64(i))
				require.NoError(t, err)
				require.Equal(t, masterTxWrap, txWrap)

				slaveTxWrap, err := slave.GetTxOfSeq(uint64(i))
				require.NoError(t, err)
				require.Equal(t, slaveTxWrap, txWrap)
			}
		})

		// Restart master.

		master.Close()

		master, err = newCXOChainDB(
			masterDir, true, false, MasterAddr, []string{DiscoveryAddr}, "MASTER")
		require.NoError(t, err)

		// Check again.

		// Check txs stored in master and slave nodes.
		for i, txWrap := range txWraps {

			masterTxWrap, err := master.GetTxOfSeq(uint64(i))
			require.NoError(t, err)
			require.Equal(t, masterTxWrap, txWrap)

			slaveTxWrap, err := slave.GetTxOfSeq(uint64(i))
			require.NoError(t, err)
			require.Equal(t, slaveTxWrap, txWrap)
		}

		// Restart slave.

		slave.Close()

		slave, err = newCXOChainDB(
			slaveDir, false, false, SlaveAddr, []string{DiscoveryAddr}, "SLAVE")
		require.NoError(t, err)

		// Check again.

		// Check txs stored in master and slave nodes.
		for i, txWrap := range txWraps {

			masterTxWrap, err := master.GetTxOfSeq(uint64(i))
			require.NoError(t, err)
			require.Equal(t, masterTxWrap, txWrap)

			slaveTxWrap, err := slave.GetTxOfSeq(uint64(i))
			require.NoError(t, err)
			require.Equal(t, slaveTxWrap, txWrap)
		}

		master.Close()
		slave.Close()
	})
}
