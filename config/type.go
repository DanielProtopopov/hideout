package config

type (
	ServerConfig struct {
		Host   string
		Port   int
		Domain string
	}

	RepositoryConfig struct {
		Type            uint
		PreloadInMemory bool
		FileName        string
		FileEncoding    uint
	}

	EnvironmentConfig struct {
		FullName  string
		ShortName string
	}

	RedisConfig struct {
		Host     string
		Port     int
		Password string
		DB       int
		Proto    string
	}

	DatabaseConfig struct {
		Type    string
		Host    string
		Port    int
		User    string
		Pass    string
		Name    string
		Proto   string
		SSLMode bool
	}
)
