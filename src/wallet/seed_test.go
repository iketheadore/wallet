package wallet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeedBitSizeFromString_EmptyString(t *testing.T) {
	val, err := SeedBitSizeFromString("")
	require.Nil(t, err, "Should not return an error")
	require.Equal(t, val, DefaultSeedBitSize,
		"Should return the default seed bit size")
}

func TestSeedBitSizeFromString_Malformed(t *testing.T) {
	_, err := SeedBitSizeFromString("don't think this is an int...")
	require.NotNil(t, err, "Should return an error")
}

func TestSeedBitSizeFromString_Unsupported(t *testing.T) {
	unsupportedSize := 5

	// let's just confirm we're not using a supported bit size
	for _, supportedSize := range ValidSeedBitSizes() {
		if unsupportedSize == supportedSize {
			require.NotEqual(t, supportedSize, unsupportedSize,
				"We should be using an unsupported bit size, I must've messed up...")
		}
	}

	unsupportedSizeString := fmt.Sprintf("%d", unsupportedSize)

	_, err := SeedBitSizeFromString(unsupportedSizeString)
	require.NotNil(t, err, "Should return an error")
}

func TestSeedBitSizeFromString_ValidIntString(t *testing.T) {
	// let's test with a non-default value, which is index 1
	supportedSize := ValidSeedBitSizes()[1]
	supportedSizeStr := fmt.Sprintf("%d", supportedSize)

	value, err := SeedBitSizeFromString(supportedSizeStr)
	require.Nil(t, err, "Shouldn't return an error")
	require.Equal(t, supportedSize, value, "Should return the int we gave it")
}
