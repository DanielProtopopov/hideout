package secrets

const (
	CompressionType_Brotli    = 1  // .br
	CompressionType_Bzip2     = 2  // .bzip2
	CompressionType_Flate     = 3  // .zip
	CompressionType_Gzip      = 4  // .gz
	CompressionType_Lz4       = 5  // .lz4
	CompressionType_Lzip      = 6  // .lz
	CompressionType_Minlz     = 7  // .mz
	CompressionType_Snappy    = 8  // .sz and .s2
	CompressionType_XZ        = 9  // .xz
	CompressionType_Zlib      = 10 // .zz
	CompressionType_Zstandard = 11 // .zst

	ArchiveType_None = 0
	ArchiveType_Zip  = 1
	ArchiveType_Tar  = 2

	ExportFormat_DotEnv = 1
)
