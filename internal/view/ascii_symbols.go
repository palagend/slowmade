// internal/mvc/views/ascii_symbols.go
package view

// ASCII符号常量定义
/*const (
	WalletIcon  = "💰"   // 如终端不支持可备选为 "[W]"
	CheckIcon   = "✅"   // 备选 "[√]"
	ErrorIcon   = "❌"   // 备选 "[X]"
	InfoIcon    = "ℹ️ " // 备选 "[i]"
	AddressIcon = "🌐"   // 备选 "[A]"
	QRCodeIcon  = "📱"   // 备选 "[QR]"
	ListIcon    = "📋"   // 备选 "[Li]"
	LockIcon    = "🔒"   // 备选 "[Lo]"
)*/

// ANSI颜色代码
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorGray    = "\033[90m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
)

// 样式常量
const (
	StyleBold = "\033[1m"
)

// 带颜色的ASCII图标函数
func SuccessIcon() string {
	return ColorGreen + "[√]" + ColorReset
}

func ErrorIcon() string {
	return ColorRed + "[X]" + ColorReset
}

func WarningIcon() string {
	return ColorYellow + "[!]" + ColorReset
}

func InfoIcon() string {
	return ColorCyan + "[i]" + ColorReset
}

func WalletIcon() string {
	return ColorBlue + "[W]" + ColorReset
}

func AddressIcon() string {
	return ColorMagenta + "[A]" + ColorReset
}
