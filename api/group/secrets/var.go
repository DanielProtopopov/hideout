package secrets

var (
	CompressionTypesMap = map[uint][]string{CompressionType_Brotli: {"brotli"}, CompressionType_Bzip2: {"bzip2"}, CompressionType_Flate: {"zip"},
		CompressionType_Gzip: {"gzip"}, CompressionType_Lz4: {"lz4"}, CompressionType_Lzip: {"lz"}, CompressionType_Minlz: {"mz"},
		CompressionType_Snappy: {"sz", "s2"}, CompressionType_XZ: {"xz"}, CompressionType_Zlib: {"zz"}, CompressionType_Zstandard: {"zst"}}

	CompressionTypesMapInv = map[string]uint{"brotli": CompressionType_Brotli, "bzip2": CompressionType_Bzip2, "zip": CompressionType_Flate,
		"gzip": CompressionType_Gzip, "lz4": CompressionType_Lz4, "lz": CompressionType_Lzip, "mz": CompressionType_Minlz,
		"sz": CompressionType_Snappy, "s2": CompressionType_Snappy, "xz": CompressionType_XZ, "zz": CompressionType_Zlib, "zst": CompressionType_Zstandard}

	ArchiveTypesMap = map[uint]string{ArchiveType_None: "", ArchiveType_Tar: "tar", ArchiveType_Zip: "zip"}

	ArchiveTypesMapInv = map[string]uint{"": ArchiveType_None, "tar": ArchiveType_Tar, "zip": ArchiveType_Zip}

	ExportFormatsMap = map[uint]string{ExportFormat_DotEnv: "dotenv"}

	ExportFormatsMapInv = map[string]uint{"dotenv": ExportFormat_DotEnv}
)
