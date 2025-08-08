package extra

var (
	EncodingTypeMap = map[uint]string{
		Encoding_Binary:  "binary",
		Encoding_GOB:     "gob",
		Encoding_CSV:     "csv",
		Encoding_JSON:    "json",
		Encoding_XML:     "xml",
		Encoding_Archive: "archive",
	}

	EncodingTypeMapInv = map[string]uint{
		"binary":  Encoding_Binary,
		"gob":     Encoding_GOB,
		"csv":     Encoding_CSV,
		"json":    Encoding_JSON,
		"xml":     Encoding_XML,
		"archive": Encoding_Archive,
	}
)
