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
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, text, console
	Output     string `mapstructure:"output"` // stdout, stderr, file path
	TimeFormat string `mapstructure:"time_format"`

	EnableCaller bool `mapstructure:"enable_caller"` // 是否显示调用者信息
	EnableStack  bool `mapstructure:"enable_stack"`  // 是否启用堆栈跟踪
	Development  bool `mapstructure:"development"`   // 开发模式（更详细日志）
	Color        bool `mapstructure:"color"`         // 是否启用颜色（仅text/console格式）
}
