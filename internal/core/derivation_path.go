package core

import (
	"fmt"
	"strconv"
	"strings"
)

type DerivationPath struct {
	Purpose      uint32
	CoinType     uint32
	AccountIndex uint32
	Change       uint32
	AddressIndex uint32
}

func ParseDerivationPath(path string) (*DerivationPath, error) {
	// 移除前缀 "m/" 如果存在
	cleanPath := strings.TrimPrefix(path, "m/")
	if cleanPath == path {
		return nil, fmt.Errorf("invalid BIP44 path format, should start with 'm/'")
	}

	// 分割路径组件
	components := strings.Split(cleanPath, "/")
	if len(components) != 5 {
		return nil, fmt.Errorf("BIP44 path should have exactly 5 components, got %d", len(components))
	}

	result := &DerivationPath{}

	// 解析 purpose (带硬化标记)
	purpose, err := parsePathComponent(components[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse purpose: %w", err)
	}
	result.Purpose = purpose

	// 解析 coin type (带硬化标记)
	coinType, err := parsePathComponent(components[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse coin type: %w", err)
	}
	result.CoinType = coinType

	// 解析 account (带硬化标记)
	account, err := parsePathComponent(components[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse account: %w", err)
	}
	result.AccountIndex = account

	// 解析 change (不带硬化标记)
	change, err := parsePathComponent(components[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse change: %w", err)
	}
	if change != 0 && change != 1 {
		return nil, fmt.Errorf("change should be 0 or 1, got %d", change)
	}
	result.Change = change

	// 解析 address index (不带硬化标记)
	addressIndex, err := parsePathComponent(components[4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse address index: %w", err)
	}
	result.AddressIndex = addressIndex

	return result, nil
}

// parsePathComponent 解析单个路径组件，处理硬化标记
func parsePathComponent(component string) (uint32, error) {
	// 检查是否是硬化标记（以'结尾）
	isHardened := strings.HasSuffix(component, "'")
	if isHardened {
		component = strings.TrimSuffix(component, "'")
	}

	// 转换为数字
	value, err := strconv.ParseUint(component, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid component '%s': %w", component, err)
	}

	// 对于硬化标记，设置最高位（BIP32规范）
	if isHardened {
		value |= 0x80000000
	}

	return uint32(value), nil
}

// FormatDerivationPath 将DerivationPath格式化为字符串
func (p *DerivationPath) String() string {
	return fmt.Sprintf("m/%d'/%d'/%d'/%d/%d",
		p.Purpose&0x7FFFFFFF, // 移除硬化标记位
		p.CoinType&0x7FFFFFFF,
		p.AccountIndex&0x7FFFFFFF,
		p.Change,
		p.AddressIndex)
}

func (p *DerivationPath) PurposeString() string {
	return fmt.Sprintf("%d'", p.Purpose&0x7FFFFFFF)
}

func (p *DerivationPath) CoinTypeString() string {
	return fmt.Sprintf("%d'", p.CoinType&0x7FFFFFFF)
}

func (p *DerivationPath) AccountString() string {
	return fmt.Sprintf("%d'", p.AccountIndex&0x7FFFFFFF)
}

// MaskSuffix 掩盖changeType和addressIndex
func (p *DerivationPath) MaskSuffix() *DerivationPath {
	dp := &DerivationPath{
		Purpose:      p.Purpose,
		CoinType:     p.CoinType,
		AccountIndex: p.AccountIndex,
		Change:       uint32(0),
		AddressIndex: uint32(0),
	}
	return dp
}
