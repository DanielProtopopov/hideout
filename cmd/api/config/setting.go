package apiconfig

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/redis/go-redis/v9"
	"golang.org/x/text/language"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"hideout/config"
	"hideout/internal/paths"
	"hideout/internal/secrets"
	"hideout/internal/translations"
	"hideout/structs"
	"log"
	"time"
)

type Config struct {
	Server      config.ServerConfig      // API server configuration
	Environment config.EnvironmentConfig // Server environment configuration
	Bundle      *i18n.Bundle             // I18n bundle instance (localization)
	I18n        *i18n.Localizer          // I18n configuration (i18n)
	Redis       config.RedisConfig       // Redis configuration
	Database    config.DatabaseConfig    // Database configuration
	Debug       bool                     // Debugging flag
}

var Settings *Config

func Init(ctx context.Context) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, lang := range translations.Languages {
		_, errLoadLang := bundle.LoadMessageFile(fmt.Sprintf("%s/translate.%s.toml", "data/i18n", lang))
		if errLoadLang != nil {
			log.Panicf("Error loading translations for language %s: %s", lang, errLoadLang.Error())
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
			Port:   config.GetEnvAsInt("SERVER_PORT", 80),
			Domain: config.GetEnv("SERVER_DOMAIN", "localhost"),
		},
		Environment: config.EnvironmentConfig{
			FullName:  config.GetEnv("ENVIRONMENT_FULL", "development"),
			ShortName: config.GetEnv("ENVIRONMENT_SHORT", "dev"),
		},
		Redis: config.RedisConfig{
			Host:     config.GetEnv("REDIS_HOST", "localhost"),
			Port:     config.GetEnvAsInt("REDIS_PORT", 6379),
			Password: config.GetEnv("REDIS_PASSWORD", ""),
			DB:       config.GetEnvAsInt("REDIS_DB", 0),
			Proto:    config.GetEnv("REDIS_PROTOCOL", "tcp"),
		},
		Database: config.DatabaseConfig{
			Type:    config.GetEnv("DB_TYPE", "postgresql"),
			Host:    config.GetEnv("DB_HOST", "localhost"),
			Port:    config.GetEnvAsInt("DB_PORT", 5432),
			User:    config.GetEnv("DB_USER", "postgres"),
			Pass:    config.GetEnv("DB_PASSWORD", "postgres"),
			Name:    config.GetEnv("DB_NAME", "hideout"),
			Proto:   config.GetEnv("DB_PROTOCOL", "postgresql"),
			SSLMode: config.GetEnv("DB_SSL_MODE", "disable"),
		},
	}

	client := redis.NewClient(&redis.Options{
		Network: Settings.Redis.Proto, Addr: fmt.Sprintf("%s:%d", Settings.Redis.Host, Settings.Redis.Port),
		Password: Settings.Redis.Password, DB: Settings.Redis.DB, ConnMaxIdleTime: 5 * time.Minute, MaxRetries: 3,
	})
	if errPing := client.Ping(ctx).Err(); errPing != nil {
		log.Panicf("Error pinging Redis: %s", errPing)
	} else {
		log.Println("Redis was pinged successfully")
	}

	conn, errConnectSQL := sqlx.Connect(Settings.Database.Type, Settings.Database.GetDSN())
	if errConnectSQL != nil {
		log.Panicf("Error connecting database %s on host %s: %s",
			Settings.Database.Name, Settings.Database.Host, errConnectSQL.Error())
	} else {
		log.Printf("Successfully connected to database %s on host %s", Settings.Database.Name, Settings.Database.Host)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	grm, errGormOpen := gorm.Open(postgres.New(postgres.Config{DSN: Settings.Database.GetDSN()}))
	if errGormOpen != nil {
		log.Panicf("Error connecting to the database: %s", errGormOpen.Error())
	} else {
		// grm.SkipDefaultTransaction = !Settings.Debug
		structs.Gorm = grm
		log.Printf("GORM connected to database %s on host %s", Settings.Database.Name, Settings.Database.Host)
	}

	if Settings.Debug {
		structs.Gorm = structs.Gorm.Debug()
	}

	structs.Redis = client
	structs.Secrets = []secrets.Secret{}
	structs.Paths = []paths.Path{}
}
