package http

import (
	_ "errors"
	"fmt"
	_ "github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/wallet"
	"github.com/skycoin/skycoin/src/cipher"
	_ "github.com/stretchr/testify/require"
	"net/http"
	_ "sync"
	"testing"
)

func runWalletGatewayTest(t *testing.T, wallet *wallet.Manager) {

	// Get an http server
	mux := http.NewServeMux()

	fmt.Println(mux)
	gateway := walletGateway(mux, wallet)
	fmt.Println(gateway)
}

func TestWalletGateway(t *testing.T) {
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

	entry1, err := wallet.NewEntry(sk)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(entry1)
	entry2, err := wallet.NewEntry(sk2)
	if err != nil {
		fmt.Println(err, "2")
	}
	fmt.Println(entry2)
	/*
		wallet.Manager{
			mux: sync.Mutex,
			label: ["Just", "A", "String"],
			wallets map
		}
		manager, err := wallet.NewManager()

		if err != nil {
			errors.New("Cannot create manager")
			fmt.Println(manager)
		} else {
			fmt.Println("It works")
			runWalletGatewayTest(t, manager)
		}
	*/
}
