package iko

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTotalPageCount(t *testing.T) {
	require.Equal(t, totalPageCount(1, 2), uint64(1),
		"One item, two items per page, equals one page")
	require.Equal(t, totalPageCount(0, 2), uint64(0),
		"Zero items, two items per page, equals zero pages")
	require.Equal(t, totalPageCount(3, 2), uint64(2),
		"Three items, two items per page, equals two pages")
	require.Equal(t, totalPageCount(4, 2), uint64(2),
		"Four items, two items per page, equals two pages")
}
