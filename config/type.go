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
