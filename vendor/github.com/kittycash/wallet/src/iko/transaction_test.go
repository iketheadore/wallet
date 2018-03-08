package iko

import (
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
	"testing"
)

func runTransactionVerifyTest(t *testing.T, stateDB StateDB) {
	sk := cipher.SecKey([32]byte{
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
	})
	anAddress := cipher.AddressFromSecKey(sk)

	t.Run("TransactionCreated_InvalidPrevTransaction", func(t *testing.T) {
		// Add a kitty
		txHash := TxHash(cipher.SumSHA256([]byte{3, 4, 5, 6}))
		kID := KittyID(3)

		err := stateDB.AddKitty(txHash, kID, anAddress)

		// If there's an error creating kitty, then deviate testing transaction -- no kitty means no transaction
		if err == nil {
			prev := NewGenTx(nil, kID, sk)
			// Since there is no previous transactions then an error should be thrown that previous hash is invalid
			require.Errorf(t, prev.VerifyWith(prev), "There is no previous tansaction to verify")
		} else {
			fmt.Println("StateDB error, failed to create Kitty")
		}
	})

	kID := KittyID(3)

	// New secret key for toAddress
	sk2 := cipher.SecKey([32]byte{
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		5, 4, 3, 6,
		3, 4, 5, 6,
		4, 4, 5, 6,
		3, 4, 5, 6,
	})
	toAddress := cipher.AddressFromSecKey(sk2)
	prev := NewGenTx(nil, kID, sk)
	nextTrans := NewTransferTx(prev, kID, toAddress, sk)

	t.Run("TransactionCreated_InvalidDataMembers", func(t *testing.T) {
		// Change transaction previous hash to test if verify return error
		nextTrans.Prev = TxHash(cipher.SumSHA256([]byte{3, 4, 5, 6}))
		require.Errorf(t, nextTrans.VerifyWith(prev), "Previous hash was changed!!")

		// Revert transaction previous hash to its original state and change seqence number to test if verfiy returns error
		nextTrans.Prev = prev.Hash()
		nextTrans.Seq = prev.Seq + 5
		require.Errorf(t, nextTrans.VerifyWith(prev), "Previous Seq was changed and should be invalid!!")

		nextTrans.Seq = prev.Seq
		require.Errorf(t, nextTrans.VerifyWith(prev), "Previous Seq was changed and should be invalid!!")

		// Revert transaction sequence to its original state and change TS to test if Verify will return an error
		nextTrans.Seq = prev.Seq + 1

		// Temporary value for the original TS of transaction
		TS := nextTrans.TS

		// Set transaction's TS to invalid value
		nextTrans.TS = prev.TS - 10
		require.Errorf(t, nextTrans.VerifyWith(prev), "TS was changed and should be invalid")

		// Revert TS to its original value
		nextTrans.TS = TS
	})

	t.Run("Transaction_Audit_Verify_Success", func(t *testing.T) {
		require.Nil(t, nextTrans.VerifyWith(prev), "Verify should return nil for valid transactions")
	})
}

func runTransactionIsKittyGen(t *testing.T, stateDB StateDB) {
	sk := cipher.SecKey([32]byte{
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
	})

	sk2 := cipher.SecKey([32]byte{
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		5, 4, 3, 6,
		3, 4, 5, 6,
		4, 4, 5, 6,
		3, 4, 5, 6,
	})

	cAddress := cipher.AddressFromSecKey(sk)

	txHash := TxHash(cipher.SumSHA256([]byte{3, 7, 5, 6}))
	kID := KittyID(4)

	stateDB.AddKitty(txHash, kID, cAddress)

	prev := NewGenTx(nil, kID, sk)

	/**
	 * Testing if methods return false when tx.From != tx.To
	 * or pk is wrong
	 */
	t.Run("Transaction_AuditIsKittyGen_VerifyFalse", func(t *testing.T) {
		require.False(t, prev.IsKittyGen(cipher.PubKeyFromSecKey(sk2)), "Incorrect public key passed to method. Tx.IsKittyGen should return False")

		prev.From = cipher.AddressFromSecKey(sk2)
		require.False(t, prev.IsKittyGen(cipher.PubKeyFromSecKey(sk)), "Tx.From and Tx.to are not the same. Tx.IsKittyGen should return False")
	})

	t.Run("Transaction_TestIsKittyGen_Valid", func(t *testing.T) {
		prev.From = cipher.AddressFromSecKey(sk)
		require.True(t, prev.IsKittyGen(cipher.PubKeyFromSecKey(sk)), "Tx.From and Tx.To are the same. Tx.IsKittyGen should return True")
	})
}

func TestTransaction_Verify(t *testing.T) {
	stateDB := NewMemoryState()
	runTransactionVerifyTest(t, stateDB)
}

func TestTransaction_IsKittyGen(t *testing.T) {
	stateDB := NewMemoryState()
	runTransactionIsKittyGen(t, stateDB)
}
