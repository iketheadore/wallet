package transaction

import (
	"testing"

	"github.com/kittycash/kittiverse/src/kitty"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
)

func runTransactionVerifyTest(t *testing.T) {
	var (
		pk0, sk0 = cipher.GenerateDeterministicKeyPair([]byte("seed 0"))
		_        = cipher.AddressFromPubKey(pk0)
	)

	t.Run("TransactionCreated_InvalidPrevTransaction", func(t *testing.T) {
		prev := NewGenTx(kitty.ID(3), sk0)
		require.Errorf(t, prev.VerifyWith(prev, pk0),
			"should fail")
	})

	var (
		kID    = kitty.ID(3)
		pk1, _ = cipher.GenerateDeterministicKeyPair([]byte("seed 1"))
		ad1    = cipher.AddressFromPubKey(pk1)
	)

	prev := NewGenTx(kID, sk0)
	nextTrans, e := NewTransferTx(prev, ad1, sk0)
	require.NoError(t, e,
		"should succeed")

	t.Run("TransactionCreated_InvalidDataMembers", func(t *testing.T) {
		// Change transaction previous hash to test if verify return error
		nextTrans.In = ID(cipher.SumSHA256([]byte{3, 4, 5, 6}))
		require.Errorf(t, nextTrans.VerifyWith(prev, pk0),
			"input tx hash was changed!!")

		// Revert transaction previous hash to its original state and change seqence number to test if verfiy returns error
		nextTrans.In = prev.Hash()
	})

	t.Run("Transaction_Audit_Verify_Success", func(t *testing.T) {
		require.Nil(t, nextTrans.VerifyWith(prev, pk0),
			"Verify should return nil for valid transactions")
	})
}

func runTransactionIsKittyGen(t *testing.T) {
	var (
		_, sk0 = cipher.GenerateDeterministicKeyPair([]byte("seed 0"))
		genTx  = NewGenTx(kitty.ID(4), sk0)
	)

	t.Run("Transaction_AuditIsKittyGen_VerifyFalse", func(t *testing.T) {
		genTx.In = ID(cipher.SumSHA256([]byte{3, 7, 5, 6}))
		require.False(t,
			genTx.IsKittyGen(cipher.PubKeyFromSecKey(sk0)),
			"Incorrect input tx hash. Tx.IsKittyGen should return False")
	})

	t.Run("Transaction_TestIsKittyGen_Valid", func(t *testing.T) {
		genTx.In = EmptyID()
		require.True(t,
			genTx.IsKittyGen(cipher.PubKeyFromSecKey(sk0)),
			"Tx.From and Tx.To are the same. Tx.IsKittyGen should return True")
	})
}

func TestTransaction_Verify(t *testing.T) {
	runTransactionVerifyTest(t)
}

func TestTransaction_IsKittyGen(t *testing.T) {
	runTransactionIsKittyGen(t)
}
