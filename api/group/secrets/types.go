package secrets

import (
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"hideout/internal/common/rqrs"
)

type (
	Secret struct {
		ID        uint   `json:"ID" description:"Secret primary unique identifier" example:"1"`
		UID       string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		FolderUID string `json:"FolderUID" description:"Folder unique identifier" example:"/"`
		Name      string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value     string `json:"Value" description:"Secret value" example:"Test"`
		Script    string `json:"Script" description:"Script for dynamically calculated value" example:"time.RFC3339"`
	}

	Folder struct {
		ID        uint   `json:"ID" description:"Folder primary unique identifier" example:"1"`
		UID       string `json:"UID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		ParentUID string `json:"ParentUID" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name      string `json:"Name" description:"Folder name" example:"Folder #1"`
	}

	CreateSecret struct {
		FolderUID string `json:"FolderUID" description:"Folder unique identifier" example:"abc-edf-hij"`
		Name      string `json:"Name" description:"Secret name" example:"DEBUG"`
		Value     string `json:"Value" description:"Secret value" example:"Test"`
		Script    string `json:"Script" description:"Script for dynamically calculated value" example:"time.RFC3339"`
	}

	GetSecretsRQ struct {
		FolderUID         string                `json:"FolderUID" description:"Folder unique identifier" example:"abc-def-ghi"`
		SecretsPagination pagination.Pagination `json:"SecretsPagination" description:"Secrets pagination"`
		SecretsOrder      []ordering.Order      `json:"SecretsOrder" description:"Secrets order"`
		FoldersPagination pagination.Pagination `json:"FoldersPagination" description:"Folders pagination"`
		FoldersOrder      []ordering.Order      `json:"FoldersOrder" description:"Folders order"`
	}

	GetSecretsRS struct {
		Secrets []Secret `json:"Secrets" description:"Secrets list"`
		Folders []Folder `json:"Folders" description:"Folders list"`
		rqrs.ResponseListRS
	}

	CreateSecretsRQ struct {
		Data []CreateSecret `json:"Data" description:"Secrets to create"`
	}

	CreateSecretsRS struct {
		Data []Secret `json:"Data" description:"Created secrets"`
		rqrs.ResponseListRS
	}

	UpdateSecretsRQ struct {
		Data []Secret `json:"Data" description:"Secrets to update"`
	}

	UpdateSecretsRS struct {
		Data []Secret `json:"Data" description:"Updated secrets"`
		rqrs.ResponseListRS
	}

	DeleteSecretsRQ struct {
		SecretUIDs []string `json:"SecretUIDs" description:"Secret unique identifiers list for deletion"`
		FolderUIDs []string `json:"FolderUIDs" description:"Folder unique identifiers list for deletion"`
	}

	DeleteSecretsRS struct {
		rqrs.ResponseListRS
	}

	CopyPasteSecretsRQ struct {
		SecretUIDs    []string `json:"SecretUIDs" description:"Secret UIDs to copy-and-paste"`
		FolderUIDs    []string `json:"FolderUIDs" description:"Folder UIDs to copy-and-paste"`
		FromFolderUID string   `json:"FromFolderUID" description:"Source folder unique identifier"`
		ToFolderUID   string   `json:"ToFolderUID" description:"Target folder unique identifier"`
	}

	CopyPasteSecretsRS struct {
		Secrets []Secret `json:"Secrets" description:"Secrets list"`
		Folders []Folder `json:"Folders" description:"Folders list"`
		rqrs.ResponseListRS
	}

	ExportSecretsRQ struct {
		Format          string                `json:"Format" description:"Export format" enums:"dotenv"`
		CompressionType string                `json:"CompressionType" description:"Compression type" enums:"brotli,bzip2,zip,gzip,lz4,lz,mz,sz,s2,xz,zz,zst"`
		ArchiveType     string                `json:"ArchiveType" description:"Archive type" enums:"tar,zip"`
		FolderUID       string                `json:"FolderUID" description:"Folder unique identifier" example:"abc-def-ghi"`
		Pagination      pagination.Pagination `json:"Pagination" description:"Secrets pagination"`
		Order           []ordering.Order      `json:"SOrder" description:"Secrets order"`
	}

	ExportSecretsRS struct {
		Secrets []Secret `json:"Secrets" description:"Secrets list"`
		rqrs.ResponseListRS
	}

	DiffSecretsRQ struct {
		FolderUID string `json:"FolderUID" description:"Folder unique identifier" example:"abc-def-ghi"`
		Data      string `json:"Data" description:"Dotenv formatted data" example:"VAR_1=value1\nVAR_2=value2\nVAR_3=value3"`
	}

	DiffSecretsRS struct {
		Create []Secret `json:"Create" description:"Secrets to create"`
		Update []Secret `json:"Update" description:"Secrets to update"`
		Delete []Secret `json:"Delete" description:"Secrets to delete"`
		rqrs.ResponseRS
	}

	ListSecretParams struct {
		Pagination pagination.Pagination `json:"Pagination" description:"Secrets pagination parameters"`
		Order      []ordering.Order      `json:"Order" description:"Secrets order parameters"`
	}
)
