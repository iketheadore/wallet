package wallet

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testRootDir string

func initTempDir(t *testing.T) func() {
	dir, err := ioutil.TempDir("", "kittycash_test")
	require.Empty(t, err)

	testRootDir = dir

	return func() {
		os.RemoveAll(dir)
	}
}

func saveWallet(options *Options) error {
	fWallet, err := NewWallet(options)
	if err != nil {
		return err
	}
	return fWallet.Save(testRootDir)
}

func loadWallet(label, pw string) (*Wallet, error) {
	f, err := os.Open(LabelPath(testRootDir, label))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return LoadWallet(raw, label, pw)
}

func TestFloatingWallet_Save(t *testing.T) {
	rmTemp := initTempDir(t)
	defer rmTemp()

	cases0 := []*Options{
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
	t.Run("correct_credentials", func(t *testing.T) {
		for _, c := range cases0 {
			require.NoError(t, saveWallet(c))

			fw, err := loadWallet(c.Label, c.Password)
			require.NoError(t, err)

			m := fw.Meta
			require.Equal(t, m.Password, c.Password)
			require.Equal(t, m.Encrypted, c.Encrypted)
			require.Equal(t, m.Label, c.Label)
			require.Equal(t, m.Seed, c.Seed)
		}
	})

	cases1 := []struct {
		Correct    *Options
		FalsePass  string
		ShouldPass bool
	}{
		{cases0[0], "wrong", false},
		{cases0[1], "wrong", true},
	}
	t.Run("wrong_credentials", func(t *testing.T) {
		for _, c := range cases1 {
			require.NoError(t, saveWallet(c.Correct))

			if _, err := loadWallet(c.Correct.Label, c.FalsePass); c.ShouldPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		}
	})
}
