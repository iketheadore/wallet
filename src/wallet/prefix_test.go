package wallet

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
)

func TestNewPrefix(t *testing.T) {
	cases := []struct {
		Ver   uint64
		Nonce []byte
		Enc   bool
	}{
		{0, EmptyNonce(), false},
		{1, EmptyNonce(), false},
		{2, RandNonce(), true},
		{3, RandNonce(), true},
	}
	for _, c := range cases {
		p := NewPrefix(c.Ver, c.Nonce)
		require.Equal(t, c.Ver, p.Version())
		require.Equal(t, c.Nonce, p.Nonce())
		require.Equal(t, c.Enc, p.Encrypted())
	}
}

func TestExtractPrefix(t *testing.T) {
	cases := []struct {
		Prefix Prefix
		Data   []byte
	}{
		{NewPrefix(0, EmptyNonce()), cipher.RandByte(1)},
		{NewPrefix(1, EmptyNonce()), cipher.RandByte(100)},
		{NewPrefix(20, RandNonce()), cipher.RandByte(2000)},
		{NewPrefix(0, RandNonce()), cipher.RandByte(45000)},
	}
	t.Run("in_memory", func(t *testing.T) {
		for _, c := range cases {
			raw := append(c.Prefix[:], c.Data...)
			prefix, data, err := ExtractPrefix(raw)
			require.NoError(t, err)
			require.Equal(t, c.Prefix, prefix)
			require.Equal(t, c.Data, data)
		}
	})
	t.Run("on_disk", func(t *testing.T) {
		for i, c := range cases {
			var (
				fName = fmt.Sprintf("kcWalletTest_%d%s", i, FileExt)
				fPath = filepath.Join(os.TempDir(), fName)
				raw   = append(c.Prefix[:], c.Data...)
			)
			require.NoError(t, SaveBinary(fPath, raw))

			func() {
				f, err := os.Open(fPath)
				require.NoError(t, err)
				defer f.Close()

				raw, err = ioutil.ReadAll(f)
				require.NoError(t, err)

				prefix, data, err := ExtractPrefix(raw)
				require.NoError(t, err)
				require.Equal(t, c.Prefix, prefix)
				require.Equal(t, c.Data, data)
			}()
		}
	})
}
