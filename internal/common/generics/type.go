package generics

import (
	"hideout/internal/common/pagination"
	"time"
)

type (
	FromTo[T any] struct {
		From T `json:"From"`
		To   T `json:"To"`
	}

	ListParams struct {
		*pagination.Pagination
		IDs []uint
		// Deleted values:
		// * 1 - model.No
		// * 2 - model.No
		// * 3 - model.YesOrNo
		Deleted   uint
		Order     map[string]bool
		CreatedAt FromTo[time.Time]
		UpdatedAt FromTo[time.Time]
		DeletedAt FromTo[time.Time]
	}
)
