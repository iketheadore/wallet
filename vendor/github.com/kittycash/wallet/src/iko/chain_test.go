package iko

import (
	"errors"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
	"testing"
)

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
					"The last page should have%d items", finalPageCount)
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
		kittyID := KittyID(5)

		firstSecKey := cipher.SecKey([32]byte{
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
		})
		firstOwnerAddress := cipher.AddressFromSecKey(firstSecKey)

		secondSecKey := cipher.SecKey(
			cipher.SecKey([32]byte{
				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,

				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,
			}))

		secondOwnerAddress := cipher.AddressFromSecKey(secondSecKey)

		firstTransaction := NewGenTx(nil, kittyID, firstSecKey)

		t.Run("AddTx_Failure", func(t *testing.T) {
			err := chainDB.AddTx(*firstTransaction, addTxAlwaysReject)
			require.NotNil(t, err, "This shouldn't succeed")
		})

		err := chainDB.AddTx(*firstTransaction, addTxAlwaysApprove)

		require.Nil(t, err,
			"We should be able to successfully add our first transaction")

		t.Run("Head_Success_01", func(t *testing.T) {
			transaction, err := chainDB.Head()

			require.Nil(t, err, "Should not give us an error")
			require.Equal(t, transaction, *firstTransaction,
				"Should correctly return the first transaction")
		})

		secondTransaction := NewTransferTx(
			firstTransaction, kittyID, secondOwnerAddress, firstSecKey)

		err = chainDB.AddTx(*secondTransaction, addTxAlwaysApprove)

		require.Nil(t, err,
			"We should be able to successfully add our second transaction")

		transactions := []Transaction{*firstTransaction, *secondTransaction}

		t.Run("Head_Success_02", func(t *testing.T) {
			transaction, err := chainDB.Head()

			require.Nil(t, err, "Should not give us an error")
			require.Equal(t, transaction, *secondTransaction,
				"Should correctly return the second transaction")
		})

		t.Run("Len", func(t *testing.T) {
			require.Equal(t, chainDB.Len(), uint64(2),
				"We should have two transactions by now")
		})

		t.Run("GetTxOfHash_NonexistentHash_02", func(t *testing.T) {
			_, err := chainDB.GetTxOfHash(nonexistentHash)

			require.NotNil(t, err,
				"Should still give us an error because there are no transactions by that hash")
		})

		for idx, transaction := range transactions {
			// our test label is GetTxOfHash_Success_XX, where 01 is
			// firstTransaction, 02 is secondTransaction, etc
			testLabel := fmt.Sprintf("GetTxOfHash_Success_%2.2d", idx+1)

			t.Run(testLabel, func(t *testing.T) {
				reqTransaction, err := chainDB.GetTxOfHash(transaction.Hash())

				require.Nil(t, err, "Shouldn't return an error for a valid hash")
				require.Equal(t, transaction, reqTransaction,
					"Should correctly return the right transaction")
			})
		}

		for idx, transaction := range transactions {
			// same as above
			testLabel := fmt.Sprintf("GetTxOfSeq_Success_%2.2d", idx+1)

			t.Run(testLabel, func(t *testing.T) {
				reqTransaction, err := chainDB.GetTxOfSeq(transaction.Seq)

				require.Nil(t, err,
					"Shouldn't return an error for a valid sequence index")
				require.Equal(t, transaction, reqTransaction,
					"Should correctly return the right transaction")
			})
		}

		t.Run("HeadSeq", func(t *testing.T) {
			require.Equal(t, chainDB.Len()-1, chainDB.HeadSeq(),
				"HeadSeq() should be Len() - 1")
		})

		t.Run("TxChan", func(t *testing.T) {
			txChan := chainDB.TxChan()

			require.NotNil(t, txChan,
				"Should return a valid receiving channel for transactions")

			for _, transaction := range transactions {
				channelTransaction := <-txChan
				require.Equal(t, *channelTransaction, transaction,
					"Should return our transactions through the TxChan() channel in order they were added")
			}
		})

		// adding a third transaction for an odd number of transactions
		thirdTransaction := NewTransferTx(
			secondTransaction, kittyID, firstOwnerAddress, secondSecKey)

		err = chainDB.AddTx(*thirdTransaction, addTxAlwaysApprove)

		require.Nil(t, err, "We should be able to successfully transfer the kitty back to the original owner")

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

func TestChainDB_MemoryChain(t *testing.T) {
	chainDB := NewMemoryChain(0)

	require.NotNil(t, chainDB, "We should be able to create an empty MemoryChain")

	runChainDBTest(t, chainDB)
}
