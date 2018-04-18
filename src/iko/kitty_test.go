package iko

import (
	"testing"

	"github.com/kittycash/kittiverse/src/kitty"
)

func TestKittyIDs_Sort(t *testing.T) {
	ids := kitty.IDs{
		kitty.ID(65),
		kitty.ID(2),
		kitty.ID(20),
		kitty.ID(23),
		kitty.ID(12),
		kitty.ID(3),
		kitty.ID(94),
		kitty.ID(24),
	}
	ids.Sort()
	t.Log(ids)
}
