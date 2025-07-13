package apiconfig

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"hideout/config"
	"hideout/internal/translations"
	"log"
)

type Config struct {
	Server      config.ServerConfig      // Конфигурация API сервера (хост и порт)
	Environment config.EnvironmentConfig // Конфигурация для окружения
	Bundle      *i18n.Bundle             // Конфигурация для международных переводов (localization)
	I18n        *i18n.Localizer          // Конфигурация для международных переводов (i18n)
	Debug       bool                     // Настройка отладки
	SwaggerHost string                   // Хост для Swagger документации
}

var Settings *Config

func Init(ctx context.Context) {
	bundle := i18n.NewBundle(language.Russian)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, lang := range translations.Languages {
		_, errLoadLang := bundle.LoadMessageFile(fmt.Sprintf("%s/translate.%s.toml", "data/i18n", lang))
		if errLoadLang != nil {
			log.Panicf("Ошибка загрузки переводов для языка %s: %s", lang, errLoadLang.Error())
			return
		}
	}

	if !config.GetEnvAsBool("DEBUG", false) {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	}

	Settings = &Config{
		Bundle: bundle,
		I18n:   i18n.NewLocalizer(bundle, translations.Languages...),
		Debug:  config.GetEnvAsBool("DEBUG", true),
		Server: config.ServerConfig{
			Host:   config.GetEnv("SERVER_HOST", "0.0.0.0"),
			Port:   config.GetEnvAsInt("SERVER_PORT", 81),
			Domain: config.GetEnv("SERVER_DOMAIN", "https://webapi.dev.hideout.com"),
		},
		Environment: config.EnvironmentConfig{
			FullName:  config.GetEnv("ENVIRONMENT_FULL", "development"),
			ShortName: config.GetEnv("ENVIRONMENT_SHORT", "dev"),
		},

		SwaggerHost: config.GetEnv("SWAGGER_HOST", "localhost"),
	}
}
