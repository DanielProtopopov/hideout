package pagination

type (
	Pagination struct {
		PerPage uint `json:"PerPage" validate:"gte=1" binding:"required"`
		Page    uint `json:"Page" validate:"gte=1" binding:"required"`
	}

	PaginationRS struct {
		PerPage uint `json:"PerPage" example:"20"`
		Pages   uint `json:"Pages" example:"14"`
		Total   uint `json:"Total" example:"280"`
	}
)
