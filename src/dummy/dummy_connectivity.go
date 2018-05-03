package dummy

import "github.com/kittycash/wallet/src/connectivity"

func (_ *Dummy) Status() (connectivity.Status, error) {
	return connectivity.Connected, nil
}

func (_ *Dummy) Reconnect() bool {
	return true
}
