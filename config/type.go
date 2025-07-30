package config

type ServerConfig struct {
	Host   string
	Port   int
	Domain string
}

type EnvironmentConfig struct {
	FullName  string
	ShortName string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Proto    string
}

type DatabaseConfig struct {
	Type    string
	Host    string
	Port    int
	User    string
	Pass    string
	Name    string
	Proto   string
	SSLMode bool
}
