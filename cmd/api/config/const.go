package apiconfig

const (
	CasBinAdapterType_None = 0 // Do not create CasBin adapter
	CasBinAdapterType_File = 1 // File CasBin adapter
	CasBinModelPath_Sqlx   = 2 // Sqlx CasBin adapter
	CasBinModelPath_Gorm   = 3 // GORM CasBin adapter

)
