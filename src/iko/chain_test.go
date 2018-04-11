package iko

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/skycoin/cxo/node"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"

	"github.com/kittycash/wallet/src/cxo"
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

/*
	<<< TESTS BEGIN >>>
*/

func testChainDBPagination(t *testing.T, chainDB ChainDB, pageSize uint64) {
	t.Run("testChainDBPagination", func(t *testing.T) {
		// demonstrating the pagination flow
		currentPage := uint64(0)
		currentSeq := uint64(0)

		for {
			// we include this in the loop because in a real life application, the
			// page count likely will change regularly, so each update will need to
			// recalculate the maximum page count
			pageCount := totalPageCount(chainDB.Len(), pageSize)
			finalPageCount := chainDB.Len() % pageSize

			require.True(t, currentPage <= pageCount,
				"The current page should never get beyond the total page count")

			if currentPage == pageCount {
				return // went through all the pages, returning
			}

			transactions, err := chainDB.GetTxsOfSeqRange(currentSeq, pageSize)

			require.Nil(t, err, "Shouldn't have an error")
			require.NotNil(t, transactions, "Should receive some transactions")

			if currentPage == (pageCount - 1) {
				require.Lenf(t, transactions, int(finalPageCount),
					"The last page should have %d items", finalPageCount)
			} else {
				require.Lenf(t, transactions, int(pageSize),
					"A normal page should have %d items", pageSize)
			}

			// and now we increment our currentSeq and currentPage
			currentSeq = currentSeq + pageSize
			currentPage = currentPage + 1
		}
	})
}

func addTxAlwaysApprove(tx *Transaction) error {
	return nil
}

func addTxAlwaysReject(tx *Transaction) error {
	return errors.New("failure")
}

