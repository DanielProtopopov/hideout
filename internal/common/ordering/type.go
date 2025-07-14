package ordering

type (
	OrderRQ struct {
		Order   bool   `json:"Order" example:"true"`
		OrderBy string `json:"OrderBy" example:"ID"`
	}
)
