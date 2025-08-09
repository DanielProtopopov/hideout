package pagination

type (
	Pagination struct {
		PerPage uint `json:"PerPage"`
		Page    uint `json:"Page"`
	}

	PaginationRS struct {
		PerPage uint `json:"PerPage" example:"20"`
		Pages   uint `json:"Pages" example:"14"`
		Total   uint `json:"Total" example:"280"`
	}
)
