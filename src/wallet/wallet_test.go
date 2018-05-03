package wallet

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func initTempDir(t *testing.T) func() {
	dir, err := ioutil.TempDir("", "kittycash_test")
	require.Empty(t, err)

	require.Empty(t, SetRootDir(dir))

	return func() {
		os.RemoveAll(dir)
	}
}

func saveWallet(t *testing.T, options *Options) {
	fWallet, err := NewFloatingWallet(options)
	require.Empty(t, err, "failed to create floating wallet")

	require.Empty(t, fWallet.Save(), "failed to save wallet")
}

func loadWallet(t *testing.T, label, pw string) *Wallet {
	f, err := os.Open(LabelPath(label))
	require.Nilf(t, err, "failed to open wallet of label '%s'", label)
	defer f.Close()

	fw, err := LoadFloatingWallet(f, label, pw)
	require.Empty(t, err, "failed to load floating wallet")

	return fw
}

func TestFloatingWallet_Save(t *testing.T) {
	rmTemp := initTempDir(t)
	defer rmTemp()

	cases := []*Options{
		{
			Label:     "wallet0",
			Seed:      "secure seed",
			Encrypted: true,
			Password:  "password",
		},
		{
			Label:     "wallet1",
			Seed:      "secure seed",
			Encrypted: false,
			Password:  "",
		},
	}

	for _, c := range cases {
		saveWallet(t, c)
		fw := loadWallet(t, c.Label, c.Password)
		m := fw.Meta
		require.Equal(t, m.Password, c.Password)
		require.Equal(t, m.Encrypted, c.Encrypted)
		require.Equal(t, m.Label, c.Label)
		require.Equal(t, m.Seed, c.Seed)
	}

}
