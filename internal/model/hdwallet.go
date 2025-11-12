package model

import (
	"time"
)

type HDWallet struct {
	ID         string     `json:"id"`
	Mnemonic   string     `json:"mnemonic"`
	Seed       []byte     `json:"seed"`
	Accounts   []*Account `json:"accounts"`
	CoinSymbol string     `json:"coin_symbol"` // refer: https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	CreatedAt  time.Time  `json:"created_at"`
	Version    string     `json:"version"`
}

type Account struct {
	Index     uint32     `json:"index"`
	Label     string     `json:"label"`
	Path      string     `json:"path"`
	Addresses []*Address `json:"addresses"`
	Balance   uint64     `json:"balance"`
}

type Address struct {
	Index     uint32 `json:"index"`
	Address   string `json:"address"`
	PublicKey []byte `json:"public_key"`
	Used      bool   `json:"used"`
	Path      string `json:"path"`
}
