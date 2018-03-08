package exchange

import (
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

type Contract struct {
	KittyID         uint64
	KittyAddress    string
	DepositCoinType string
	DepositAddress  string
	DepositValue    int64
	Created         int64
	Expiry          int64
}

func (c *Contract) GetAddressID() string {
	return fmt.Sprintf("%s:%s", c.DepositCoinType, c.DepositAddress)
}

func (c *Contract) Hash() cipher.SHA256 {
	return cipher.SumSHA256(encoder.Serialize(*c))
}

type Exchanger interface {
	AddContract(c Contract) (cipher.SHA256, error)
}
