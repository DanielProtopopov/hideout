package structs

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"hideout/internal/paths"
	"hideout/internal/secrets"
)

var (
	Paths   []paths.Path
	Secrets []secrets.Secret // Secret path map
	Redis   *redis.Client
	Gorm    *gorm.DB
)
