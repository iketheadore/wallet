package scanner

import (
	"errors"
	"fmt"
)

type CoinType string

const (
	CoinBitcoin = CoinType("BTC")
	CoinSkycoin = CoinType("SKY")
)

// DepositStatus represents the status of the deposit, in the perspective of an external services.
type DepositStatus string

const (
	// DepositNotProcessed represents the status in which the deposit is not yet processed by the external service.
	DepositNotProcessed = DepositStatus("deposit_status:not_processed")

	// DepositRejected represents the status in which the deposit is rejected by the external service.
	DepositRejected = DepositStatus("deposit_status:rejected")

	// DepositAccepted represents the status in which the deposit is accepted by the external service.
	DepositAccepted = DepositStatus("deposit_status:accepted")
)

var (
	// ErrDepositUnexpected occurs when the deposit is unexpected by the external service.
	ErrDepositUnexpected = errors.New("deposit is unexpected")

	// ErrDepositIncorrectValue occurs when the deposit has an incorrect value allocation.
	ErrDepositIncorrectValue = errors.New("deposit has incorrect value")

	// ErrDepositContractExpired occurs when the deposit's associated contract has expired.
	ErrDepositContractExpired = errors.New("deposit contract has expired")
)

// DepositStatusUpdate is to be sent from external service -> scanner.
type DepositStatusUpdate struct {
	Status DepositStatus
	Err    error
}

// Deposit represents a transaction event.
type Deposit struct {
	CoinType CoinType // BTC or SKY
	Address  string   // Deposit address.
	Value    int64    // Deposit amount (BTC: in satoshis).
	Height   int64    // Block height in which the deposit resides.
	Tx       string   // Transaction ID.
	N        uint32   //
	Status   DepositStatus
}

func (d *Deposit) GetAddressID() string {
	return fmt.Sprintf("%s:%s", d.CoinType, d.Address)
}

type DepositNote struct {
	Deposit
	UpdateC chan DepositStatusUpdate
}
