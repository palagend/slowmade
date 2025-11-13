package config

type AppConfig struct {
	Template TemplateConfig `mapstructure:"template"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type TemplateConfig struct {
	CustomTemplateDir string `mapstructure:"custom_template_dir" yaml:"custom_template_dir"`
	EnableCustom      bool   `mapstructure:"enable_custom_templates" yaml:"enable_custom_templates"`
}

type StorageConfig struct {
	KeystoreDir string           `mapstructure:"keystore_dir"`
	Encryption  EncryptionConfig `mapstructure:"encryption"`
}

type EncryptionConfig struct {
	Algorithm string `mapstructure:"algorithm"`
	Cost      int    `mapstructure:"cost"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"levl"`
	Format string `mapstructure:"format"`
}
