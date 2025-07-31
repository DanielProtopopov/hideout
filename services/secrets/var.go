package secrets

var (
	TypeMap = map[string]uint{
		"memory":   RepositoryType_InMemory,
		"redis":    RepositoryType_Redis,
		"database": RepositoryType_Database,
	}

	TypeMapInv = map[uint]string{
		RepositoryType_InMemory: "memory",
		RepositoryType_Redis:    "redis",
		RepositoryType_Database: "database",
	}
)
