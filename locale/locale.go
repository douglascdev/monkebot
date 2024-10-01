package locale

import (
	"embed"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml
var localeFS embed.FS

var (
	SupportedLocales = []string{"en-US", "pt-BR"}
	Localizers       = loadLocalizers(SupportedLocales...)
)

func loadLocalizers(supportedLocales ...string) map[string]*i18n.Localizer {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	result := make(map[string]*i18n.Localizer)

	for _, locale := range supportedLocales {
		bundle.LoadMessageFileFS(localeFS, "active."+locale+".toml")

		localizer := i18n.NewLocalizer(bundle, locale)
		result[locale] = localizer
	}

	return result
}
