package app

import "errors"

// 错误定义
var (
	ErrExitRequested = errors.New("exit requested by user")
	ErrInvalidInput  = errors.New("invalid input")
)
