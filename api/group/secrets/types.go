package secrets

import (
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"hideout/internal/common/rqrs"
)

type (
	Secret struct {
		UID   string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Path  string `json:"Path" description:"Folder path" example:"/"`
		Name  string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value string `json:"Value" description:"Secret value" example:"Test"`
		Type  string `json:"Type" description:"Secret value type" example:"int"`
	}

	GetSecretsRQ struct {
		Path       string                  `json:"Path" description:"Folder path" example:"/"`
		Pagination pagination.PaginationRQ `json:"Pagination"`
		Ordering   []ordering.OrderRQ      `json:"Ordering"`
	}

	GetSecretsRS struct {
		Data []Secret `json:"Data"`
		rqrs.ResponseListRS
	}

	GetSecretRS struct {
		Data Secret `json:"Data"`
		rqrs.ResponseRS
	}

	ListSecretParams struct {
		Pagination pagination.PaginationRQ
		Ordering   map[string]bool
	}
)
