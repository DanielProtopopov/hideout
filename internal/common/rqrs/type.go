package rqrs

import "hideout/internal/common/pagination"

type (
	Error struct {
		Message     string `json:"Message" example:"Message"`
		Description string `json:"Description" example:"Description"`
		Code        uint   `json:"Code" example:"511"`
	}

	ResponseRS struct {
		Errors []Error `json:"Errors"`
	}

	ResponseListRS struct {
		Errors []Error `json:"Errors"`
		pagination.PaginationRS
	}
)
