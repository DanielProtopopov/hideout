package secrets

import (
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"hideout/internal/common/rqrs"
)

type (
	Secret struct {
		UID       string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		FolderUID string `json:"FolderUID" description:"Folder unique identifier" example:"/"`
		Name      string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value     string `json:"Value" description:"Secret value" example:"Test"`
		Type      string `json:"Type" description:"Secret value type" example:"int"`
		IsDynamic bool   `json:"IsDynamic" description:"Does secret has dynamic value" example:"false"`
	}

	Folder struct {
		UID       string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		ParentUID string `json:"ParentUID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name      string `json:"Name" description:"Folder name" example:"Folder #1"`
	}

	CreateSecret struct {
		FolderUID string `json:"FolderUID" description:"Folder unique identifier" example:"/"`
		Name      string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value     string `json:"Value" description:"Secret value" example:"Test"`
		Type      string `json:"Type" description:"Secret value type" example:"int"`
		IsDynamic bool   `json:"IsDynamic" description:"Does secret has dynamic value" example:"false"`
	}

	GetSecretsRQ struct {
		FolderUID         string                `json:"FolderUID" description:"Folder unique identifier" example:"abc-def-ghi"`
		SecretsPagination pagination.Pagination `json:"SecretsPagination" description:"Secrets pagination"`
		SecretsOrder      []ordering.Order      `json:"SecretsOrder" description:"Secrets order"`
		FoldersPagination pagination.Pagination `json:"FoldersPagination" description:"Folders pagination"`
		FoldersOrder      []ordering.Order      `json:"FoldersOrder" description:"Folders order"`
	}

	GetSecretsRS struct {
		Secrets []Secret `json:"Secrets"`
		Folders []Folder `json:"Folders"`
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
		SecretUIDs []string `json:"SecretUIDs"`
		FolderUIDs []string `json:"FolderUIDs"`
	}

	DeleteSecretsRS struct {
		rqrs.ResponseListRS
	}

	CopyPasteSecretsRQ struct {
		SecretUIDs    []string `json:"SecretUIDs"`
		FolderUIDs    []string `json:"FolderUIDs"`
		FromFolderUID string   `json:"FromFolderUID"`
		ToFolderUID   string   `json:"ToFolderUID"`
	}

	CopyPasteSecretsRS struct {
		Secrets []Secret `json:"Secrets"`
		Folders []Folder `json:"Folders"`
		rqrs.ResponseListRS
	}

	ListSecretParams struct {
		Pagination pagination.Pagination
		Order      []ordering.Order
	}
)
