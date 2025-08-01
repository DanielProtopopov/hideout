package structs

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"hideout/internal/folders"
	"hideout/internal/secrets"
)

var (
	Folders []folders.Folder
	Secrets []secrets.Secret // Secret folder map
	Redis   *redis.Client
	Gorm    *gorm.DB
)
