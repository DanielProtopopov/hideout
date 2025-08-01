package ordering

type (
	Order struct {
		Order   bool   `json:"Order" example:"true"`
		OrderBy string `json:"OrderBy" example:"ID"`
	}
)
