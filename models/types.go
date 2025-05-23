package models

import (
	"math/big"
)

// Call represents a single call to be executed
type Call struct {
	To    string   `json:"to"`
	Value *big.Int `json:"value"`
	Data  string   `json:"data"`
}

// Transaction represents a batch of calls to be executed atomically
type Transaction struct {
	Nonce uint64 `json:"nonce"`
	Calls []Call `json:"calls"`
}

// TransactionGroup represents a group of calls that share the same nonce
type TransactionGroup struct {
	Nonce uint64 `json:"nonce"`
	Calls []Call `json:"calls"`
}

// TransactionBatch represents a collection of transaction groups to be merklized
type TransactionBatch struct {
	Groups []TransactionGroup `json:"groups"`
}
