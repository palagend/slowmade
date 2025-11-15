// wallet.go
package core

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palagend/slowmade/pkg/i18n"
	"github.com/palagend/slowmade/pkg/logging"
	"go.uber.org/zap"
)

// SendTransaction 发送加密货币交易
func (wm *WalletManager) SendTransaction(toAddress string, amount string) (string, error) {
	if wm.current == nil || wm.current.status != StatusUnlocked {
		return "", fmt.Errorf(i18n.Tr("ERR_WALLET_LOCKED"))
	}

	// 1. 验证接收地址格式
	if !common.IsHexAddress(toAddress) {
		return "", fmt.Errorf(i18n.Tr("ERR_INVALID_ADDRESS"))
	}
	to := common.HexToAddress(toAddress)

	// 2. 解析转账金额
	value, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf(i18n.Tr("ERR_INVALID_AMOUNT"))
	}

	// 3. 获取当前网络参数（简化实现）
	chainID := big.NewInt(1) // 主网
	gasLimit := uint64(21000)
	gasPrice, err := wm.getSuggestedGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf(i18n.Tr("ERR_GET_GAS_PRICE"), err)
	}

	// 4. 构建交易
	nonce := uint64(0) // 实际应从节点获取
	tx := types.NewTransaction(
		nonce,
		to,
		value,
		gasLimit,
		gasPrice,
		nil,
	)

	// 5. 签名交易
	privateKey, err := wm.current.getDecryptedPrivateKey()
	if err != nil {
		return "", fmt.Errorf(i18n.Tr("ERR_DECRYPT_KEY"), err)
	}
	defer wm.secureClearPrivateKey(privateKey)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf(i18n.Tr("ERR_SIGN_TX"), err)
	}

	// 6. 广播交易（简化实现）
	txHash := signedTx.Hash().Hex()
	logging.Get().Info("Transaction signed",
		zap.String("to", to.Hex()),
		zap.String("value", value.String()),
		zap.String("txHash", txHash))

	// 实际应通过RPC发送到区块链节点
	if err := wm.broadcastTransaction(context.Background(), signedTx); err != nil {
		return "", fmt.Errorf(i18n.Tr("ERR_BROADCAST_TX"), err)
	}

	return txHash, nil
}

// getDecryptedPrivateKey 解密当前钱包的私钥
func (w *Wallet) getDecryptedPrivateKey() (*ecdsa.PrivateKey, error) {
	if w.status != StatusUnlocked {
		return nil, fmt.Errorf(i18n.Tr("ERR_WALLET_LOCKED"))
	}

	// 简化实现：生成一个测试私钥
	// 实际实现应从keystore解密私钥
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate test private key: %w", err)
	}

	return privateKey, nil
}

// secureClearPrivateKey 安全清理私钥内存
func (wm *WalletManager) secureClearPrivateKey(privateKey *ecdsa.PrivateKey) {
	if privateKey != nil {
		// 将私钥的D值（大整数）清零
		if privateKey.D != nil {
			privateKey.D.SetInt64(0)
		}
	}
}

// getSuggestedGasPrice 获取建议的Gas价格（简化实现）
func (wm *WalletManager) getSuggestedGasPrice(ctx context.Context) (*big.Int, error) {
	// 实际实现应通过RPC查询节点
	return big.NewInt(20000000000), nil // 20 Gwei
}

// broadcastTransaction 广播交易（简化实现）
func (wm *WalletManager) broadcastTransaction(ctx context.Context, tx *types.Transaction) error {
	// 实际实现应通过RPC发送到区块链节点
	logging.Get().Info("Transaction broadcast simulation",
		zap.String("hash", tx.Hash().Hex()),
		zap.Uint64("nonce", tx.Nonce()),
		zap.String("to", tx.To().Hex()),
		zap.String("value", tx.Value().String()))

	// 模拟广播成功
	logging.Get().Info("Transaction broadcast completed")
	return nil
}

// GetBalance 获取钱包余额（简化实现）
func (w *Wallet) GetBalance() (*Balance, error) {
	// 简化实现，实际应从区块链节点获取
	return &Balance{
		Currency:  "ETH",
		Amount:    "0.0",
		FiatValue: "0.0",
	}, nil
}

// Balance 余额结构体
type Balance struct {
	Currency  string
	Amount    string
	FiatValue string
}
