package storage

// Keystore加密结构（类似Geth格式）
type Keystore struct {
	Version int    `json:"version"`
	ID      string `json:"id"`
	Address string `json:"address"`
	Crypto  struct {
		Cipher       string `json:"cipher"`
		CipherText   string `json:"ciphertext"`
		CipherParams struct {
			IV string `json:"iv"`
		} `json:"cipherparams"`
		KDF       string `json:"kdf"` // 密钥派生函数
		KDFParams struct {
			Salt string `json:"salt"`
			N    int    `json:"n"` // CPU/内存成本参数
			R    int    `json:"r"`
			P    int    `json:"p"`
		} `json:"kdfparams"`
		MAC string `json:"mac"` // 完整性校验
	} `json:"crypto"`
}
