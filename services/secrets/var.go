package secrets

var (
	TypeMap = map[string]uint{
		"memory":   RepositoryType_InMemory,
		"redis":    RepositoryType_Redis,
		"database": RepositoryType_Database,
		"file":     RepositoryType_File,
		"git":      RepositoryType_Git,
	}

	TypeMapInv = map[uint]string{
		RepositoryType_InMemory: "memory",
		RepositoryType_Redis:    "redis",
		RepositoryType_Database: "database",
		RepositoryType_File:     "file",
		RepositoryType_Git:      "git",
	}
)
