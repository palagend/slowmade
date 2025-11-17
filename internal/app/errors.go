// internal/app/errors.go
package app

import "errors"

// 错误定义
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrExitRequested     = errors.New("exit requested")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInvalidAddress    = errors.New("invalid address format")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidPrivateKey = errors.New("invalid private key")
)
