package i18n

import (
	"fmt"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	bundle      *i18n.Bundle
	localizer   *i18n.Localizer
	currentLang string
	mu          sync.RWMutex
)

func Init(configPath string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// 加载语言文件
	languages := []string{"en", "zh", "ja"}
	for _, lang := range languages {
		_, err := bundle.LoadMessageFile(fmt.Sprintf("pkg/i18n/locales/active.%s.yaml", lang))
		if err != nil {
			return fmt.Errorf("failed to load language file for %s: %v", lang, err)
		}
	}

	// 设置默认语言
	SetLanguage("en")
	return nil
}

func SetLanguage(lang string) {
	mu.Lock()
	defer mu.Unlock()

	currentLang = lang
	localizer = i18n.NewLocalizer(bundle, lang)
}

func Tr(messageID string, args ...interface{}) string {
	mu.RLock()
	defer mu.RUnlock()

	if localizer == nil {
		return messageID
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})

	if err != nil {
		return messageID
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