func runChainDBTest(t *testing.T, chainDB ChainDB) {
	t.Run("Head_NoTransactions", func(t *testing.T) {
		_, err := chainDB.Head()

		require.NotNil(t, err,
			"Should give us an error because there are no transactions yet")
	})

	nonexistentHash := TxHash(cipher.SumSHA256([]byte{3, 4, 5, 6}))

	t.Run("GetTxOfHash_NonexistentHash_01", func(t *testing.T) {
		_, err := chainDB.GetTxOfHash(nonexistentHash)

		require.NotNil(t, err,
			"Should give us an error because there are no transactions yet")
	})

	t.Run("GetTxOfSeq_NonexistentSeq", func(t *testing.T) {
		_, err := chainDB.GetTxOfSeq(0)

		require.NotNil(t, err,
			"Should give us an error because there are no transactions yet")
	})

	t.Run("withTransactions", func(t *testing.T) {
		var (
			kittyID = KittyID(5)
			_, sk2  = cipher.GenerateDeterministicKeyPair([]byte("2nd tx seed"))
			addr2   = cipher.AddressFromSecKey(sk2)
		)

		firstTxWrap := TxWrapper{
			Tx: *NewGenTx(kittyID, GenSK),
			Meta: TxMeta{
				Seq: 0,
				TS:  time.Now().UnixNano(),
			},
		}

		t.Run("AddTx_Failure", func(t *testing.T) {
			err := chainDB.AddTx(firstTxWrap, addTxAlwaysReject)
			require.Error(t, err, "This shouldn't succeed")
		})

		err := chainDB.AddTx(firstTxWrap, addTxAlwaysApprove)
		require.NoError(t, err,
			"We should be able to successfully add our first transaction")

		t.Run("Head_Success_01", func(t *testing.T) {
			txWrap, err := chainDB.Head()

			require.NoError(t, err, "Should not give us an error")
			require.Equal(t, txWrap, firstTxWrap,
				"Should correctly return the first transaction")
		})

		secondTx, err := NewTransferTx(
			&firstTxWrap.Tx, addr2, GenSK)
		require.NoError(t, err, "should create second tx with no error")

		var secondTxWrap = TxWrapper{
			Tx: *secondTx,
			Meta: TxMeta{
				Seq: 1,
				TS:  time.Now().UnixNano(),
			},
		}

		err = chainDB.AddTx(secondTxWrap, addTxAlwaysApprove)
		require.NoError(t, err,
			"We should be able to successfully add our second transaction")

		txWraps := []TxWrapper{firstTxWrap, secondTxWrap}

		t.Run("Head_Success_02", func(t *testing.T) {
			txWrap, err := chainDB.Head()
			require.NoError(t, err,
				"Should not give us an error")
			require.Equal(t, txWrap, secondTxWrap,
				"Should correctly return the second transaction")
		})

		t.Run("Len", func(t *testing.T) {
			require.Equal(t, chainDB.Len(), uint64(2),
				"We should have two transactions by now")
		})

		t.Run("GetTxOfHash_NonexistentHash_02", func(t *testing.T) {
			_, err := chainDB.GetTxOfHash(nonexistentHash)

			require.Error(t, err,
				"Should still give us an error because there are no transactions by that hash")
		})

		for idx, txWrap := range txWraps {
			// our test label is GetTxOfHash_Success_XX, where 01 is
			// firstTransaction, 02 is secondTransaction, etc
			testLabel := fmt.Sprintf("GetTxOfHash_Success_%2.2d", idx+1)

			t.Run(testLabel, func(t *testing.T) {
				reqTxWrap, err := chainDB.GetTxOfHash(txWrap.Tx.Hash())

				require.NoError(t, err,
					"Shouldn't return an error for a valid hash")
				require.Equal(t, txWrap, reqTxWrap,
					"Should correctly return the right transaction")
			})
		}

		for idx, txWrap := range txWraps {
			// same as above
			testLabel := fmt.Sprintf("GetTxOfSeq_Success_%2.2d", idx+1)

			t.Run(testLabel, func(t *testing.T) {
				reqTxWrap, err := chainDB.GetTxOfSeq(txWrap.Meta.Seq)

				require.NoError(t, err,
					"Shouldn't return an error for a valid sequence index")
				require.Equal(t, txWrap, reqTxWrap,
					"Should correctly return the right transaction")
			})
		}

		// adding a third transaction for an odd number of transactions
		thirdTx, err := NewTransferTx(
			secondTx, cipher.AddressFromPubKey(GenPK), sk2)

		require.NoError(t, err, "should create second tx with no error")

		var thirdTxWrap = TxWrapper{
			Tx: *thirdTx,
			Meta: TxMeta{
				Seq: 2,
				TS:  time.Now().UnixNano(),
			},
		}

		err = chainDB.AddTx(thirdTxWrap, addTxAlwaysApprove)
		require.NoError(t, err,
			"We should be able to successfully transfer the kitty back to the original owner")

		t.Run("GetTxsOfSeqRange_BadPageSize", func(t *testing.T) {
			transactions, err := chainDB.GetTxsOfSeqRange(0, 0)

			require.Nil(t, transactions,
				"We shouldn't return anything because the caller passed a bad page size")
			require.NotNil(t, err, "We should get an error for a bad page size")
		})

		t.Run("GetTxsOfSeqRange_BadStartSeq", func(t *testing.T) {
			transactions, err := chainDB.GetTxsOfSeqRange(5, 2)

			require.Nil(t, transactions,
				"We shouldn't return anything because the caller passed a bad start sequence index")
			require.NotNil(t, err,
				"We should get an error for a bad start sequence index")
		})

		testChainDBPagination(t, chainDB, 2)
	})
}

func TestChainDB_CXOChain(t *testing.T) {

	chainDB, err := newCXOChainDB(
		"", true, true, ":7999", nil)

	require.NoError(t, err,
		"master root should init with no problem")

	runChainDBTest(t, chainDB)
}

func newCXOChainDB(
	dir string,
	master bool,
	doInit bool,
	addr string,
	dAddrs []string,
	logPrefix ...string,
) (*cxo.CXO, error) {
	chainDB, err := cxo.New(
		&cxo.Config{
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
	return chainDB, chainDB.RunTxService(func(tx *Transaction) error {
		return nil
	})
}
