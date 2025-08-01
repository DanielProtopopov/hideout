package secrets

import (
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"hideout/internal/common/rqrs"
)

type (
	Secret struct {
		UID     string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		PathUID string `json:"PathUID" description:"Path unique identifier" example:"/"`
		Name    string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value   string `json:"Value" description:"Secret value" example:"Test"`
		Type    string `json:"Type" description:"Secret value type" example:"int"`
	}

	Path struct {
		UID       string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		ParentUID string `json:"ParentUID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name      string `json:"Name" description:"Path name" example:"Folder #1"`
	}

	CreateSecret struct {
		PathUID string `json:"PathUID" description:"Path unique identifier" example:"/"`
		Name    string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value   string `json:"Value" description:"Secret value" example:"Test"`
		Type    string `json:"Type" description:"Secret value type" example:"int"`
	}

	GetSecretsRQ struct {
		PathUID    string                  `json:"PathUID" description:"Folder unique identifier" example:"abc-def-ghi"`
		Pagination pagination.PaginationRQ `json:"Pagination" description:"Pagination"`
		Order      []ordering.Order        `json:"Order" description:"Order"`
	}

	GetSecretsRS struct {
		Secrets []Secret `json:"Secrets"`
		Paths   []Path   `json:"Paths"`
		rqrs.ResponseListRS
	}

	CreateSecretsRQ struct {
		Data []CreateSecret `json:"Data"`
	}

	CreateSecretsRS struct {
		Data []Secret `json:"Data"`
		rqrs.ResponseListRS
	}

	UpdateSecretsRQ struct {
		Data []Secret `json:"Data"`
	}

	UpdateSecretsRS struct {
		Data []Secret `json:"Data"`
		rqrs.ResponseListRS
	}

	DeleteSecretsRQ struct {
		UIDs []string `json:"UIDs"`
	}

	DeleteSecretsRS struct {
		rqrs.ResponseListRS
	}

	ListSecretParams struct {
		Pagination pagination.PaginationRQ
		Order      []ordering.Order
	}
)
