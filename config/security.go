package config

type SecurityConfig struct {
	// 加密配置
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	KDFIterations       int    `json:"kdf_iterations"`

	// 会话安全
	SessionTimeout   int  `json:"session_timeout"` // 分钟
	MaxLoginAttempts int  `json:"max_login_attempts"`
	Enable2FA        bool `json:"enable_2fa"`

	// 审计配置
	EnableAuditLog   bool `json:"enable_audit_log"`
	LogRetentionDays int  `json:"log_retention_days"`

	// 网络安全
	EnableTLS   bool   `json:"enable_tls"`
	TLSCertPath string `json:"tls_cert_path"`
	TLSKeyPath  string `json:"tls_key_path"`
}

func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EncryptionAlgorithm: "AES-256-GCM",
		KDFIterations:       32768,
		SessionTimeout:      30,
		MaxLoginAttempts:    5,
		Enable2FA:           true,
		EnableAuditLog:      true,
		LogRetentionDays:    365,
		EnableTLS:           true,
	}
}
